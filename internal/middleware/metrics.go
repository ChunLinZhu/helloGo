package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Metrics 定义自定义指标
var (
	// HTTP 请求总数计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求持续时间直方图（毫秒）
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_ms",
			Help:    "HTTP request duration in milliseconds",
			Buckets: []float64{50, 100, 200, 500, 1000, 2000, 5000},
		},
		[]string{"method", "path"},
	)
)

func init() {
	// 注册自定义指标
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// MetricsMiddleware Prometheus 指标中间件
// 记录每个请求的计数器和持续时间
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		// 在调用 c.Next() 之前捕获请求信息
		// 使用 byte slice 复制避免引用 Fiber 内部缓冲区
		method := c.Method()
		pathBytes := c.Path()
		pathCopy := make([]byte, len(pathBytes))
		copy(pathCopy, pathBytes)
		path := string(pathCopy)

		// 执行后续处理
		err := c.Next()

		// 计算持续时间（毫秒）
		duration := float64(time.Since(start).Milliseconds())
		status := strconv.Itoa(c.Response().StatusCode())

		// 记录指标
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

// MetricsHandler 返回 Prometheus 指标处理器
// 将标准库 http.Handler 适配为 Fiber 处理器
func MetricsHandler() fiber.Handler {
	// 使用 promhttp 提供的标准指标处理器
	h := promhttp.Handler()

	return func(c *fiber.Ctx) error {
		// 使用 fasthttpadaptor 将标准库 handler 转换为 Fiber 兼容
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	}
}
