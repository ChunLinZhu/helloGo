// Package response — 统一响应格式单元测试
package response

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseBody 解析 Fiber 测试响应体为 Response 结构体
func parseBody(t *testing.T, resp *http.Response) Response {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var r Response
	err = json.Unmarshal(body, &r)
	require.NoError(t, err)
	return r
}

func TestSuccessOK(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessOK(c, fiber.Map{"name": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	r := parseBody(t, resp)
	assert.Equal(t, "OK", r.Code)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "success", r.Message)
	assert.Equal(t, "/test", r.Path)
	assert.NotEmpty(t, r.Timestamp)
}

func TestSuccessCreated(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return SuccessCreated(c, fiber.Map{"id": "123"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	r := parseBody(t, resp)
	assert.Equal(t, "OK", r.Code)
	assert.Equal(t, "created", r.Message)
}

func TestSuccessPage(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		items := []string{"a", "b", "c"}
		return SuccessPage(c, items, 25, 2, 10)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	r := parseBody(t, resp)
	assert.Equal(t, "OK", r.Code)

	// data 是 PageData 结构，需要再次解析
	dataBytes, err := json.Marshal(r.Data)
	require.NoError(t, err)
	var pageData PageData
	err = json.Unmarshal(dataBytes, &pageData)
	require.NoError(t, err)

	assert.Equal(t, int64(25), pageData.Total)
	assert.Equal(t, 2, pageData.Page)
	assert.Equal(t, 10, pageData.Limit)
	assert.Equal(t, 3, pageData.TotalPages) // 25/10 = 2 余 5 → 3 页
}

func TestSuccessPage_ExactDivision(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessPage(c, []string{}, 20, 1, 10)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	r := parseBody(t, resp)
	dataBytes, _ := json.Marshal(r.Data)
	var pageData PageData
	json.Unmarshal(dataBytes, &pageData)

	assert.Equal(t, 2, pageData.TotalPages, "20/10 应恰好为 2 页")
}

func TestGetRequestID_Present(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("requestId", "trace-123")
		return c.Next()
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessOK(c, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	r := parseBody(t, resp)
	assert.Equal(t, "trace-123", r.RequestID)
}

func TestGetRequestID_Missing(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessOK(c, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	r := parseBody(t, resp)
	assert.Equal(t, "", r.RequestID, "缺少 requestId 时应返回空字符串")
}
