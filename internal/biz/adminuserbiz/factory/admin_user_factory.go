package factory

import (
	"kratos-admin/pkg/toolbox/utils"

	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

func NewAdminUserInfo(
	in *adminmodel.AdminUser,
	userRoles map[int64][]*adminmodel.AdminRole,
	userGroups map[int64][]*adminmodel.AdminGroup,
	operators map[int64]*adminmodel.AdminUser,
) *pb.AdminUserInfo {
	data := utils.CopyPtr[pb.AdminUserInfo](in)
	data.CreateTime = in.CreateTime.UnixMilli()
	data.Roles = utils.Map(userRoles[in.Id], func(item *adminmodel.AdminRole) *pb.AdminUserInfo_UserRole {
		return &pb.AdminUserInfo_UserRole{Id: item.Id, Name: item.Name}
	})
	data.Groups = utils.Map(userGroups[in.Id], func(item *adminmodel.AdminGroup) *pb.AdminUserInfo_UserGroup {
		return &pb.AdminUserInfo_UserGroup{Id: item.Id, Name: item.Name}
	})
	if operator, ok := operators[in.OperatorId]; ok {
		data.Operator = operator.Nickname
	}
	return data
}
