package datax

import (
	"context"
	"fmt"
)

type KVRepo[T Entity] struct {
	cache  Cache
	config *RepoConfig[T]
}

func (c *KVRepo[T]) Get(ctx context.Context, id int64) (*T, error) {
	var entity T
	err := c.cache.Get(ctx, c.keyByPK(id), &entity)
	return &entity, err
}

func (c *KVRepo[T]) MGet(ctx context.Context, ids []int64) (map[int64]*T, []int64, error) {
	entityMap := make(map[string]any)
	for _, id := range ids {
		entityMap[c.keyByPK(id)] = new(T)
	}
	entities := make(map[int64]*T)
	missIds := make([]int64, 0)
	hit, _, err := c.cache.MGet(ctx, entityMap)
	if err != nil {
		return nil, nil, err
	}
	for _, entity := range hit {
		entities[entity.(Entity).PKVal()] = entity.(*T)
	}
	for _, id := range ids {
		if _, ok := entities[id]; !ok {
			missIds = append(missIds, id)
		}
	}
	return entities, missIds, nil
}

func (c *KVRepo[T]) Set(ctx context.Context, id int64, entity *T) error {
	return c.cache.Set(ctx, c.keyByPK(id), entity, c.config.cacheTTL)
}

func (c *KVRepo[T]) MSet(ctx context.Context, entityList []*T) error {
	entityMap := make(map[string]any)
	for _, entity := range entityList {
		entityMap[c.keyByPK(any(entity).(Entity).PKVal())] = entity
	}
	return c.cache.MSet(ctx, entityMap, c.config.cacheTTL)
}

func (c *KVRepo[T]) Del(ctx context.Context, entity *T) error {
	key := c.keyByPK(any(entity).(Entity).PKVal())
	return c.cache.Del(ctx, key)
}

func (c *KVRepo[T]) DelById(ctx context.Context, id int64) error {
	key := c.keyByPK(id)
	return c.cache.Del(ctx, key)
}

func (c *KVRepo[T]) keyByPK(id int64) string {
	return fmt.Sprintf("%s:%s:%d", c.config.cachePrefix, c.config.tableName, id)
}

func NewKVRepo[T Entity](cache Cache, config *RepoConfig[T]) KVRepo[T] {
	ec := KVRepo[T]{cache: cache, config: config}
	return ec
}
