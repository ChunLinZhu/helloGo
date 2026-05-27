// Package permission — 权限请求/响应结构体
package permission

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Key         string  `json:"key" validate:"required,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=256"`
	RoleID      string  `json:"roleId" validate:"required"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Key         *string `json:"key" validate:"omitempty,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=256"`
	RoleID      *string `json:"roleId" validate:"omitempty"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Description *string `json:"description"`
	RoleID      string  `json:"roleId"`
}

// ToPermissionResponse 将 Permission 模型转换为响应结构
func ToPermissionResponse(p *Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Key:         p.Key,
		Description: p.Description,
		RoleID:      p.RoleID,
	}
}
