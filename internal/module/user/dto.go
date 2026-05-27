// Package user — 用户请求/响应结构体
package user

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string   `json:"username" validate:"required,min=2,max=64"`
	Password string   `json:"password" validate:"required,min=6,max=128"`
	Email    *string  `json:"email" validate:"omitempty,email"`
	Phone    *string  `json:"phone" validate:"omitempty"`
	IsActive *bool    `json:"isActive"`
	RoleIDs  []string `json:"roleIds"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    *string  `json:"email" validate:"omitempty,email"`
	Phone    *string  `json:"phone" validate:"omitempty"`
	IsActive *bool    `json:"isActive"`
	RoleIDs  []string `json:"roleIds"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    *string  `json:"email"`
	Phone    *string  `json:"phone"`
	IsActive bool     `json:"isActive"`
	Roles    []string `json:"roles"` // 角色代码列表
}

// ToUserResponse 将 User 模型转换为响应结构
func ToUserResponse(u *User) *UserResponse {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Code
	}

	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Phone:    u.Phone,
		IsActive: u.IsActive,
		Roles:    roles,
	}
}
