// Package dict — 字典请求/响应结构体
package dict

// CreateDictRequest 创建字典项请求
type CreateDictRequest struct {
	Type        string  `json:"type" validate:"required,min=2,max=128"`
	Key         string  `json:"key" validate:"required,min=2,max=128"`
	Value       string  `json:"value" validate:"required,min=1,max=256"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

// UpdateDictRequest 更新字典项请求
type UpdateDictRequest struct {
	Type        *string `json:"type" validate:"omitempty,min=2,max=128"`
	Key         *string `json:"key" validate:"omitempty,min=2,max=128"`
	Value       *string `json:"value" validate:"omitempty,min=1,max=256"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

// DictResponse 字典响应
type DictResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description"`
}

// ToDictResponse 将 Dict 模型转换为响应结构
func ToDictResponse(d *Dict) *DictResponse {
	return &DictResponse{
		ID:          d.ID,
		Type:        d.Type,
		Key:         d.Key,
		Value:       d.Value,
		Description: d.Description,
	}
}
