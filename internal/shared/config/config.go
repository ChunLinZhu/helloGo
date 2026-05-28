// Package config — 微服务共享配置加载器
// 服务导向的 Viper 配置，独立于 Phase 1 的 internal/config
// 优先级：环境变量 > {service}.env > .env.{APP_ENV} > .env > 默认值
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 微服务共享配置项
type Config struct {
	Service  ServiceConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// ServiceConfig 服务基本信息
type ServiceConfig struct {
	Name     string // 服务名称：user / auth / permission / biz / gateway
	Env      string // 运行环境：development / production / test
	GRPCPort int    // gRPC 监听端口（50001-50004）
	HTTPPort int    // HTTP 监听端口（仅 Gateway 使用：8000）
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string // 数据库类型：sqlite / mysql / postgres
	SQLite   SQLiteConfig
	MySQL    MySQLConfig
	Postgres PostgresConfig
}

// SQLiteConfig SQLite 专属配置
type SQLiteConfig struct {
	Path string // 数据库文件路径
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret         string // 签名密钥
	Expires        string // access token 过期时间（如 "1d", "2h"）
	RefreshExpires string // refresh token 过期时间（如 "7d"）
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

// 各服务的默认 gRPC 端口
var defaultGRPCPorts = map[string]int{
	"user":       50001,
	"auth":       50002,
	"permission": 50003,
	"biz":        50004,
	"gateway":    8000,
}

// Load 加载指定服务的配置
// serviceName: 服务名称（用于设置默认端口）
// configDir: 配置文件目录（通常为 "./configs"）
func Load(serviceName string, configDir string) (*Config, error) {
	v := viper.New()

	// ── 设置默认值 ────────────────────────────────────────
	setDefaults(v, serviceName)

	// ── 加载 .env 文件（基础配置）─────────────────────────
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		_ = err // .env 文件不存在不是错误
	}

	// ── 加载 .env.{APP_ENV} 文件（环境配置，覆盖 .env）─────
	appEnv := v.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	v.SetConfigName(fmt.Sprintf(".env.%s", appEnv))
	if err := v.MergeInConfig(); err != nil {
		_ = err
	}

	// ── 加载服务专属 .env 文件（最高文件优先级）─────────────
	// 例如 configs/user.env — 覆盖 .env 和 .env.{APP_ENV}
	v.SetConfigName(serviceName)
	v.SetConfigType("env")
	v.AddConfigPath(configDir)
	if err := v.MergeInConfig(); err != nil {
		_ = err
	}

	// ── 环境变量优先级最高 ─────────────────────────────────
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ── 绑定到结构体 ─────────────────────────────────────
	cfg := &Config{}

	// 服务
	cfg.Service.Name = serviceName
	cfg.Service.Env = v.GetString("APP_ENV")
	cfg.Service.GRPCPort = v.GetInt("GRPC_PORT")
	cfg.Service.HTTPPort = v.GetInt("HTTP_PORT")

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

	return cfg, nil
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper, serviceName string) {
	// 服务
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("GRPC_PORT", defaultGRPCPorts[serviceName])
	v.SetDefault("HTTP_PORT", 8000)

	// 数据库
	v.SetDefault("DB_TYPE", "sqlite")
	v.SetDefault("SQLITE_PATH", "./data/"+serviceName+".db")
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
}
