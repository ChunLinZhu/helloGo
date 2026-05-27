// Package config — 使用 Viper 加载应用配置
// 优先级：环境变量 > .env.{APP_ENV} > .env > 默认值
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全部配置项
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Security SecurityConfig
	Throttle ThrottleConfig
	Metrics  MetricsConfig
	Swagger  SwaggerConfig
	Upload   UploadConfig
	Login    LoginConfig
}

// AppConfig 通用配置
type AppConfig struct {
	Env  string `mapstructure:"APP_ENV"`  // 运行环境: development / production / test
	Port int    `mapstructure:"PORT"`     // HTTP 监听端口
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `mapstructure:"DB_TYPE"`      // 数据库类型: sqlite / mysql / postgres
	SQLite   SQLiteConfig
	MySQL    MySQLConfig
	Postgres PostgresConfig
}

// SQLiteConfig SQLite 专用配置
type SQLiteConfig struct {
	Path string `mapstructure:"SQLITE_PATH"` // 数据库文件路径
}

// MySQLConfig MySQL 专用配置
type MySQLConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASS"`
	Name     string `mapstructure:"DB_NAME"`
}

// PostgresConfig PostgreSQL 专用配置
type PostgresConfig struct {
	Host     string `mapstructure:"PG_HOST"`
	Port     int    `mapstructure:"PG_PORT"`
	User     string `mapstructure:"PG_USER"`
	Password string `mapstructure:"PG_PASS"`
	Name     string `mapstructure:"PG_DB"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASS"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret         string `mapstructure:"JWT_SECRET"`          // 签名密钥
	Expires        string `mapstructure:"JWT_EXPIRES"`         // access token 过期时间（如 "1d", "2h"）
	RefreshExpires string `mapstructure:"JWT_REFRESH_EXPIRES"` // refresh token 过期时间（如 "7d"）
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CSRFEnabled bool   `mapstructure:"CSRF_ENABLED"` // 是否启用 CSRF 防护
	CSRFMode    string `mapstructure:"CSRF_MODE"`    // CSRF 模式: header / cookie
	CSRFSecret  string `mapstructure:"CSRF_SECRET"`  // CSRF token 签名密钥
	CORSOrigins string `mapstructure:"CORS_ORIGINS"` // 允许的跨域来源（逗号分隔）
}

// ThrottleConfig 限流配置
type ThrottleConfig struct {
	TTL   int `mapstructure:"THROTTLE_TTL"`   // 限流时间窗口（秒）
	Limit int `mapstructure:"THROTTLE_LIMIT"` // 窗口内最大请求数
}

// MetricsConfig Prometheus 指标配置
type MetricsConfig struct {
	Enabled bool `mapstructure:"ENABLE_METRICS"` // 是否启用指标采集
}

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Enabled bool `mapstructure:"SWAGGER_ENABLE"` // 是否启用 Swagger 文档
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	Dest            string `mapstructure:"UPLOAD_DEST"`              // 上传文件存储目录
	MaxSize         int64  `mapstructure:"UPLOAD_MAX_SIZE"`          // 单文件大小限制（字节）
	AllowedTypes    string `mapstructure:"UPLOAD_ALLOWED_TYPES"`     // 允许的 MIME 类型（逗号分隔）
	CleanInterval   int    `mapstructure:"UPLOAD_CLEAN_INTERVAL_SEC"` // 定时清理间隔（秒）
	TTLDays         int    `mapstructure:"UPLOAD_TTL_DAYS"`          // 上传文件保留天数
}

// LoginConfig 登录安全配置
type LoginConfig struct {
	MaxFails  int `mapstructure:"LOGIN_MAX_FAILS"` // 最大连续失败次数
	LockTTL   int `mapstructure:"LOGIN_LOCK_TTL"`  // 账户锁定时长（秒）
}

// Load 加载配置，configDir 为 configs/ 目录路径
func Load(configDir string) (*Config, error) {
	v := viper.New()

	// ── 设置默认值 ────────────────────────────────────────
	setDefaults(v)

	// ── 加载 .env 文件 ───────────────────────────────────
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		// .env 文件不存在不是错误，只是警告
		_ = err
	}

	// ── 加载 .env.{APP_ENV} 文件（覆盖 .env） ───────────
	appEnv := v.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	v.SetConfigName(fmt.Sprintf(".env.%s", appEnv))
	if err := v.MergeInConfig(); err != nil {
		_ = err
	}

	// ── 环境变量优先级最高 ───────────────────────────────
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ── 绑定到结构体 ─────────────────────────────────────
	cfg := &Config{}

	// 通用
	cfg.App.Env = v.GetString("APP_ENV")
	cfg.App.Port = v.GetInt("PORT")

	// 数据库
	cfg.Database.Type = v.GetString("DB_TYPE")
	cfg.Database.SQLite.Path = v.GetString("SQLITE_PATH")
	cfg.Database.MySQL.Host = v.GetString("DB_HOST")
	cfg.Database.MySQL.Port = v.GetInt("DB_PORT")
	cfg.Database.MySQL.User = v.GetString("DB_USER")
	cfg.Database.MySQL.Password = v.GetString("DB_PASS")
	cfg.Database.MySQL.Name = v.GetString("DB_NAME")
	cfg.Database.Postgres.Host = v.GetString("PG_HOST")
	cfg.Database.Postgres.Port = v.GetInt("PG_PORT")
	cfg.Database.Postgres.User = v.GetString("PG_USER")
	cfg.Database.Postgres.Password = v.GetString("PG_PASS")
	cfg.Database.Postgres.Name = v.GetString("PG_DB")

	// Redis
	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASS")

	// JWT
	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.Expires = v.GetString("JWT_EXPIRES")
	cfg.JWT.RefreshExpires = v.GetString("JWT_REFRESH_EXPIRES")

	// 安全
	cfg.Security.CSRFEnabled = v.GetBool("CSRF_ENABLED")
	cfg.Security.CSRFMode = v.GetString("CSRF_MODE")
	cfg.Security.CSRFSecret = v.GetString("CSRF_SECRET")
	cfg.Security.CORSOrigins = v.GetString("CORS_ORIGINS")

	// 限流
	cfg.Throttle.TTL = v.GetInt("THROTTLE_TTL")
	cfg.Throttle.Limit = v.GetInt("THROTTLE_LIMIT")

	// 指标
	cfg.Metrics.Enabled = v.GetBool("ENABLE_METRICS")

	// Swagger
	cfg.Swagger.Enabled = v.GetBool("SWAGGER_ENABLE")

	// 上传
	cfg.Upload.Dest = v.GetString("UPLOAD_DEST")
	cfg.Upload.MaxSize = v.GetInt64("UPLOAD_MAX_SIZE")
	cfg.Upload.AllowedTypes = v.GetString("UPLOAD_ALLOWED_TYPES")
	cfg.Upload.CleanInterval = v.GetInt("UPLOAD_CLEAN_INTERVAL_SEC")
	cfg.Upload.TTLDays = v.GetInt("UPLOAD_TTL_DAYS")

	// 登录安全
	cfg.Login.MaxFails = v.GetInt("LOGIN_MAX_FAILS")
	cfg.Login.LockTTL = v.GetInt("LOGIN_LOCK_TTL")

	return cfg, nil
}

// setDefaults 设置所有配置项的默认值
func setDefaults(v *viper.Viper) {
	// 通用
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", 8000)

	// 数据库
	v.SetDefault("DB_TYPE", "sqlite")
	v.SetDefault("SQLITE_PATH", "./data/sqlite.db")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 3306)
	v.SetDefault("DB_USER", "root")
	v.SetDefault("DB_PASS", "")
	v.SetDefault("DB_NAME", "hellogo")
	v.SetDefault("PG_HOST", "localhost")
	v.SetDefault("PG_PORT", 5432)
	v.SetDefault("PG_USER", "postgres")
	v.SetDefault("PG_PASS", "")
	v.SetDefault("PG_DB", "hellogo")

	// Redis
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASS", "")

	// JWT
	v.SetDefault("JWT_SECRET", "change_me_please")
	v.SetDefault("JWT_EXPIRES", "1d")
	v.SetDefault("JWT_REFRESH_EXPIRES", "7d")

	// 安全
	v.SetDefault("CSRF_ENABLED", true)
	v.SetDefault("CSRF_MODE", "header")
	v.SetDefault("CSRF_SECRET", "change-me")
	v.SetDefault("CORS_ORIGINS", "*")

	// 限流
	v.SetDefault("THROTTLE_TTL", 60)
	v.SetDefault("THROTTLE_LIMIT", 100)

	// 指标
	v.SetDefault("ENABLE_METRICS", true)

	// Swagger
	v.SetDefault("SWAGGER_ENABLE", true)

	// 上传
	v.SetDefault("UPLOAD_DEST", "./upload")
	v.SetDefault("UPLOAD_MAX_SIZE", 5242880)          // 5MB
	v.SetDefault("UPLOAD_ALLOWED_TYPES", "image/jpeg,image/png,application/pdf")
	v.SetDefault("UPLOAD_CLEAN_INTERVAL_SEC", 3600)   // 每小时清理一次
	v.SetDefault("UPLOAD_TTL_DAYS", 30)               // 保留 30 天

	// 登录安全
	v.SetDefault("LOGIN_MAX_FAILS", 5)
	v.SetDefault("LOGIN_LOCK_TTL", 600) // 10 分钟
}

// GetDSN 根据数据库类型返回 GORM 连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.MySQL.User, c.MySQL.Password,
			c.MySQL.Host, c.MySQL.Port, c.MySQL.Name)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Postgres.Host, c.Postgres.Port,
			c.Postgres.User, c.Postgres.Password, c.Postgres.Name)
	default: // sqlite
		return c.SQLite.Path
	}
}
