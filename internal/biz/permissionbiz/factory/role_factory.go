package factory

import (
	"context"
	"fmt"
	"slices"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"

	toolboxauthz "kratos-admin/pkg/toolbox/authz"
	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/utils"
)

// NewRoleSubject 创建角色权限key
func NewRoleSubject(roleId int64) string {
	return fmt.Sprintf("%s%d", authz.RolePermissionPrefix, roleId)
}

// NewRoleModel 创建角色模型
func NewRoleModel(ctx context.Context, req *pb.SaveRoleRequest) *adminmodel.AdminRole {
	data := utils.CopyPtr[adminmodel.AdminRole](req)

	data.Status = adminmodel.StatusOn
	data.OperatorId = helpx.GetUserId(ctx)

	return data
}

// NewRoleInfo 创建角色信息
func NewRoleInfo(
	in *adminmodel.AdminRole,
	groups map[int64]*adminmodel.AdminGroup,
	roleUsers map[int64][]int64,
	operators map[int64]*adminmodel.AdminUser,
) *pb.RoleInfo {
	data := utils.CopyPtr[pb.RoleInfo](in)

	data.CreateTime = in.CreateTime.UnixMilli()

	if group, ok := groups[in.GroupId]; ok {
		data.Group = utils.CopyPtr[pb.RoleInfo_RoleGroup](group)
	}

	data.UserCount = int32(len(roleUsers[in.Id]))

	if operator, ok := operators[in.OperatorId]; ok {
		data.Operator = operator.Nickname
	}

	return data
}

// NewRolePermissions 创建角色权限列表
func NewRolePermissions(permissions []string, resources []*pb.ResourceInfo) []*pb.RoleResource {
	rolePermissions := utils.Map(resources, func(in *pb.ResourceInfo) *pb.RoleResource {
		item := utils.CopyPtr[pb.RoleResource](in)

		item.HasPermission = slices.Contains(permissions, newPermissionObject(in))

		return item
	})

	for _, item := range rolePermissions {
		resource, ok := utils.FindFirst(resources, func(in *pb.ResourceInfo) bool {
			return in.Id == item.Id
		})

		if !ok || len(resource.Children) == 0 {
			continue
		}

		item.Children = NewRolePermissions(permissions, resource.Children)
	}

	return rolePermissions
}

// NewRolePermissionPolicies 创建角色权限策略
func NewRolePermissionPolicies(roleId int64, resources []*pb.ResourceInfo) []*toolboxauthz.Policy {
	newData := make([]*toolboxauthz.Policy, 0, len(resources))

	for _, item := range resources {
		newData = append(newData, toolboxauthz.NewPolicy(NewRoleSubject(roleId), authz.Domain, newPermissionObject(item), authz.Act))

		if len(item.Apis) == 0 {
			continue
		}

		apiData := utils.Map(item.Apis, func(api string) *toolboxauthz.Policy {
			return toolboxauthz.NewPolicy(
				NewRoleSubject(roleId),
				authz.Domain,
				newPermissionObject(&pb.ResourceInfo{Type: string(valueobject.APIResourceType), Path: api}),
				authz.Act,
			)
		})

		newData = append(newData, apiData...)
	}

	return utils.Distinct(newData, func(in *toolboxauthz.Policy) string { return in.Obj })
}
