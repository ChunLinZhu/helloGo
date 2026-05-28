// Package health — K8s 健康检查探针
// 提供 /healthz（存活）和 /readyz（就绪）HTTP 端点，供 K8s 探针调用
package health

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CheckFunc 就绪检查函数，返回 nil 表示健康
type CheckFunc func() error

// Server 健康检查 HTTP 服务器
type Server struct {
	port   int
	checks []CheckFunc
	logger *zap.Logger
	srv    *http.Server
}

// NewServer 创建健康检查服务器
// port: 监听端口（通常为 8080）
// checks: 就绪检查函数列表（任一失败则 /readyz 返回 503）
func NewServer(port int, logger *zap.Logger, checks ...CheckFunc) *Server {
	return &Server{
		port:   port,
		checks: checks,
		logger: logger,
	}
}

// Start 非阻塞启动健康检查服务器
func (s *Server) Start() {
	mux := http.NewServeMux()

	// /healthz — 存活探针（进程运行即返回 200）
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// /readyz — 就绪探针（所有 check 通过才返回 200）
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		for _, check := range s.checks {
			if err := check(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, "not ready: %v", err)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	go func() {
		s.logger.Info("健康检查服务启动", zap.Int("port", s.port))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("健康检查服务异常", zap.Error(err))
		}
	}()
}

// Stop 优雅关闭
func (s *Server) Stop() {
	if s.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.srv.Shutdown(ctx)
}

// DBCheck 返回数据库连接检查函数
// 传入 *sql.DB（通过 gorm.DB.DB() 获取）
func DBCheck(db *sql.DB) CheckFunc {
	return func() error {
		if db == nil {
			return fmt.Errorf("database not initialized")
		}
		return db.Ping()
	}
}

// RedisCheck 返回 Redis 连接检查函数
// pingFunc 通常为 redisClient.Ping
func RedisCheck(pingFunc func(context.Context) error) CheckFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return pingFunc(ctx)
	}
}
