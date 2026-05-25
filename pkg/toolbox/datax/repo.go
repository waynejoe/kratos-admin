package datax

import (
	"context"
	"errors"
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"
)

var limitRe = regexp.MustCompile(`(?i)\s+LIMIT\s+(\d+)(?:\s*,\s*(\d+))?(?:\s+OFFSET\s+(\d+))?`)

type Repo[T Entity] struct {
	db     *gorm.DB
	config *RepoConfig[T]
}

func (repo *Repo[T]) DB(ctx context.Context) *gorm.DB {
	db, _ := repo.Tx(ctx)
	return db
}

func (repo *Repo[T]) Tx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	if ok {
		return tx.Model(repo.config.model).WithContext(ctx), true
	}
	db := repo.db.Model(repo.config.model).WithContext(ctx)
	if _, ok := ctx.Value(forceMasterContextKey{}).(bool); repo.config.forceMaster || ok {
		db = db.Clauses(ForceMasterHint)
	}
	return db, false
}

func (repo *Repo[T]) RawGet(ctx context.Context, id int64) (*T, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var entity T
	err := repo.DB(ctx).Where("id = ?", id).Take(&entity).Error
	return &entity, err
}

func (repo *Repo[T]) RawGetsMap(ctx context.Context, ids []int64) (map[int64]*T, error) {
	ids = utils.Filter(ids, func(id int64) bool {
		return id != 0
	})
	entityMap := make(map[int64]*T)
	if len(ids) == 0 {
		return entityMap, nil
	}
	entityList := make([]*T, 0)
	err := repo.DB(ctx).Where("id in ?", ids).Find(&entityList).Error
	if err != nil {
		return nil, err
	}
	for _, entity := range entityList {
		entityMap[any(entity).(Entity).PKVal()] = entity
	}
	return entityMap, nil
}

func (repo *Repo[T]) RawGets(ctx context.Context, ids []int64) ([]*T, error) {
	ids = utils.Distinct(ids, func(id int64) int64 { return id })

	// 结果保证按id顺序
	entityMap, err := repo.RawGetsMap(ctx, ids)
	if err != nil {
		return nil, err
	}
	entityList := make([]*T, 0)
	for _, id := range ids {
		if entity, ok := entityMap[id]; ok {
			entityList = append(entityList, entity)
		}
	}
	return entityList, nil
}

func (repo *Repo[T]) Create(ctx context.Context, entity *T) error {
	err := repo.DB(ctx).Omit(omitColumns...).Create(entity).Error
	return errorx.WithStack(err)
}

func (repo *Repo[T]) Update(ctx context.Context, entity *T, columns ...string) error {
	current := any(entity).(Entity)

	db := repo.DB(ctx).Omit(omitColumns...).Where("id = ?", current.PKVal())

	if len(columns) > 0 {
		// 更新指定列
		db = db.Select(columns)
	} else {
		// 更新全部列
		db = db.Select("*")
	}

	err := db.Updates(entity).Error
	return errorx.WithStack(err)
}

func (repo *Repo[T]) Save(ctx context.Context, entity *T) error {
	if any(entity).(Entity).PKVal() != 0 {
		return repo.Update(ctx, entity)
	} else {
		return repo.Create(ctx, entity)
	}
}

func (repo *Repo[T]) DelById(ctx context.Context, id int64) error {
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return err
	}
	return repo.Del(ctx, entity)
}

func (repo *Repo[T]) Del(ctx context.Context, entity *T) error {
	if any(entity).(Entity).PKVal() == 0 {
		//nolint:wrapcheck
		return errors.New("pk val = 0")
	}
	err := repo.DB(ctx).Delete(entity).Error
	return errorx.WithStack(err)
}

func (repo *Repo[T]) GetIdsBy(ctx context.Context, query string, args ...any) ([]int64, error) {
	ids := make([]int64, 0)
	err := repo.GetsFieldBy(ctx, "id", &ids, query, args...)
	return ids, err
}

func (repo *Repo[T]) GetsFieldBy(ctx context.Context, field string, res any, query string, args ...any) error {
	return repo.DB(ctx).Where(query, args...).Pluck(field, res).Error
}

func (repo *Repo[T]) RawGetsBy(ctx context.Context, query string, args ...any) ([]*T, error) {
	entityList := make([]*T, 0)
	err := repo.DB(ctx).Where(query, args...).Find(&entityList).Error
	return entityList, err
}

func (repo *Repo[T]) GetIdBy(ctx context.Context, query string, args ...any) (int64, error) {
	ids, err := repo.GetIdsBy(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	return ids[0], nil
}

func (repo *Repo[T]) RawGetBy(ctx context.Context, query string, args ...any) (*T, error) {
	var entity T
	err := repo.DB(ctx).Where(limitRe.ReplaceAllString(query, ""), args...).Take(&entity).Error
	return &entity, err
}

func (repo *Repo[T]) Visible(entity *T) bool {
	if entity == nil {
		return false
	}
	return repo.config.visible(entity)
}

func (repo *Repo[T]) FilterVisible(list []*T) []*T {
	res := make([]*T, 0)
	for _, entity := range list {
		if repo.Visible(entity) {
			res = append(res, entity)
		}
	}
	return res
}

func (repo *Repo[T]) Get(ctx context.Context, id int64) (*T, error) {
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return nil, err
	}
	if repo.Visible(entity) {
		return entity, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (repo *Repo[T]) Gets(ctx context.Context, ids []int64) ([]*T, error) {
	list, err := repo.RawGets(ctx, ids)
	if err != nil {
		return nil, err
	}
	return repo.FilterVisible(list), nil
}

func (repo *Repo[T]) GetsMap(ctx context.Context, ids []int64) (map[int64]*T, error) {
	entities, err := repo.RawGetsMap(ctx, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*T)
	for k, v := range entities {
		if repo.Visible(v) {
			res[k] = v
		}
	}
	return res, nil
}

func (repo *Repo[T]) GetBy(ctx context.Context, query string, args ...any) (*T, error) {
	entity, err := repo.RawGetBy(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if repo.Visible(entity) {
		return entity, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (repo *Repo[T]) GetsBy(ctx context.Context, query string, args ...any) ([]*T, error) {
	list, err := repo.RawGetsBy(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return repo.FilterVisible(list), nil
}

func (repo *Repo[T]) CutGets(ctx context.Context, sep string, args ...any) ([]*T, error) {
	return repo.Gets(ctx, utils.CutIds(sep, args...))
}

func (repo *Repo[T]) Lock(ctx context.Context, id int64) (*T, error) {
	tx, ok := repo.Tx(ctx)
	if !ok {
		return nil, errorx.New("lock must be in transaction context")
	}
	var entity T
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&entity).Error
	return &entity, err
}

// Upsert 唯一索引/主键冲突执行update
func (repo *Repo[T]) Upsert(ctx context.Context, entity *T, columns ...string) error {
	id := any(entity).(Entity).PKVal()
	if id > 0 {
		return repo.Update(ctx, entity, columns...)
	}
	conflict := clause.OnConflict{UpdateAll: true}
	if len(columns) > 0 {
		conflict.UpdateAll = false
		conflict.DoUpdates = clause.AssignmentColumns(columns)
	}
	err := repo.DB(ctx).Clauses(conflict).Omit(omitColumns...).Create(entity).Error
	return errorx.WithStack(err)
}

// InsertIgnore 唯一索引/主键冲突执行insert ignore
func (repo *Repo[T]) InsertIgnore(ctx context.Context, entity *T) error {
	err := repo.DB(ctx).Clauses(clause.Insert{
		Modifier: "IGNORE",
	}).Omit(omitColumns...).Create(entity).Error
	return errorx.WithStack(err)
}

func (repo *Repo[T]) Scan(ctx context.Context, batchSize int, fn func(*T) error) error {
	return ScanRepo(ctx, repo, batchSize, fn)
}

func NewRepo[T Entity](db *gorm.DB, opts ...func(config *RepoConfig[T])) Repo[T] {
	config := NewRepoConfig[T](opts...)
	return Repo[T]{db: db, config: config}
}
