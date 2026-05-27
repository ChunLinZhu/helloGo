// Package auth — JWT 令牌工具
// 生成和验证 access token / refresh token
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷声明
type Claims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// RefreshClaims refresh token 载荷声明
type RefreshClaims struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
	jwt.RegisteredClaims
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret         string // 签名密钥
	Expires        string // access token 过期时间（如 "1d", "2h"）
	RefreshExpires string // refresh token 过期时间（如 "7d"）
}

// JWTService JWT 服务
type JWTService struct {
	config JWTConfig
}

// NewJWTService 创建 JWT 服务
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

// GenerateAccessToken 生成 access token
func (s *JWTService) GenerateAccessToken(userID, username string, roles []string) (string, error) {
	expires, err := parseDuration(s.config.Expires)
	if err != nil {
		return "", fmt.Errorf("解析 access token 过期时间失败: %w", err)
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "helloGo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

// GenerateRefreshToken 生成 refresh token
func (s *JWTService) GenerateRefreshToken(userID, sessionID string) (string, error) {
	expires, err := parseDuration(s.config.RefreshExpires)
	if err != nil {
		return "", fmt.Errorf("解析 refresh token 过期时间失败: %w", err)
	}

	claims := RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "helloGo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

// ValidateToken 验证并解析 access token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken 验证并解析 refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// parseDuration 解析时间字符串（如 "1d", "2h", "30m"）
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// 支持 d (天), h (小时), m (分钟), s (秒)
	unit := s[len(s)-1]
	value := s[:len(s)-1]

	var multiplier time.Duration
	switch unit {
	case 'd':
		multiplier = 24 * time.Hour
	case 'h':
		multiplier = time.Hour
	case 'm':
		multiplier = time.Minute
	case 's':
		multiplier = time.Second
	default:
		// 如果没有单位，尝试直接解析
		return time.ParseDuration(s)
	}

	var num int
	_, err := fmt.Sscanf(value, "%d", &num)
	if err != nil {
		return 0, err
	}

	return time.Duration(num) * multiplier, nil
}
