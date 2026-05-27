// Package log — 日志请求/响应结构体
package log

// CreateLogRequest 创建日志请求
type CreateLogRequest struct {
	Level   string  `json:"level" validate:"required,oneof=info warn error debug"`
	Message string  `json:"message" validate:"required,min=1,max=256"`
	Meta    *string `json:"meta" validate:"omitempty"`
}

// LogResponse 日志响应
type LogResponse struct {
	ID        string  `json:"id"`
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	Meta      *string `json:"meta"`
	CreatedAt string  `json:"createdAt"`
}

// ToLogResponse 将 Log 模型转换为响应结构
func ToLogResponse(l *Log) *LogResponse {
	return &LogResponse{
		ID:        l.ID,
		Level:     l.Level,
		Message:   l.Message,
		Meta:      l.Meta,
		CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
