// Package errors — 错误码与 i18n 单元测试
package errors

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestBadRequest(t *testing.T) {
	err := BadRequest("参数不合法")
	assert.Equal(t, CodeValidationError, err.Code)
	assert.Equal(t, fiber.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "参数不合法", err.Detail)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized()
	assert.Equal(t, CodeUnauthorized, err.Code)
	assert.Equal(t, fiber.StatusUnauthorized, err.HTTPStatus)
}

func TestForbidden(t *testing.T) {
	err := Forbidden()
	assert.Equal(t, CodeForbidden, err.Code)
	assert.Equal(t, fiber.StatusForbidden, err.HTTPStatus)
}

func TestNotFound(t *testing.T) {
	err := NotFound()
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, fiber.StatusNotFound, err.HTTPStatus)
}

func TestInternal(t *testing.T) {
	err := Internal()
	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, fiber.StatusInternalServerError, err.HTTPStatus)
}

func TestAppError_Error(t *testing.T) {
	err := BadRequest("test detail")
	str := err.Error()
	assert.Contains(t, str, "VALIDATION_ERROR")
}

func TestGetMessage_ZhCN(t *testing.T) {
	err := &AppError{Code: CodeUnauthorized}
	msg := err.GetMessage("zh-CN")
	assert.Equal(t, "未认证或认证失效", msg)
}

func TestGetMessage_EnUS(t *testing.T) {
	err := &AppError{Code: CodeUnauthorized}
	msg := err.GetMessage("en-US")
	assert.Equal(t, "Unauthorized", msg)
}

func TestGetMessage_UnknownLang_FallbackZhCN(t *testing.T) {
	err := &AppError{Code: CodeNotFound}
	msg := err.GetMessage("fr-FR")
	assert.Equal(t, "资源不存在", msg, "未知语言应回退到 zh-CN")
}

func TestGetMessage_CustomMessage_TakesPrecedence(t *testing.T) {
	err := &AppError{Code: CodeNotFound, Message: "自定义消息"}
	msg := err.GetMessage("zh-CN")
	assert.Equal(t, "自定义消息", msg, "自定义 Message 应优先于 i18n")
}

func TestGetMessage_UnknownCode(t *testing.T) {
	err := &AppError{Code: ErrorCode("NON_EXISTENT")}
	msg := err.GetMessage("zh-CN")
	assert.Equal(t, "未知错误", msg, "未知错误码应返回 UNKNOWN 消息")
}

func TestHTTPStatusToErrorCode(t *testing.T) {
	tests := []struct {
		status int
		code   ErrorCode
	}{
		{400, CodeValidationError},
		{401, CodeUnauthorized},
		{403, CodeForbidden},
		{404, CodeNotFound},
		{500, CodeInternalError},
		{502, CodeInternalError},
		{418, CodeHTTPError},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.code, HTTPStatusToErrorCode(tt.status))
		})
	}
}

func TestGetLang_Default(t *testing.T) {
	app := fiber.New()
	var lang string
	app.Get("/test", func(c *fiber.Ctx) error {
		lang = GetLang(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req) //nolint:errcheck
	assert.Equal(t, "zh-CN", lang)
}

func TestGetLang_XLang(t *testing.T) {
	app := fiber.New()
	var lang string
	app.Get("/test", func(c *fiber.Ctx) error {
		lang = GetLang(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Lang", "en-US")
	app.Test(req) //nolint:errcheck
	assert.Equal(t, "en-US", lang)
}

func TestGetLang_AcceptLanguage(t *testing.T) {
	app := fiber.New()
	var lang string
	app.Get("/test", func(c *fiber.Ctx) error {
		lang = GetLang(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	app.Test(req) //nolint:errcheck
	assert.Equal(t, "en-US", lang)
}

func TestGetLang_XLang_TakesPrecedence(t *testing.T) {
	app := fiber.New()
	var lang string
	app.Get("/test", func(c *fiber.Ctx) error {
		lang = GetLang(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Lang", "en")
	req.Header.Set("Accept-Language", "zh-CN")
	app.Test(req) //nolint:errcheck
	assert.Equal(t, "en-US", lang, "X-Lang 应优先于 Accept-Language")
}

func TestNormalizeLang(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "en-US"},
		{"en-US", "en-US"},
		{"en_US", "en-US"},
		{"en-us", "en-US"},
		{"zh", "zh-CN"},
		{"zh-CN", "zh-CN"},
		{"  en-US  ", "en-US"},
		{"en-US,en;q=0.9", "en-US"},
		{"fr", "zh-CN"}, // 不支持的语言回退到 zh-CN
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeLang(tt.input))
		})
	}
}
