package adminrepo

import (
	"context"
	"slices"

	toolboxauthz "kratos-admin/pkg/toolbox/authz"
	"kratos-admin/pkg/toolbox/datax"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/data"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

// AdminRoleRepo 管理员角色仓库
type AdminRoleRepo struct {
	datax.Repo[adminmodel.AdminRole]
	authz *authz.Store
}

// NewAdminRoleRepo 新建管理员角色仓库
func NewAdminRoleRepo(data *data.Data, authzStore *authz.Store) *AdminRoleRepo {
	visible := func(e *adminmodel.AdminRole) bool { return !e.Deleted }

	return &AdminRoleRepo{
		Repo:  datax.NewRepo(data.G.AdminDB, datax.WithVisible(visible)),
		authz: authzStore,
	}
}

func (r *AdminRoleRepo) QueryRole(ctx context.Context, req *pb.ListRoleRequest) ([]*adminmodel.AdminRole, int64, error) {
	var (
		query = r.DB(ctx).Where("`deleted` = ?", adminmodel.NotDel)
		data  []*adminmodel.AdminRole
		total int64
		page  = utils.NewPage(req.PageIndex, req.PageSize)
	)

	if req.Name != "" {
		query = query.Where("`name` LIKE ?", "%"+req.Name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errorx.WithStack(err)
	}

	err := query.Order("`id` DESC").Offset(page.GetOffset()).Limit(page.GetLimit()).Find(&data).Error

	return data, total, errorx.WithStack(err)
}

func (r *AdminRoleRepo) GetRoleByGroupId(ctx context.Context, groupIds ...int64) ([]*adminmodel.AdminRole, error) {
	var data []*adminmodel.AdminRole

	err := r.DB(ctx).Where("`group_id` IN ? AND `deleted` = ?", groupIds, adminmodel.NotDel).Find(&data).Error

	return data, errorx.WithStack(err)
}

func (r *AdminRoleRepo) UpdateRolePermissions(ctx context.Context, newData, deleteData []*toolboxauthz.Policy) error {
	return r.authz.UpdateRolePermissions(ctx, newData, deleteData)
}

func (r *AdminRoleRepo) DeleteRole(ctx context.Context, role *adminmodel.AdminRole, roleSub string) error {
	if err := r.authz.DeleteRole(ctx, roleSub); err != nil {
		return err
	}

	role.Deleted = true
	role.Status = adminmodel.StatusOff

	return r.Save(ctx, role)
}

func (r *AdminRoleRepo) GetRolePermissions(roleSub string) []string {
	return r.authz.GetRolePermissions(roleSub)
}

func (r *AdminRoleRepo) GetUsersForRole(roleSub string) []int64 {
	return r.authz.GetUsersForRole(roleSub)
}

func (r *AdminRoleRepo) GetUserRoles(ctx context.Context, userIds ...int64) (map[int64][]*adminmodel.AdminRole, error) {
	roleIdMap, err := r.authz.GetUserRoles(ctx, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	var roleIds []int64
	for _, ids := range roleIdMap {
		roleIds = append(roleIds, ids...)
	}

	roles, err := r.Gets(ctx, roleIds)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	data := make(map[int64][]*adminmodel.AdminRole)
	for userId, ids := range roleIdMap {
		data[userId] = utils.Filter(roles, func(item *adminmodel.AdminRole) bool {
			return slices.ContainsFunc(ids, func(id int64) bool { return id == item.Id })
		})
	}

	return data, nil
}
