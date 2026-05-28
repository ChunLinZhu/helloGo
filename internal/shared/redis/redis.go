// Package redis — 微服务共享 Redis 客户端封装
// 从 Phase 1 的 internal/pkg/redis 适配，导入 shared/config
// 连接失败时自动降级为 sync.Map 内存缓存
package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"helloGo/internal/shared/config"
)

// Client Redis 客户端封装
// 内部持有 *goredis.Client 或 sync.Map 降级缓存
type Client struct {
	rdb      *goredis.Client
	fallback *sync.Map // 降级内存缓存
	logger   *zap.Logger
	isRedis  bool // 是否使用真实 Redis（false 表示降级模式）
}

// fallbackEntry 内存缓存条目（含过期时间）
type fallbackEntry struct {
	value     string
	expiresAt time.Time
}

// New 创建 Redis 客户端
// 连接失败时降级为内存缓存，不阻塞启动
func New(cfg config.RedisConfig, logger *zap.Logger) *Client {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})

	// 尝试 Ping 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis 连接失败，降级为内存缓存",
			zap.String("addr", addr),
			zap.Error(err),
		)
		return &Client{
			fallback: &sync.Map{},
			logger:   logger,
			isRedis:  false,
		}
	}

	logger.Info("Redis 连接成功", zap.String("addr", addr))
	return &Client{
		rdb:     rdb,
		logger:  logger,
		isRedis: true,
	}
}

// IsRedis 返回当前是否使用真实 Redis
func (c *Client) IsRedis() bool {
	return c.isRedis
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c.isRedis {
		return c.rdb.Get(ctx, key).Result()
	}
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
			return entry.value, nil
		}
		c.fallback.Delete(key)
	}
	return "", goredis.Nil
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if c.isRedis {
		n, err := c.rdb.Exists(ctx, key).Result()
		return n > 0, err
	}
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
			return true, nil
		}
		c.fallback.Delete(key)
	}
	return false, nil
}

// Set 设置键值（带 TTL）
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.isRedis {
		return c.rdb.Set(ctx, key, value, ttl).Err()
	}
	entry := fallbackEntry{value: fmt.Sprintf("%v", value)}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	c.fallback.Store(key, entry)
	return nil
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if c.isRedis {
		return c.rdb.Del(ctx, keys...).Err()
	}
	for _, key := range keys {
		c.fallback.Delete(key)
	}
	return nil
}

// Incr 原子递增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	if c.isRedis {
		return c.rdb.Incr(ctx, key).Result()
	}
	var current int64
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
			fmt.Sscanf(entry.value, "%d", &current)
		}
	}
	current++
	c.fallback.Store(key, fallbackEntry{value: fmt.Sprintf("%d", current)})
	return current, nil
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if c.isRedis {
		return c.rdb.Expire(ctx, key, ttl).Err()
	}
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		entry.expiresAt = time.Now().Add(ttl)
		c.fallback.Store(key, entry)
	}
	return nil
}

// TTL 获取键的剩余生存时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.isRedis {
		return c.rdb.TTL(ctx, key).Result()
	}
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() {
			return -1, nil
		}
		remaining := time.Until(entry.expiresAt)
		if remaining < 0 {
			return -2, nil
		}
		return remaining, nil
	}
	return -2, nil
}

// Ping 检查 Redis 连接
func (c *Client) Ping(ctx context.Context) error {
	if c.isRedis {
		return c.rdb.Ping(ctx).Err()
	}
	return nil
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	if c.isRedis && c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}
