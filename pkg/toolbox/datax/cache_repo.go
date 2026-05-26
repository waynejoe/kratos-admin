package datax

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"
)

type CacheRepo[T Entity] struct {
	db     *gorm.DB
	kv     KVRepo[T]
	config *RepoConfig[T]
}

func (repo *CacheRepo[T]) DB(ctx context.Context) *gorm.DB {
	db, _ := repo.Tx(ctx)
	return db
}

func (repo *CacheRepo[T]) Tx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	if ok {
		return tx.Model(&repo.config.model).WithContext(ctx), true
	}
	db := repo.db.Model(&repo.config.model).WithContext(ctx)
	if _, ok := ctx.Value(forceMasterContextKey{}).(bool); repo.config.forceMaster || ok {
		db = db.Clauses(ForceMasterHint)
	}
	return db, false
}

func (repo *CacheRepo[T]) RawGet(ctx context.Context, id int64) (*T, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	tx, ok := repo.Tx(ctx)
	if ok {
		// 走事务
		var entity T
		err := tx.Where("id = ?", id).Take(&entity).Error
		return &entity, err
	}
	entityPtr, err := repo.kv.Get(ctx, id)
	if err == nil {
		if any(entityPtr).(Entity).PKVal() == 0 {
			// 命中主键却为空值,返回ErrRecordNotFound
			return nil, gorm.ErrRecordNotFound
		}
		return entityPtr, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if !repo.config.cacheOrigin {
		return nil, gorm.ErrRecordNotFound
	}
	var entity T
	err = tx.Where("id = ?", id).Take(&entity).Error
	if err == nil {
		err = repo.kv.Set(ctx, id, &entity)
		return &entity, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 防止缓存穿透, NotFound时也写入
		if err := repo.kv.Set(ctx, id, &entity); err != nil {
			return nil, err
		}
		return nil, gorm.ErrRecordNotFound
	} else {
		return nil, err
	}
}

func (repo *CacheRepo[T]) RawGetsMap(ctx context.Context, ids []int64) (map[int64]*T, error) {
	ids = utils.Filter(ids, func(id int64) bool {
		return id != 0
	})
	if len(ids) == 0 {
		return make(map[int64]*T), nil
	}
	entityMap, missIds, err := repo.kv.MGet(ctx, ids)
	if err != nil {
		return nil, err
	}
	if repo.config.cacheOrigin && len(missIds) > 0 {
		entityList := make([]*T, 0)
		err := repo.DB(ctx).Where("id in ?", missIds).Find(&entityList).Error
		if err != nil {
			return nil, err
		}
		if err := repo.kv.MSet(ctx, entityList); err != nil {
			return nil, err
		}
		for _, entity := range entityList {
			entityMap[any(entity).(Entity).PKVal()] = entity
		}
	}
	return entityMap, nil
}

func (repo *CacheRepo[T]) RawGets(ctx context.Context, ids []int64) ([]*T, error) {
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

func (repo *CacheRepo[T]) Create(ctx context.Context, entity *T) error {
	if repo.config.readonly {
		return fmt.Errorf("%s readonly", repo.config.tableName)
	}
	err := repo.db.WithContext(ctx).Model(entity).Omit(omitColumns...).Create(entity).Error
	if err != nil {
		return errorx.WithStack(err)
	}
	return repo.kv.Del(ctx, entity)
}

func (repo *CacheRepo[T]) Update(ctx context.Context, entity *T, columns ...string) error {
	if repo.config.readonly {
		return fmt.Errorf("%s readonly", repo.config.tableName)
	}
	current := any(entity).(Entity)

	db := repo.db.WithContext(ctx).Model(entity).Where("id = ?", current.PKVal())

	if len(columns) > 0 {
		// 更新指定列
		db = db.Select(columns)
	} else {
		// 更新全部列
		db = db.Select("*")
	}

	err := db.Omit(omitColumns...).Updates(entity).Error
	if err != nil {
		return errorx.WithStack(err)
	}

	return repo.kv.Del(ctx, entity)
}

func (repo *CacheRepo[T]) Save(ctx context.Context, entity *T) error {
	if any(entity).(Entity).PKVal() != 0 {
		return repo.Update(ctx, entity)
	} else {
		return repo.Create(ctx, entity)
	}
}

func (repo *CacheRepo[T]) DelById(ctx context.Context, id int64) error {
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return err
	}
	return repo.Del(ctx, entity)
}

func (repo *CacheRepo[T]) Del(ctx context.Context, entity *T) error {
	if repo.config.readonly {
		return fmt.Errorf("%s readonly", repo.config.tableName)
	}
	if any(entity).(Entity).PKVal() == 0 {
		//nolint:wrapcheck
		return errors.New("pk val = 0")
	}
	err := repo.db.WithContext(ctx).Model(entity).Delete(entity).Error
	if err != nil {
		return errorx.WithStack(err)
	}
	return repo.kv.Del(ctx, entity)
}

func (repo *CacheRepo[T]) GetIdsBy(ctx context.Context, query string, args ...any) ([]int64, error) {
	ids := make([]int64, 0)

	if err := repo.GetsFieldBy(ctx, "id", &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (repo *CacheRepo[T]) GetsFieldBy(ctx context.Context, field string, res any, query string, args ...any) error {
	return repo.DB(ctx).Where(query, args...).Pluck(field, res).Error
}

func (repo *CacheRepo[T]) RawGetsBy(ctx context.Context, query string, args ...any) ([]*T, error) {
	entityList := make([]*T, 0)
	err := repo.DB(ctx).Where(query, args...).Find(&entityList).Error
	return entityList, err
}

func (repo *CacheRepo[T]) GetIdBy(ctx context.Context, query string, args ...any) (int64, error) {
	ids, err := repo.GetIdsBy(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return ids[0], nil
}

func (repo *CacheRepo[T]) RawGetBy(ctx context.Context, query string, args ...any) (*T, error) {
	var entity T
	err := repo.DB(ctx).Where(limitRe.ReplaceAllString(query, ""), args...).Take(&entity).Error
	return &entity, err
}

func (repo *CacheRepo[T]) Visible(entity *T) bool {
	if entity == nil {
		return false
	}
	return repo.config.visible(entity)
}

func (repo *CacheRepo[T]) FilterVisible(list []*T) []*T {
	res := make([]*T, 0)
	for _, entity := range list {
		if repo.Visible(entity) {
			res = append(res, entity)
		}
	}
	return res
}

func (repo *CacheRepo[T]) Get(ctx context.Context, id int64) (*T, error) {
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return nil, err
	}
	if repo.Visible(entity) {
		return entity, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (repo *CacheRepo[T]) Gets(ctx context.Context, ids []int64) ([]*T, error) {
	list, err := repo.RawGets(ctx, ids)
	if err != nil {
		return nil, err
	}
	return repo.FilterVisible(list), nil
}

func (repo *CacheRepo[T]) GetsMap(ctx context.Context, ids []int64) (map[int64]*T, error) {
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

func (repo *CacheRepo[T]) GetBy(ctx context.Context, query string, args ...any) (*T, error) {
	entity, err := repo.RawGetBy(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if repo.Visible(entity) {
		return entity, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (repo *CacheRepo[T]) GetsBy(ctx context.Context, query string, args ...any) ([]*T, error) {
	list, err := repo.RawGetsBy(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return repo.FilterVisible(list), nil
}

func (repo *CacheRepo[T]) CutGets(ctx context.Context, sep string, args ...any) ([]*T, error) {
	return repo.Gets(ctx, utils.CutIds(sep, args...))
}

func (repo *CacheRepo[T]) DirectGet(ctx context.Context, id int64) (*T, error) {
	var entity T
	err := repo.DB(ctx).Where("id = ?", id).Take(&entity).Error
	return &entity, err
}

func (repo *CacheRepo[T]) Lock(ctx context.Context, id int64) (*T, error) {
	tx, ok := repo.Tx(ctx)
	if !ok {
		return nil, errorx.New("lock must be in transaction context")
	}
	var entity T
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&entity).Error
	return &entity, err
}

func (repo *CacheRepo[T]) Upsert(ctx context.Context, entity *T, columns ...string) error {
	if repo.config.readonly {
		return fmt.Errorf("%s readonly", repo.config.tableName)
	}
	id := any(entity).(Entity).PKVal()
	if id > 0 {
		return repo.Update(ctx, entity, columns...)
	}
	conflict := clause.OnConflict{UpdateAll: true}
	if len(columns) > 0 {
		// 指定字段更新
		conflict.UpdateAll = false
		conflict.DoUpdates = clause.AssignmentColumns(columns)
	}
	err := repo.DB(ctx).Clauses(conflict).Omit(omitColumns...).Create(entity).Error
	if err != nil {
		return errorx.WithStack(err)
	}
	return repo.kv.Del(ctx, entity)
}

func (repo *CacheRepo[T]) InsertIgnore(ctx context.Context, entity *T) error {
	if repo.config.readonly {
		return fmt.Errorf("%s readonly", repo.config.tableName)
	}
	err := repo.DB(ctx).Clauses(clause.Insert{
		Modifier: "IGNORE",
	}).Omit(omitColumns...).Create(entity).Error
	if err != nil {
		return errorx.WithStack(err)
	}
	return repo.kv.Del(ctx, entity)
}

func (repo *CacheRepo[T]) Scan(ctx context.Context, batchSize int, fn func(*T) error) error {
	return ScanRepo(ctx, repo, batchSize, fn)
}

func NewCacheRepo[T Entity](db *gorm.DB, cache Cache, opts ...func(config *RepoConfig[T])) CacheRepo[T] {
	config := NewRepoConfig[T](opts...)
	r := CacheRepo[T]{
		db:     db,
		kv:     NewKVRepo[T](cache, config),
		config: config,
	}
	return r
}
