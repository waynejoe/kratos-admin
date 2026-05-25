package permissionbiz

import (
	"context"
	"errors"
	"slices"

	toolboxauthz "kratos-admin/pkg/toolbox/authz"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"
	"gorm.io/gorm"

	"kratos-admin/internal/authz"
	factory "kratos-admin/internal/biz/permissionbiz/factory"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type AdminRoleUsecase struct {
	adminRoleRepo  *adminrepo.AdminRoleRepo
	adminGroupRepo *adminrepo.AdminGroupRepo
	resourceRepo   *adminrepo.ResourceRepo
	adminUserRepo  *adminrepo.AdminUserRepo
	config         *conf.Permission
}

func NewAdminRoleUsecase(
	adminRoleRepo *adminrepo.AdminRoleRepo,
	adminGroupRepo *adminrepo.AdminGroupRepo,
	resourceRepo *adminrepo.ResourceRepo,
	adminUserRepo *adminrepo.AdminUserRepo,
	config *conf.Permission,
) *AdminRoleUsecase {
	return &AdminRoleUsecase{
		adminRoleRepo:  adminRoleRepo,
		adminGroupRepo: adminGroupRepo,
		resourceRepo:   resourceRepo,
		adminUserRepo:  adminUserRepo,
		config:         config,
	}
}

func (s *AdminRoleUsecase) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	data, total, err := s.adminRoleRepo.QueryRole(ctx, req)
	if err != nil {
		return nil, err
	}

	groupIds := utils.Map(data, func(in *adminmodel.AdminRole) int64 { return in.GroupId })

	groups, err := s.adminGroupRepo.GetsMap(ctx, groupIds)
	if err != nil {
		return nil, err
	}

	var (
		roleIds   = utils.Map(data, func(in *adminmodel.AdminRole) int64 { return in.Id })
		roleUsers = make(map[int64][]int64)
	)

	for _, roleId := range roleIds {
		roleUsers[roleId] = s.adminRoleRepo.GetUsersForRole(factory.NewRoleSubject(roleId))
	}

	operators, err := s.adminUserRepo.GetsMap(ctx, utils.Map(data, func(in *adminmodel.AdminRole) int64 { return in.OperatorId }))
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.ListRoleReply{
		List: utils.Map(data, func(in *adminmodel.AdminRole) *pb.RoleInfo {
			return factory.NewRoleInfo(in, groups, roleUsers, operators)
		}),
		Page: adminmodel.NewPageInfo(req.PageIndex, req.PageSize, total, len(data) < int(req.PageSize)),
	}, nil
}

func (s *AdminRoleUsecase) SaveRole(ctx context.Context, req *pb.SaveRoleRequest) (*pb.SaveRoleReply, error) {
	var (
		oldGroupId = req.GroupId
		newRole    = factory.NewRoleModel(ctx, req)
	)

	if req.Id > 0 {
		oldRole, err := s.adminRoleRepo.Get(ctx, req.Id)
		if err != nil {
			return nil, errorx.WithStack(err)
		}

		oldGroupId = oldRole.GroupId
		newRole.OperatorId = oldRole.OperatorId
	}

	if err := s.adminRoleRepo.Save(ctx, newRole); err != nil {
		return nil, errorx.WithStack(err)
	}

	if newRole.GroupId != oldGroupId {
		roleUsers := s.adminRoleRepo.GetUsersForRole(factory.NewRoleSubject(newRole.Id))

		if err := s.adminGroupRepo.UpdateUsersGroup(
			factory.NewGroupSubject(newRole.GroupId),
			factory.NewGroupSubject(oldGroupId),
			roleUsers,
		); err != nil {
			return nil, errorx.WithStack(err)
		}
	}

	return &pb.SaveRoleReply{Id: newRole.Id}, nil
}

func (s *AdminRoleUsecase) DeleteRole(ctx context.Context, id int64) error {
	role, err := s.adminRoleRepo.Get(ctx, id)
	if err != nil {
		return errorx.WithStack(err)
	}

	group, err := s.adminGroupRepo.Get(ctx, role.GroupId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errorx.WithStack(err)
	}

	if group != nil && valueobject.GroupDataIsolation(group.DataIsolation).IsOpen() {
		return pb.ErrorPermissionInvalidParams("角色所属团队开启数据隔离，不允许删除")
	}

	roleUsers := s.adminRoleRepo.GetUsersForRole(factory.NewRoleSubject(id))
	if len(roleUsers) > 0 {
		return pb.ErrorPermissionInvalidParams("角色已绑定用户，不允许删除")
	}

	return s.adminRoleRepo.DeleteRole(ctx, role, factory.NewRoleSubject(role.Id))
}

func (s *AdminRoleUsecase) ListRolePermission(ctx context.Context, req *pb.ListRolePermissionRequest) (*pb.ListRolePermissionReply, error) {
	permissions := s.adminRoleRepo.GetRolePermissions(factory.NewRoleSubject(req.RoleId))

	raw, err := s.resourceRepo.GetAllResource(ctx)
	if err != nil {
		return nil, err
	}

	resources := factory.BuildResourceInfos(ctx, 0, raw)

	return &pb.ListRolePermissionReply{
		List: factory.NewRolePermissions(permissions, resources),
	}, nil
}

func (s *AdminRoleUsecase) SaveRolePermission(ctx context.Context, req *pb.SaveRolePermissionRequest) error {
	role, err := s.adminRoleRepo.Get(ctx, req.RoleId)
	if err != nil {
		return errorx.WithStack(err)
	}

	resources, err := s.resourceRepo.Gets(ctx, req.ResourceIds)
	if err != nil {
		return errorx.WithStack(err)
	}

	oldPermissions := s.adminRoleRepo.GetRolePermissions(factory.NewRoleSubject(req.RoleId))
	oldData := utils.Map(oldPermissions, func(in string) *toolboxauthz.Policy {
		return toolboxauthz.NewPolicy(factory.NewRoleSubject(role.Id), authz.Domain, in, authz.Act)
	})

	resourceData, err := utils.MapWithErr(resources, factory.NewResourceFromModel)
	if err != nil {
		return errorx.WithStack(err)
	}

	newData := factory.NewRolePermissionPolicies(role.Id, resourceData)

	return s.adminRoleRepo.UpdateRolePermissions(ctx, newData, oldData)
}

func (s *AdminRoleUsecase) ListRoleMember(ctx context.Context, req *pb.ListRoleMemberRequest) (*pb.ListRoleMemberReply, error) {
	adminUsers, err := s.adminUserRepo.FuzzyQueryUsers(ctx, req.FuzzyQuery)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	if req.FuzzyQuery != "" && len(adminUsers) == 0 {
		return &pb.ListRoleMemberReply{}, nil
	}

	roleUsers := s.adminRoleRepo.GetUsersForRole(factory.NewRoleSubject(req.RoleId))
	if len(roleUsers) == 0 {
		return &pb.ListRoleMemberReply{}, nil
	}

	if len(adminUsers) != 0 {
		roleUsers = utils.Filter(roleUsers, func(id int64) bool { return slices.Contains(adminUsers, id) })
	}

	slices.SortFunc(roleUsers, func(a, b int64) int { return int(b - a) })

	userIds := utils.CutPage(roleUsers, req.PageIndex, req.PageSize)

	users, err := s.adminUserRepo.Gets(ctx, userIds)
	if err != nil {
		return nil, err
	}

	return &pb.ListRoleMemberReply{
		List: utils.Map(users, func(item *adminmodel.AdminUser) *pb.RoleMemberInfo {
			user := utils.CopyPtr[pb.RoleMemberInfo](item)
			user.UserId = item.Id
			return user
		}),
		Page: adminmodel.NewPageInfo(req.PageIndex, req.PageSize, int64(len(roleUsers)), len(users) < int(req.PageSize)),
	}, nil
}
