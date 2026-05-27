// Package pagination — 分页参数绑定与 GORM Scope
// 从 query string 自动解析 page/limit，生成 GORM 分页 Scope
package pagination

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Pagination 分页参数
type Pagination struct {
	Page  int `json:"page" query:"page"`   // 页码，默认 1
	Limit int `json:"limit" query:"limit"` // 每页条数，默认 10
}

// GetPagination 从 Fiber context 的 query string 中解析分页参数
func GetPagination(c *fiber.Ctx) Pagination {
	p := Pagination{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}
	// 边界校验
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100 // 防止一次拉取过多数据
	}
	return p
}

// Offset 计算 SQL OFFSET 值
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Paginate 返回 GORM Scope，用于链式调用
// 用法：db.Scopes(pagination.Paginate(p)).Find(&users)
func Paginate(p Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.Offset()).Limit(p.Limit)
	}
}
