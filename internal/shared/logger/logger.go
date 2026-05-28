// Package logger — 共享 Zap 日志工厂函数
// 从 Phase 1 的 cmd/server/main.go initLogger 提取，增加 serviceName 参数
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New 创建 *zap.Logger，根据环境选择输出方式
// env="production" → JSON 格式写入文件（带轮转）；其他 → 彩色控制台输出
// serviceName 会作为默认字段添加到每条日志中（如 [service=user]）
func New(env string, serviceName string) *zap.Logger {
	// 日志轮转配置
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logs/" + serviceName + ".log",
		MaxSize:    50,  // 每个日志文件最大 50MB
		MaxBackups: 10,  // 最多保留 10 个旧文件
		MaxAge:     30,  // 旧文件最多保留 30 天
		Compress:   true, // 旧文件压缩
	}

	// 默认字段：标识服务名
	defaultFields := []zap.Field{zap.String("service", serviceName)}

	if env == "production" {
		// 生产环境：JSON 格式写入文件
		encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(lumberjackLogger),
			zap.InfoLevel,
		)
		return zap.New(core, zap.AddCaller()).With(defaultFields...)
	}

	// 开发环境：彩色控制台输出
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)
	return zap.New(core, zap.AddCaller()).With(defaultFields...)
}
