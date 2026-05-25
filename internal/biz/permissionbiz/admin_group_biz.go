package permissionbiz

import (
	"context"
	"slices"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"

	factory "kratos-admin/internal/biz/permissionbiz/factory"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type AdminGroupUsecase struct {
	adminGroupRepo *adminrepo.AdminGroupRepo
	resourceRepo   *adminrepo.ResourceRepo
	adminUserRepo  *adminrepo.AdminUserRepo
	adminRoleRepo  *adminrepo.AdminRoleRepo
	config         *conf.Permission
}

func NewAdminGroupUsecase(
	adminGroupRepo *adminrepo.AdminGroupRepo,
	resourceRepo *adminrepo.ResourceRepo,
	adminUserRepo *adminrepo.AdminUserRepo,
	adminRoleRepo *adminrepo.AdminRoleRepo,
	config *conf.Permission,
) *AdminGroupUsecase {
	return &AdminGroupUsecase{
		adminGroupRepo: adminGroupRepo,
		resourceRepo:   resourceRepo,
		adminUserRepo:  adminUserRepo,
		adminRoleRepo:  adminRoleRepo,
		config:         config,
	}
}

func (uc *AdminGroupUsecase) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	groups, total, err := uc.adminGroupRepo.QueryGroup(ctx, req, 0)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	var (
		groupIds   = utils.Map(groups, func(in *adminmodel.AdminGroup) int64 { return in.Id })
		userGroups = make(map[int64][]int64)
	)

	for _, groupId := range groupIds {
		userGroups[groupId] = uc.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(groupId))
	}

	roles, err := uc.adminRoleRepo.GetRoleByGroupId(ctx, groupIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	operators, err := uc.adminUserRepo.GetsMap(ctx, utils.Map(groups, func(in *adminmodel.AdminGroup) int64 { return in.OperatorId }))
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.ListGroupReply{
		List: utils.Map(groups, func(in *adminmodel.AdminGroup) *pb.GroupInfo {
			return factory.NewGroupInfo(in, userGroups, roles, operators)
		}),
		Page: adminmodel.NewPageInfo(req.PageIndex, req.PageSize, total, len(groups) < int(req.PageSize)),
	}, nil
}

func (uc *AdminGroupUsecase) SaveGroup(ctx context.Context, req *pb.SaveGroupRequest) (*pb.SaveGroupReply, error) {
	newGroup := factory.NewGroupModel(ctx, req)

	if req.Id > 0 {
		group, err := uc.adminGroupRepo.Get(ctx, req.Id)
		if err != nil {
			return nil, errorx.WithStack(err)
		}

		newGroup.UserPrivileges = group.UserPrivileges
		newGroup.OperatorId = group.OperatorId
	}

	if err := uc.adminGroupRepo.Save(ctx, newGroup); err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.SaveGroupReply{Id: newGroup.Id}, nil
}

func (uc *AdminGroupUsecase) DeleteGroup(ctx context.Context, id int64) error {
	group, err := uc.adminGroupRepo.Get(ctx, id)
	if err != nil {
		return errorx.WithStack(err)
	}

	if valueobject.GroupDataIsolation(group.DataIsolation).IsOpen() {
		return pb.ErrorPermissionInvalidParams("角色所属团队开启数据隔离，不允许删除")
	}

	userGroups := uc.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(group.Id))
	if len(userGroups) > 0 {
		return errorx.WithStack(errorx.New("团队成员不为空，不允许删除"))
	}

	groupRoles, err := uc.adminRoleRepo.GetRoleByGroupId(ctx, group.Id)
	if err != nil {
		return errorx.WithStack(err)
	}

	if len(groupRoles) > 0 {
		return pb.ErrorPermissionInvalidParams("团队下存在角色，请先清理角色")
	}

	return uc.adminGroupRepo.DeleteGroup(ctx, group, factory.NewGroupSubject(group.Id))
}

func (uc *AdminGroupUsecase) ListGroupMember(ctx context.Context, req *pb.ListGroupMemberRequest) (*pb.ListGroupMemberReply, error) {
	adminUsers, err := uc.adminUserRepo.FuzzyQueryUsers(ctx, req.FuzzyQuery)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	if req.FuzzyQuery != "" && len(adminUsers) == 0 {
		return &pb.ListGroupMemberReply{}, nil
	}

	group, err := uc.adminGroupRepo.Get(ctx, req.GroupId)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	userPrivileges, err := utils.UnmarshalList[*valueobject.GroupUserPrivilege](group.UserPrivileges)
	if err != nil {
		return nil, err
	}

	groupUsers := uc.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(group.Id))
	if len(groupUsers) == 0 {
		return &pb.ListGroupMemberReply{}, nil
	}

	if len(adminUsers) != 0 {
		groupUsers = utils.Filter(groupUsers, func(id int64) bool { return slices.Contains(adminUsers, id) })
	}

	slices.SortFunc(groupUsers, func(a, b int64) int { return int(b - a) })

	userIds := utils.CutPage(groupUsers, req.PageIndex, req.PageSize)

	users, err := uc.adminUserRepo.Gets(ctx, userIds)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	userRoles, err := uc.adminRoleRepo.GetUserRoles(ctx, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.ListGroupMemberReply{
		List: utils.Map(users, func(in *adminmodel.AdminUser) *pb.GroupMemberInfo {
			return factory.NewGroupMemberInfo(in, userPrivileges, userRoles)
		}),
		Page: adminmodel.NewPageInfo(req.PageIndex, req.PageSize, int64(len(groupUsers)), len(userIds) < int(req.PageSize)),
	}, nil
}

func (uc *AdminGroupUsecase) SaveGroupMember(ctx context.Context, req *pb.SaveGroupMemberRequest) error {
	group, err := uc.adminGroupRepo.Get(ctx, req.GroupId)
	if err != nil {
		return errorx.WithStack(err)
	}

	groupUsers := uc.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(group.Id))

	if !slices.Contains(groupUsers, req.UserId) {
		return errorx.WithStack(errorx.New("用户不是团队成员"))
	}

	data, err := utils.UnmarshalList[*valueobject.GroupUserPrivilege](group.UserPrivileges)
	if err != nil {
		return err
	}

	userPrivileges := utils.Filter(data, func(in *valueobject.GroupUserPrivilege) bool { return slices.Contains(groupUsers, in.UserId) })

	user, ok := utils.FindFirst(userPrivileges, func(in *valueobject.GroupUserPrivilege) bool { return in.UserId == req.UserId })
	if ok {
		user.PrivilegeLevel = valueobject.PrivilegeLevel(req.PrivilegeLevel)
	} else {
		data = append(userPrivileges, valueobject.NewGroupUserPrivilege(req.UserId, req.PrivilegeLevel))
	}

	group.UserPrivileges, err = utils.MarshalList(data)
	if err != nil {
		return err
	}

	return uc.adminGroupRepo.Save(ctx, group)
}
