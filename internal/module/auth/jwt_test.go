// Package auth — JWT 工具单元测试
package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService() *JWTService {
	return NewJWTService(JWTConfig{
		Secret:         "test-secret-key",
		Expires:        "1h",
		RefreshExpires: "7d",
	})
}

// ── AccessToken 测试 ───────────────────────────────────────

func TestGenerateAccessToken_Success(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateAccessToken("user-1", "alice", []string{"admin"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateAccessToken_Success(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateAccessToken("user-1", "alice", []string{"admin", "user"})
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, []string{"admin", "user"}, claims.Roles)
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	svc := newTestJWTService()
	token, _ := svc.GenerateAccessToken("user-1", "alice", []string{"admin"})

	otherSvc := NewJWTService(JWTConfig{
		Secret:  "wrong-secret",
		Expires: "1h",
	})
	claims, err := otherSvc.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	svc := NewJWTService(JWTConfig{
		Secret:  "test-secret",
		Expires: "1s",
	})

	// 手动创建已过期的 token
	claims := Claims{
		UserID:   "user-1",
		Username: "alice",
		Roles:    []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "helloGo",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	parsed, err := svc.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, parsed)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateAccessToken_InvalidFormat(t *testing.T) {
	svc := newTestJWTService()

	claims, err := svc.ValidateToken("not-a-valid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateAccessToken_EmptyToken(t *testing.T) {
	svc := newTestJWTService()

	claims, err := svc.ValidateToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// ── RefreshToken 测试 ──────────────────────────────────────

func TestGenerateRefreshToken_Success(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateRefreshToken("user-1", "session-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateRefreshToken("user-1", "session-123")
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "session-123", claims.SessionID)
}

func TestValidateRefreshToken_WrongSecret(t *testing.T) {
	svc := newTestJWTService()
	token, _ := svc.GenerateRefreshToken("user-1", "session-123")

	otherSvc := NewJWTService(JWTConfig{
		Secret:         "wrong-secret",
		RefreshExpires: "7d",
	})
	claims, err := otherSvc.ValidateRefreshToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// ── parseDuration 测试 ─────────────────────────────────────

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1d", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"2h", 2 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"60s", 60 * time.Second, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := parseDuration(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, d)
			}
		})
	}
}

// ── 跨 token 类型验证 ──────────────────────────────────────

func TestAccessToken_ParsedAsRefresh_HasEmptySessionID(t *testing.T) {
	svc := newTestJWTService()

	// access token 中不含 SessionID，用 RefreshClaims 解析时 SessionID 为空
	accessToken, _ := svc.GenerateAccessToken("user-1", "alice", []string{"admin"})

	claims, err := svc.ValidateRefreshToken(accessToken)
	// JWT 库不区分自定义字段，只要签名有效就会解析成功
	// 但 SessionID 字段在 access token 中不存在，解析结果为空字符串
	require.NoError(t, err)
	assert.Equal(t, "", claims.SessionID, "access token 不含 SessionID")
}
