package adminrepo

import (
	"context"
	"slices"

	"kratos-admin/pkg/toolbox/datax"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/data"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

// AdminGroupRepo 管理员组仓库
type AdminGroupRepo struct {
	datax.CacheRepo[adminmodel.AdminGroup]
	authz *authz.Store
}

// NewAdminGroupRepo 新建管理员组仓库
func NewAdminGroupRepo(data *data.Data, authzStore *authz.Store) *AdminGroupRepo {
	visible := func(e *adminmodel.AdminGroup) bool { return !e.Deleted }

	return &AdminGroupRepo{
		CacheRepo: datax.NewCacheRepo(data.G.AdminDB, data.G.LocalCache, datax.WithVisible(visible)),
		authz:     authzStore,
	}
}

func (r *AdminGroupRepo) QueryGroup(ctx context.Context, req *pb.ListGroupRequest, dataIsolation int32) ([]*adminmodel.AdminGroup, int64, error) {
	var (
		query = r.DB(ctx).Where("`deleted` = ?", adminmodel.NotDel)
		data  []*adminmodel.AdminGroup
		total int64
		page  = utils.NewPage(req.PageIndex, req.PageSize)
	)

	if len(req.Ids) > 0 {
		query = query.Where("`id` IN ?", req.Ids)
	}

	if req.Name != "" {
		query = query.Where("`name` LIKE ?", "%"+req.Name+"%")
	}

	if dataIsolation > 0 {
		query = query.Where("`data_isolation` = ?", dataIsolation)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errorx.WithStack(err)
	}

	err := query.Order("`id` DESC").Offset(page.GetOffset()).Limit(page.GetLimit()).Find(&data).Error

	return data, total, errorx.WithStack(err)
}

func (r *AdminGroupRepo) DeleteGroup(ctx context.Context, group *adminmodel.AdminGroup, groupSub string) error {
	if err := r.authz.DeleteGroup(ctx, groupSub); err != nil {
		return err
	}

	group.Deleted = true
	group.Status = adminmodel.StatusOff

	return r.Save(ctx, group)
}

func (r *AdminGroupRepo) GetUsersForGroup(groupSub string) []int64 {
	return r.authz.GetUsersForGroup(groupSub)
}

func (r *AdminGroupRepo) GetUserGroups(ctx context.Context, userIds ...int64) (map[int64][]*adminmodel.AdminGroup, error) {
	groupIdMap, err := r.authz.GetUserGroups(ctx, userIds...)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	var groupIds []int64
	for _, ids := range groupIdMap {
		groupIds = append(groupIds, ids...)
	}

	groups, err := r.RawGetsBy(ctx, "`id` IN ? AND `deleted` = ?", groupIds, adminmodel.NotDel)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	data := make(map[int64][]*adminmodel.AdminGroup)
	for userId, ids := range groupIdMap {
		data[userId] = utils.Filter(groups, func(item *adminmodel.AdminGroup) bool {
			return slices.ContainsFunc(ids, func(id int64) bool { return id == item.Id })
		})
	}

	return data, nil
}

func (r *AdminGroupRepo) UpdateUsersGroup(newGroupSub, oldGroupSub string, userIds []int64) error {
	return r.authz.UpdateUsersGroup(context.Background(), newGroupSub, oldGroupSub, userIds)
}
