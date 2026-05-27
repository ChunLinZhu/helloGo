// Package role — 角色请求/响应结构体
package role

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string  `json:"code" validate:"required,min=2,max=64"`
	Name        string  `json:"name" validate:"required,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permissionIds" validate:"required"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"` // 权限 key 列表
}

// ToRoleResponse 将 Role 模型转换为响应结构
func ToRoleResponse(r *Role) *RoleResponse {
	perms := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = p.Key
	}

	return &RoleResponse{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Permissions: perms,
	}
}
