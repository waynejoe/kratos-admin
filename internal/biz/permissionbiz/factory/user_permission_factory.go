package factory

import (
	"slices"

	"kratos-admin/internal/biz/permissionbiz/valueobject"
	pb "kratos-admin/pb/admin/v1"
)

// NewUserPermission 创建用户权限
func NewUserPermission(userPermissions []string, resources []*pb.ResourceInfo) ([]string, []*pb.UserPermissioPage) {
	var (
		buttons = make([]string, 0)
		pages   = make([]*pb.UserPermissioPage, 0)
	)

	for _, item := range resources {
		if !slices.Contains(userPermissions, newPermissionObject(item)) {
			continue
		}

		resourceType := valueobject.ResourceType(item.Type)

		switch resourceType {
		case valueobject.MenuResourceType:
			page := &pb.UserPermissioPage{Path: item.Path}

			pages = append(pages, page)

			if len(item.Children) == 0 {
				continue
			}

			childrenButtons, childrenPages := NewUserPermission(userPermissions, item.Children)

			page.Children = childrenPages
			buttons = append(buttons, childrenButtons...)
		case valueobject.ButtonResourceType:
			buttons = append(buttons, item.Path)
		}
	}

	return buttons, pages
}
