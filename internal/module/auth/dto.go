// Package auth — 认证相关请求/响应结构体
package auth

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=2,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	SessionID    string `json:"sessionId"`
}

// RefreshRequest 刷新 token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	SessionID    string `json:"sessionId" validate:"required"`
}

// RefreshResponse 刷新 token 响应
type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	SessionID    string `json:"sessionId"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	SessionID string `json:"sessionId" validate:"required"`
}

// RequestPasswordResetRequest 请求密码重置
type RequestPasswordResetRequest struct {
	Username string `json:"username" validate:"required"`
}

// RequestPasswordResetResponse 请求密码重置响应
type RequestPasswordResetResponse struct {
	Token string `json:"token"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Username    string `json:"username" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6,max=128"`
	Token       string `json:"token" validate:"required"`
}

// UnlockRequest 解锁账户请求
type UnlockRequest struct {
	Username string `json:"username" validate:"required"`
}
