package factory

import (
	"context"
	"fmt"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"

	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/utils"
)

// NewGroupSubject 创建团队权限key
func NewGroupSubject(groupId int64) string {
	return fmt.Sprintf("%s%d", authz.GroupPermissionPrefix, groupId)
}

// NewGroupModel 创建团队模型
func NewGroupModel(ctx context.Context, in *pb.SaveGroupRequest) *adminmodel.AdminGroup {
	data := utils.CopyPtr[adminmodel.AdminGroup](in)

	data.Status = adminmodel.StatusOn
	data.OperatorId = helpx.GetUserId(ctx)
	data.UserPrivileges = "[]"

	return data
}

// NewGroupInfo 创建团队信息
func NewGroupInfo(
	in *adminmodel.AdminGroup,
	userGroups map[int64][]int64,
	roles []*adminmodel.AdminRole,
	operators map[int64]*adminmodel.AdminUser,
) *pb.GroupInfo {
	data := utils.CopyPtr[pb.GroupInfo](in)

	data.CreateTime = in.CreateTime.UnixMilli()
	data.UserCount = int32(len(userGroups[in.Id]))

	groupRoles := utils.Filter(roles, func(item *adminmodel.AdminRole) bool { return item.GroupId == in.Id })

	data.Roles = utils.Map(groupRoles, func(item *adminmodel.AdminRole) *pb.GroupInfo_Role {
		return utils.CopyPtr[pb.GroupInfo_Role](item)
	})

	if operator, ok := operators[in.OperatorId]; ok {
		data.Operator = operator.Nickname
	}

	return data
}

// NewGroupMemberInfo 创建团队成员信息
func NewGroupMemberInfo(
	in *adminmodel.AdminUser,
	userPrivileges []*valueobject.GroupUserPrivilege,
	userRoles map[int64][]*adminmodel.AdminRole,
) *pb.GroupMemberInfo {
	data := utils.CopyPtr[pb.GroupMemberInfo](in)
	data.UserId = in.Id

	privilege, ok := utils.FindFirst(userPrivileges, func(item *valueobject.GroupUserPrivilege) bool {
		return item.UserId == in.Id
	})
	if ok {
		data.PrivilegeLevel = int32(privilege.PrivilegeLevel)
	} else {
		data.PrivilegeLevel = int32(valueobject.MemberPrivilegeLevel)
	}

	for _, role := range userRoles[in.Id] {
		data.Roles = append(data.Roles, role.Name)
	}

	return data
}
