// Package pagination — 分页工具单元测试
package pagination

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// testGetPagination 发送请求并在 handler 内调用 GetPagination
func testGetPagination(t *testing.T, url string) Pagination {
	t.Helper()
	var result Pagination
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result = GetPagination(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", url, nil)
	_, err := app.Test(req)
	assert.NoError(t, err)
	return result
}

func TestGetPagination_Defaults(t *testing.T) {
	p := testGetPagination(t, "/test")
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.Limit)
}

func TestGetPagination_CustomValues(t *testing.T) {
	p := testGetPagination(t, "/test?page=3&limit=25")
	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 25, p.Limit)
}

func TestGetPagination_NegativePage(t *testing.T) {
	p := testGetPagination(t, "/test?page=-1&limit=5")
	assert.Equal(t, 1, p.Page, "负数页码应回退到 1")
	assert.Equal(t, 5, p.Limit)
}

func TestGetPagination_NegativeLimit(t *testing.T) {
	p := testGetPagination(t, "/test?page=1&limit=-5")
	assert.Equal(t, 10, p.Limit, "负数 Limit 应回退到默认值 10")
}

func TestGetPagination_ExceedMaxLimit(t *testing.T) {
	p := testGetPagination(t, "/test?page=1&limit=500")
	assert.Equal(t, 100, p.Limit, "超过 100 应限制为 100")
}

func TestGetPagination_OnlyPage(t *testing.T) {
	p := testGetPagination(t, "/test?page=5")
	assert.Equal(t, 5, p.Page)
	assert.Equal(t, 10, p.Limit, "未指定 limit 时应使用默认值 10")
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name   string
		page   int
		limit  int
		offset int
	}{
		{"第 1 页", 1, 10, 0},
		{"第 2 页", 2, 10, 10},
		{"第 3 页 limit=25", 3, 25, 50},
		{"第 1 页 limit=1", 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			assert.Equal(t, tt.offset, p.Offset())
		})
	}
}

func TestPaginate_ReturnsScope(t *testing.T) {
	p := Pagination{Page: 2, Limit: 15}
	scope := Paginate(p)
	assert.NotNil(t, scope, "Paginate 应返回非 nil 的 GORM scope 函数")
}
