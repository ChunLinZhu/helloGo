// Package redis — Redis 客户端封装
// 初始化 go-redis 客户端，封装 Get/Set/Del/Ping；
// 连接失败时自动降级为 sync.Map 内存缓存（适合开发环境无 Redis 的场景）
package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"helloGo/internal/config"

	goredis "github.com/redis/go-redis/v9"
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

// Init 初始化 Redis 客户端
// 连接失败时降级为内存缓存，不阻塞启动
func Init(cfg config.RedisConfig, logger *zap.Logger) *Client {
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

// IsRedis 返回当前是否使用真实 Redis（false = 降级内存缓存）
func (c *Client) IsRedis() bool {
	return c.isRedis
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c.isRedis {
		return c.rdb.Get(ctx, key).Result()
	}
	// 降级模式
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
			return entry.value, nil
		}
		c.fallback.Delete(key) // 已过期
	}
	return "", goredis.Nil
}

// Set 设置键值（带 TTL）
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.isRedis {
		return c.rdb.Set(ctx, key, value, ttl).Err()
	}
	// 降级模式
	entry := fallbackEntry{
		value: fmt.Sprintf("%v", value),
	}
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
	// 降级模式
	for _, key := range keys {
		c.fallback.Delete(key)
	}
	return nil
}

// Incr 原子递增（用于登录失败计数等）
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	if c.isRedis {
		return c.rdb.Incr(ctx, key).Result()
	}
	// 降级模式：简单实现（非严格原子，开发环境足够用）
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
	// 降级模式：更新过期时间
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
	// 降级模式
	if v, ok := c.fallback.Load(key); ok {
		entry := v.(fallbackEntry)
		if entry.expiresAt.IsZero() {
			return -1, nil // 无过期
		}
		remaining := time.Until(entry.expiresAt)
		if remaining < 0 {
			return -2, nil // 已过期
		}
		return remaining, nil
	}
	return -2, nil // 键不存在
}

// Ping 检查 Redis 连接
func (c *Client) Ping(ctx context.Context) error {
	if c.isRedis {
		return c.rdb.Ping(ctx).Err()
	}
	return nil // 内存缓存始终可用
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	if c.isRedis && c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}
