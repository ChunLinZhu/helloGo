// Package department — 部门请求/响应结构体
package department

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=256"`
	ParentID    *string `json:"parentId" validate:"omitempty"`
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=256"`
	ParentID    *string `json:"parentId" validate:"omitempty"`
}

// DepartmentResponse 部门响应（含子部门）
type DepartmentResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description *string               `json:"description"`
	ParentID    *string               `json:"parentId"`
	Children    []*DepartmentResponse `json:"children"`
}

// ToDepartmentResponse 将 Department 模型转换为响应结构（不含子部门）
func ToDepartmentResponse(d *Department) *DepartmentResponse {
	return &DepartmentResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		ParentID:    d.ParentID,
		Children:    make([]*DepartmentResponse, 0),
	}
}

// BuildDepartmentTree 将扁平部门列表构建为树结构
func BuildDepartmentTree(depts []Department) []*DepartmentResponse {
	nodeMap := make(map[string]*DepartmentResponse)
	roots := make([]*DepartmentResponse, 0)

	// 第一遍：创建所有节点
	for i := range depts {
		node := ToDepartmentResponse(&depts[i])
		nodeMap[depts[i].ID] = node
	}

	// 第二遍：构建父子关系
	for i := range depts {
		node := nodeMap[depts[i].ID]
		if depts[i].ParentID == nil || *depts[i].ParentID == "" {
			roots = append(roots, node)
		} else {
			parent, ok := nodeMap[*depts[i].ParentID]
			if ok {
				parent.Children = append(parent.Children, node)
			} else {
				// 父节点不存在，作为根节点处理
				roots = append(roots, node)
			}
		}
	}

	return roots
}
