package permissionbiz

import (
	"context"
	"slices"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/logx"
	"kratos-admin/pkg/toolbox/utils"

	factory "kratos-admin/internal/biz/permissionbiz/factory"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	"kratos-admin/internal/authz"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type UserPermissionUsecase struct {
	authz          *authz.Store
	adminUserRepo  *adminrepo.AdminUserRepo
	adminRoleRepo  *adminrepo.AdminRoleRepo
	adminGroupRepo *adminrepo.AdminGroupRepo
	resourceRepo   *adminrepo.ResourceRepo
	config         *conf.Permission
}

func NewUserPermissionUsecase(
	authzStore *authz.Store,
	adminUserRepo *adminrepo.AdminUserRepo,
	adminRoleRepo *adminrepo.AdminRoleRepo,
	adminGroupRepo *adminrepo.AdminGroupRepo,
	resourceRepo *adminrepo.ResourceRepo,
	config *conf.Permission,
) *UserPermissionUsecase {
	return &UserPermissionUsecase{
		authz:          authzStore,
		adminUserRepo:  adminUserRepo,
		adminRoleRepo:  adminRoleRepo,
		adminGroupRepo: adminGroupRepo,
		resourceRepo:   resourceRepo,
		config:         config,
	}
}

func (uc *UserPermissionUsecase) GetUserDataPermission(ctx context.Context, userId int64) (*adminmodel.UserDataPermission, error) {
	permission := &adminmodel.UserDataPermission{UserIds: make([]int64, 0)}

	if userId == 0 {
		return permission, nil
	}

	data, err := uc.adminGroupRepo.GetUserGroups(ctx, userId)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	userGroups, ok := data[userId]
	if !ok || len(userGroups) == 0 {
		return permission, nil
	}

	if slices.ContainsFunc(userGroups, func(group *adminmodel.AdminGroup) bool {
		return valueobject.GroupDataIsolation(group.DataIsolation).IsClose()
	}) {
		permission.HasDataPermission = true
		return permission, nil
	}

	userDataPermission := []int64{userId}

	for _, group := range userGroups {
		privileges, err := utils.UnmarshalList[*valueobject.GroupUserPrivilege](group.UserPrivileges)
		if err != nil {
			logx.Errorf(ctx, "NewGroupUserPrivilegesFromMode err: %+v", err)
			continue
		}

		isLeader := slices.ContainsFunc(privileges, func(privilege *valueobject.GroupUserPrivilege) bool {
			return privilege.UserId == userId && privilege.PrivilegeLevel.IsLeader()
		})

		if isLeader {
			groupUsers := uc.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(group.Id))
			userDataPermission = append(userDataPermission, groupUsers...)
		}
	}

	permission.HasDataPermission = true
	permission.UserIds = utils.Distinct(userDataPermission, func(item int64) int64 { return item })

	return permission, nil
}

func (s *UserPermissionUsecase) GetUserPermission(ctx context.Context, userId int64) (*pb.UserPermission, bool, error) {
	permissions := s.authz.GetUserPermissions(ctx, userId)

	raw, err := s.resourceRepo.GetAllResource(ctx)
	if err != nil {
		return nil, s.config.ClosePermission, errorx.WithStack(err)
	}

	resources := factory.BuildResourceInfos(ctx, 0, raw)

	buttons, pages := factory.NewUserPermission(permissions, resources)

	return &pb.UserPermission{Buttons: buttons, Pages: pages}, s.config.ClosePermission, nil
}

func (s *UserPermissionUsecase) GrantUserPermission(ctx context.Context, userId int64, roleIds, groupIds []int64) error {
	roleIds = utils.Distinct(roleIds, func(item int64) int64 { return item })
	groupIds = utils.Distinct(groupIds, func(item int64) int64 { return item })

	permissions := utils.Map(roleIds, func(roleId int64) string { return factory.NewRoleSubject(roleId) })
	permissions = append(permissions, utils.Map(groupIds, func(groupId int64) string { return factory.NewGroupSubject(groupId) })...)

	return s.authz.UpdateUserPermissions(ctx, userId, permissions)
}

func (s *UserPermissionUsecase) GetOptimizers(ctx context.Context) (*pb.ListOptimizerReply, error) {
	permission, err := s.GetUserDataPermission(ctx, helpx.GetUserId(ctx))
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	if !permission.HasDataPermission {
		return &pb.ListOptimizerReply{}, nil
	}

	var optimizers []*adminmodel.AdminUser

	if len(permission.UserIds) == 0 {
		groups, _, err := s.adminGroupRepo.QueryGroup(ctx, &pb.ListGroupRequest{}, int32(valueobject.OpenDataIsolation))
		if err != nil {
			return nil, errorx.WithStack(err)
		}

		for _, group := range groups {
			permission.UserIds = append(permission.UserIds, s.adminGroupRepo.GetUsersForGroup(factory.NewGroupSubject(group.Id))...)
		}
	}

	optimizers, err = s.adminUserRepo.Gets(ctx, permission.UserIds)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	slices.SortFunc(optimizers, func(a, b *adminmodel.AdminUser) int { return int(b.Id - a.Id) })

	return &pb.ListOptimizerReply{
		Users: utils.Map(optimizers, func(item *adminmodel.AdminUser) *pb.Optimizer { return utils.CopyPtr[pb.Optimizer](item) }),
	}, nil
}
