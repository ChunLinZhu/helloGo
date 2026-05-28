// Package database — GORM 数据库初始化
// 支持 SQLite / MySQL / PostgreSQL 三种数据库，通过配置切换；
// 包含连接池配置和日志级别控制
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"helloGo/internal/config"
)

// ensureDatabaseExists 确保 MySQL/PostgreSQL 数据库存在
func ensureDatabaseExists(cfg config.DatabaseConfig, log *zap.Logger) error {
	switch cfg.Type {
	case "mysql":
		// 连接服务器（不指定数据库）
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
			cfg.MySQL.User, cfg.MySQL.Password,
			cfg.MySQL.Host, cfg.MySQL.Port)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("连接 MySQL 服务器失败: %w", err)
		}
		defer db.Close()
		// 建库
		if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4", cfg.MySQL.Name)); err != nil {
			return fmt.Errorf("创建数据库失败: %w", err)
		}
		log.Info("MySQL 数据库就绪", zap.String("name", cfg.MySQL.Name))
	case "postgres":
		// 先连到默认的 postgres 库，再检查/创建目标库
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
			cfg.Postgres.Host, cfg.Postgres.Port,
			cfg.Postgres.User, cfg.Postgres.Password)
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return fmt.Errorf("连接 PostgreSQL 服务器失败: %w", err)
		}
		defer db.Close()
		// 检查库是否存在
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Postgres.Name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("检查数据库失败: %w", err)
		}
		if !exists {
			// CREATE DATABASE 不能在事务中执行，也不能用参数化查询
			// 库名来自配置（可控），直接拼接
			if _, err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.Postgres.Name)); err != nil {
				return fmt.Errorf("创建数据库失败: %w", err)
			}
			log.Info("PostgreSQL 数据库已创建", zap.String("name", cfg.Postgres.Name))
		} else {
			log.Info("PostgreSQL 数据库已存在", zap.String("name", cfg.Postgres.Name))
		}
	}
	return nil
}

// Init 初始化 GORM 数据库连接
func Init(cfg config.DatabaseConfig, appEnv string, log *zap.Logger) (*gorm.DB, error) {
	// ── 确保数据库存在 ──────────────────────────────────
	if err := ensureDatabaseExists(cfg, log); err != nil {
		return nil, err
	}

	// ── 选择数据库驱动 ────────────────────────────────────
	var dialector gorm.Dialector
	dsn := cfg.GetDSN()

	switch cfg.Type {
	case "mysql":
		dialector = mysql.Open(dsn)
		log.Info("使用 MySQL 数据库",
			zap.String("host", cfg.MySQL.Host),
			zap.Int("port", cfg.MySQL.Port),
			zap.String("database", cfg.MySQL.Name),
		)
	case "postgres":
		dialector = postgres.Open(dsn)
		log.Info("使用 PostgreSQL 数据库",
			zap.String("host", cfg.Postgres.Host),
			zap.Int("port", cfg.Postgres.Port),
			zap.String("database", cfg.Postgres.Name),
		)
	default: // sqlite
		// 确保 SQLite 文件目录存在
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建 SQLite 目录失败: %w", err)
		}
		dialector = sqlite.Open(dsn)
		log.Info("使用 SQLite 数据库", zap.String("path", dsn))
	}

	// ── 日志级别：开发环境详细，生产环境静默 ────────────────
	logLevel := logger.Warn
	if appEnv == "development" {
		logLevel = logger.Info
	}

	// ── 打开数据库连接 ────────────────────────────────────
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// ── 连接池配置（MySQL / PostgreSQL 生效） ─────────────
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 *sql.DB 失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)                 // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)                 // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(30 * 60e9)       // 连接最大存活时间（30 分钟）

	// ── SQLite 专属优化 ──────────────────────────────────
	if cfg.Type == "sqlite" {
		if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
			log.Warn("SQLite WAL 模式设置失败", zap.Error(err))
		}
		if err := db.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
			log.Warn("SQLite busy_timeout 设置失败", zap.Error(err))
		}
	}

	log.Info("数据库连接成功")
	return db, nil
}
