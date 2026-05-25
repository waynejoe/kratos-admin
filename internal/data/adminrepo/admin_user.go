package adminrepo

import (
	"context"

	"kratos-admin/pkg/toolbox/datax"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/data"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

// AdminUserRepo 管理员用户仓库
type AdminUserRepo struct {
	datax.CacheRepo[adminmodel.AdminUser]
	data *data.Data
}

// NewAdminUserRepo 创建管理员用户仓库
func NewAdminUserRepo(data *data.Data) *AdminUserRepo {
	return &AdminUserRepo{
		CacheRepo: datax.NewCacheRepo[adminmodel.AdminUser](data.G.AdminDB, data.G.LocalCache),
		data:      data,
	}
}

func (r *AdminUserRepo) GetAdminUserByPhone(ctx context.Context, phone string) (*adminmodel.AdminUser, error) {
	return r.GetBy(ctx, "`phone` = ?", phone)
}

func (r *AdminUserRepo) GetAdminUserByAccount(ctx context.Context, account string) (*adminmodel.AdminUser, error) {
	return r.GetBy(ctx, "`username` =? OR `phone` =?", account, account)
}

func (r *AdminUserRepo) GetCache() datax.Cache {
	return r.data.G.LocalCache
}

func (r *AdminUserRepo) FuzzyQueryUsers(ctx context.Context, fuzzyQuery string) ([]int64, error) {
	if fuzzyQuery == "" {
		return make([]int64, 0), nil
	}

	var (
		q    = r.DB(ctx)
		data []*adminmodel.AdminUser
	)

	if adminmodel.FuzzyQueryIsId(fuzzyQuery) {
		q = q.Where("`id` = ?", fuzzyQuery)
	} else {
		q = q.Where("`nickname` LIKE ?", "%"+fuzzyQuery+"%")
	}

	if err := q.Find(&data).Error; err != nil {
		return nil, errorx.WithStack(err)
	}

	return datax.GetIds(data), nil
}

func (r *AdminUserRepo) QueryUser(ctx context.Context, req *pb.ListAdminUserRequest, userIds ...int64) ([]*adminmodel.AdminUser, int64, error) {
	var (
		query = r.DB(ctx)
		data  []*adminmodel.AdminUser
		total int64
		page  = utils.NewPage(req.PageIndex, req.PageSize)
	)

	if len(userIds) > 0 {
		query = query.Where("`id` IN ?", utils.Distinct(userIds, func(item int64) int64 { return item }))
	}

	if req.Nickname != "" {
		query = query.Where("`nickname` LIKE ?", "%"+req.Nickname+"%")
	}

	if req.Status != 0 {
		query = query.Where("`status` = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errorx.WithStack(err)
	}

	err := query.Order("`id` DESC").Offset(page.GetOffset()).Limit(page.GetLimit()).Find(&data).Error

	return data, total, errorx.WithStack(err)
}
