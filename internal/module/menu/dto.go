// Package menu — 菜单请求/响应结构体
package menu

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=128"`
	Path     *string `json:"path" validate:"omitempty,max=256"`
	Icon     *string `json:"icon" validate:"omitempty,max=128"`
	Order    int     `json:"order"`
	ParentID *string `json:"parentId" validate:"omitempty"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=2,max=128"`
	Path     *string `json:"path" validate:"omitempty,max=256"`
	Icon     *string `json:"icon" validate:"omitempty,max=128"`
	Order    *int    `json:"order"`
	ParentID *string `json:"parentId" validate:"omitempty"`
}

// MenuResponse 菜单响应（含子菜单）
type MenuResponse struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Path     *string         `json:"path"`
	Icon     *string         `json:"icon"`
	Order    int             `json:"order"`
	ParentID *string         `json:"parentId"`
	Children []*MenuResponse `json:"children"`
}

// ToMenuResponse 将 Menu 模型转换为响应结构（不含子菜单）
func ToMenuResponse(m *Menu) *MenuResponse {
	return &MenuResponse{
		ID:       m.ID,
		Name:     m.Name,
		Path:     m.Path,
		Icon:     m.Icon,
		Order:    m.Order,
		ParentID: m.ParentID,
		Children: make([]*MenuResponse, 0),
	}
}

// BuildMenuTree 将扁平菜单列表构建为树结构
func BuildMenuTree(menus []Menu) []*MenuResponse {
	// 创建响应切片和索引映射
	nodeMap := make(map[string]*MenuResponse)
	roots := make([]*MenuResponse, 0)

	// 第一遍：创建所有节点
	for i := range menus {
		node := ToMenuResponse(&menus[i])
		nodeMap[menus[i].ID] = node
	}

	// 第二遍：构建父子关系
	for i := range menus {
		node := nodeMap[menus[i].ID]
		if menus[i].ParentID == nil || *menus[i].ParentID == "" {
			roots = append(roots, node)
		} else {
			parent, ok := nodeMap[*menus[i].ParentID]
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
