package adminrepo

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"kratos-admin/pkg/toolbox/datax"
	"kratos-admin/pkg/toolbox/errorx"

	"kratos-admin/internal/data"
	"kratos-admin/pkg/model/adminmodel"
)

// ResourceRepo 资源仓库
type ResourceRepo struct {
	datax.Repo[adminmodel.Resource]
	cache datax.Cache
}

// NewResourceRepo 新建资源仓库
func NewResourceRepo(data *data.Data) *ResourceRepo {
	return &ResourceRepo{
		Repo:  datax.NewRepo[adminmodel.Resource](data.G.AdminDB),
		cache: data.G.LocalCache,
	}
}

const (
	resourceCacheKey = "mome-admin:permission:resource:all"
	resourceCacheTTL = 3600 * 24
)

func (r *ResourceRepo) GetResourceByPath(ctx context.Context, path string) (*adminmodel.Resource, error) {
	var data adminmodel.Resource

	err := r.DB(ctx).Where("`path` = ?", path).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &data, errorx.WithStack(err)
}

func (r *ResourceRepo) GetResourcesByParentId(ctx context.Context, parentId int64) ([]*adminmodel.Resource, error) {
	var data []*adminmodel.Resource

	err := r.DB(ctx).Where("`parent_id` = ?", parentId).Find(&data).Error

	return data, errorx.WithStack(err)
}

func (r *ResourceRepo) GetAllChildren(ctx context.Context, parentIds ...int64) ([]int64, error) {
	return r.getAllChildren(ctx, 0, parentIds...)
}

func (r *ResourceRepo) getAllChildren(ctx context.Context, depth int, parentIds ...int64) ([]int64, error) {
	var ids []int64

	if depth > 10 {
		return ids, nil
	}

	if err := r.DB(ctx).Where("`parent_id` IN ?", parentIds).Pluck("id", &ids).Error; err != nil {
		return nil, errorx.WithStack(err)
	}

	if len(ids) == 0 {
		return ids, nil
	}

	childrenIds, err := r.getAllChildren(ctx, depth+1, ids...)
	if err != nil {
		return nil, err
	}

	return append(ids, childrenIds...), nil
}

func (r *ResourceRepo) GetAllResource(ctx context.Context) ([]*adminmodel.Resource, error) {
	var data []*adminmodel.Resource
	if err := r.DB(ctx).Find(&data).Error; err != nil {
		return nil, errorx.WithStack(err)
	}

	return data, nil
}

func (r *ResourceRepo) UpdateResources(ctx context.Context, saveData []*adminmodel.Resource, deleteIds []int64) error {
	if err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if len(deleteIds) != 0 {
			if err := tx.Where("`id` IN ?", deleteIds).Delete(&adminmodel.Resource{}).Error; err != nil {
				return errorx.WithStack(err)
			}
		}

		if len(saveData) > 0 {
			tx.Omit("`create_time`", "`update_time`").Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{{Name: "id"}},
					DoUpdates: clause.AssignmentColumns([]string{"`parent_id`", "`name`", "`type`", "`path`", "`apis`", "`weight`", "`operator_id`"}),
				},
			).Create(&saveData)
		}

		return nil
	}); err != nil {
		return errorx.WithStack(err)
	}

	_ = r.cache.Del(ctx, resourceCacheKey)

	return nil
}
