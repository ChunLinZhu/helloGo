// Package pagination — 框架无关的分页工具
// 从 Phase 1 的 internal/pkg/pagination 改造，移除 Fiber 和 GORM 依赖
package pagination

// Pagination 分页参数
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// New 创建分页参数，带边界检查
// page 默认 1，limit 默认 10，最大 100
func New(page, limit int) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return Pagination{Page: page, Limit: limit}
}

// Offset 计算 SQL OFFSET 值
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// TotalPages 根据总记录数计算总页数
func (p Pagination) TotalPages(total int64) int {
	if total == 0 {
		return 0
	}
	pages := int(total) / p.Limit
	if int(total)%p.Limit > 0 {
		pages++
	}
	return pages
}
