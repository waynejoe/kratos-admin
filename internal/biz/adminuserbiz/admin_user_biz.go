package adminuserbiz

import (
	"context"
	"fmt"
	"slices"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/biz/adminuserbiz/factory"
	"kratos-admin/internal/biz/permissionbiz"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type AdminUserUsecase struct {
	adminUserRepo         *adminrepo.AdminUserRepo
	adminRoleRepo         *adminrepo.AdminRoleRepo
	adminGroupRepo        *adminrepo.AdminGroupRepo
	userPermissionUsecase *permissionbiz.UserPermissionUsecase
}

func NewAdminUserUsecase(
	adminUserRepo *adminrepo.AdminUserRepo,
	adminRoleRepo *adminrepo.AdminRoleRepo,
	adminGroupRepo *adminrepo.AdminGroupRepo,
	userPermissionUsecase *permissionbiz.UserPermissionUsecase,
) *AdminUserUsecase {
	return &AdminUserUsecase{
		adminUserRepo:         adminUserRepo,
		adminRoleRepo:         adminRoleRepo,
		adminGroupRepo:        adminGroupRepo,
		userPermissionUsecase: userPermissionUsecase,
	}
}

func (uc *AdminUserUsecase) GetUserInfo(ctx context.Context, _ *pb.GetUserInfoRequest) (*pb.GetUserInfoReply, error) {
	userId := helpx.GetUserId(ctx)

	user, err := uc.adminUserRepo.Get(ctx, userId)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	reply := utils.CopyPtr[pb.GetUserInfoReply](user)
	reply.Permissions, reply.ClosePermission, err = uc.userPermissionUsecase.GetUserPermission(ctx, userId)

	return reply, err
}

func (uc *AdminUserUsecase) ListAdminUser(ctx context.Context, req *pb.ListAdminUserRequest) (*pb.ListAdminUserReply, error) {
	userIds := []int64{}

	if req.GroupId > 0 {
		userIds = uc.adminGroupRepo.GetUsersForGroup(fmt.Sprintf("%s%d", authz.GroupPermissionPrefix, req.GroupId))
		if len(userIds) == 0 {
			return &pb.ListAdminUserReply{}, nil
		}
	}

	if req.RoleId > 0 {
		roleUserIds := uc.adminRoleRepo.GetUsersForRole(fmt.Sprintf("%s%d", authz.RolePermissionPrefix, req.RoleId))
		userIds = utils.Filter(roleUserIds, func(id int64) bool { return req.GroupId == 0 || slices.Contains(userIds, id) })
		if len(userIds) == 0 {
			return &pb.ListAdminUserReply{}, nil
		}
	}

	users, total, err := uc.adminUserRepo.QueryUser(ctx, req, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	userIds = utils.Map(users, func(item *adminmodel.AdminUser) int64 { return item.Id })

	userRoles, err := uc.adminRoleRepo.GetUserRoles(ctx, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	userGroups, err := uc.adminGroupRepo.GetUserGroups(ctx, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	operators, err := uc.adminUserRepo.GetsMap(ctx, utils.Map(users, func(in *adminmodel.AdminUser) int64 { return in.OperatorId }))
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.ListAdminUserReply{
		List: utils.Map(users, func(item *adminmodel.AdminUser) *pb.AdminUserInfo {
			return factory.NewAdminUserInfo(item, userRoles, userGroups, operators)
		}),
		Page: adminmodel.NewPageInfo(req.PageIndex, req.PageSize, total, len(users) < int(req.PageSize)),
	}, nil
}

func (uc *AdminUserUsecase) SaveAdminUser(ctx context.Context, req *pb.SaveAdminUserRequest) (*pb.SaveAdminUserReply, error) {
	if req.Id == 0 {
		return nil, pb.ErrorUserParamsError("暂不支持新增用户")
	}

	user, err := uc.adminUserRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	user.Status = req.Status
	user.OperatorId = helpx.GetUserId(ctx)

	if err = uc.adminUserRepo.Save(ctx, user); err != nil {
		return nil, errorx.WithStack(err)
	}

	if req.Status != adminmodel.UserStatusNormal {
		_ = uc.adminUserRepo.GetCache().Del(ctx, fmt.Sprintf(AuthAccessKey, user.Id))
	}

	if err := uc.userPermissionUsecase.GrantUserPermission(ctx, user.Id, req.RoleIds, req.GroupIds); err != nil {
		return nil, errorx.WithStack(err)
	}

	return &pb.SaveAdminUserReply{Id: user.Id}, nil
}

func (uc *AdminUserUsecase) ListOptimizer(ctx context.Context, _ *pb.ListOptimizerRequest) (*pb.ListOptimizerReply, error) {
	return uc.userPermissionUsecase.GetOptimizers(ctx)
}
