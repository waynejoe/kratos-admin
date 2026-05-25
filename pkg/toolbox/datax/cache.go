package datax

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rand "math/rand/v2"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"

	"kratos-admin/pkg/toolbox/idx"
)

var (
	ErrNotFound = errors.New("not found")
)

// KeyNotFoundError 返回带key的not found错误
func KeyNotFoundError(key string) error {
	return fmt.Errorf("%w: key=%s", ErrNotFound, key)
}

func RandTTL(ttl int) func() int {
	return func() int {
		if ttl <= 0 {
			return 0
		}
		return ttl + rand.IntN(ttl/10)
	}
}

func decodeTTL(ttl any) int {
	switch v := ttl.(type) {
	case func() int:
		return v()
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	}
	// 未识别默认1个小时
	return 3600
}

type Cache interface {
	Get(ctx context.Context, key string, value any) error
	MGet(ctx context.Context, kv map[string]any) (map[string]any, map[string]any, error)
	Set(ctx context.Context, key string, value any, ttl any) error
	MSet(ctx context.Context, kv map[string]any, ttl any) error
	Del(ctx context.Context, keys ...string) error
}

type LocalCache struct {
	pubsub      *Pubsub
	cache       *freecache.Cache
	expireTopic string
	instanceId  int64
}

type expireEvent struct {
	InstanceId int64    `json:"instanceId"`
	Keys       []string `json:"keys"`
}

func (c *LocalCache) Get(ctx context.Context, key string, value any) error {
	val, err := c.cache.Get([]byte(key))
	if err == nil {
		return json.Unmarshal(val, value)
	} else if errors.Is(err, freecache.ErrNotFound) {
		return KeyNotFoundError(key)
	} else {
		return err
	}
}

func (c *LocalCache) MGet(ctx context.Context, kv map[string]any) (map[string]any, map[string]any, error) {
	hit := make(map[string]any)
	miss := make(map[string]any)
	for k, v := range kv {
		err := c.Get(ctx, k, v)
		if err == nil {
			hit[k] = v
		} else if errors.Is(err, ErrNotFound) {
			miss[k] = v
		} else {
			return nil, nil, err
		}
	}
	return hit, miss, nil
}

func (c *LocalCache) Set(ctx context.Context, key string, value any, ttl any) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set([]byte(key), val, c.ttl(ttl))
}

func (c *LocalCache) MSet(ctx context.Context, kv map[string]any, ttl any) error {
	for k, v := range kv {
		if err := c.Set(ctx, k, v, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (c *LocalCache) Del(ctx context.Context, keys ...string) error {
	c.del(ctx, keys...)
	return c.pubExpireEvent(ctx, keys)
}

//nolint:unparam
func (c *LocalCache) del(ctx context.Context, keys ...string) {
	for _, key := range keys {
		c.cache.Del([]byte(key))
	}
}

func (c *LocalCache) ttl(ttl any) int {
	res := decodeTTL(ttl)
	if res >= 3600 || res <= 0 {
		// 最多缓存1h
		return 3000 + rand.IntN(600)
	}
	return res/2 + rand.IntN(res/2)
}

func (c *LocalCache) pubExpireEvent(ctx context.Context, keys []string) error {
	// 通过广播来失效各个节点的key
	if len(keys) == 0 {
		return nil
	}
	event := &expireEvent{InstanceId: c.instanceId, Keys: keys}
	return c.pubsub.Publish(ctx, c.expireTopic, event)
}

func (c *LocalCache) subExpireEvent() {
	// 监听失效的key并删除
	ctx := context.Background()
	_ = c.pubsub.Subscribe(context.Background(), c.expireTopic, func(data []byte) {
		event := &expireEvent{}
		err := json.Unmarshal(data, event)
		if err != nil || event.InstanceId == c.instanceId {
			return
		}
		c.del(ctx, event.Keys...)
	})
}

type DistributeCache struct {
	rdb *redis.Client
}

func (c *DistributeCache) Get(ctx context.Context, key string, value any) error {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return json.Unmarshal([]byte(val), value)
	} else if errors.Is(err, redis.Nil) {
		return KeyNotFoundError(key)
	} else {
		return err
	}
}

func (c *DistributeCache) MGet(ctx context.Context, kv map[string]any) (map[string]any, map[string]any, error) {
	keys := make([]string, 0)
	for k := range kv {
		keys = append(keys, k)
	}
	val, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}
	hit := make(map[string]any)
	miss := make(map[string]any)
	for index, k := range keys {
		v := val[index]
		if v == nil {
			miss[k] = kv[k]
			continue
		}
		if err := json.Unmarshal([]byte(v.(string)), kv[k]); err != nil {
			return nil, nil, err
		}
		hit[k] = kv[k]
	}
	return hit, miss, nil
}

func (c *DistributeCache) Set(ctx context.Context, key string, value any, ttl any) error {
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = c.rdb.Set(ctx, key, bs, time.Duration(decodeTTL(ttl))*time.Second).Result()
	return err
}

func (c *DistributeCache) MSet(ctx context.Context, kv map[string]any, ttl any) error {
	_, err := c.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for k, v := range kv {
			bs, err := json.Marshal(v)
			if err != nil {
				return err
			}
			_ = pipe.Set(ctx, k, bs, time.Duration(decodeTTL(ttl))*time.Second)
		}
		return nil
	})
	return err
}

func (c *DistributeCache) Del(ctx context.Context, keys ...string) error {
	_, err := c.rdb.Del(ctx, keys...).Result()
	return err
}

type LevelCache struct {
	first      Cache
	second     Cache
	defaultTTL any
}

func (c *LevelCache) Get(ctx context.Context, key string, value any) error {
	err := c.first.Get(ctx, key, value)
	if err == nil {
		return nil
	} else if errors.Is(err, ErrNotFound) {
		err = c.second.Get(ctx, key, value)
		if err != nil {
			return err
		}
		return c.first.Set(ctx, key, value, c.defaultTTL)
	} else {
		return err
	}
}

func (c *LevelCache) MGet(ctx context.Context, kv map[string]any) (map[string]any, map[string]any, error) {
	firstHit, firstMiss, err := c.first.MGet(ctx, kv)
	if err != nil {
		return nil, nil, err
	}
	if len(firstMiss) == 0 {
		return firstHit, firstMiss, nil
	}
	secondHit, secondMiss, err := c.second.MGet(ctx, firstMiss)
	if err != nil {
		return nil, nil, err
	}
	if len(secondHit) > 0 {
		if err := c.first.MSet(ctx, secondHit, c.defaultTTL); err != nil {
			return nil, nil, err
		}
		for k, v := range secondHit {
			firstHit[k] = v
		}
	}
	return firstHit, secondMiss, nil
}

func (c *LevelCache) Set(ctx context.Context, key string, value any, ttl any) error {
	if err := c.second.Set(ctx, key, value, ttl); err != nil {
		return err
	}
	if err := c.first.Del(ctx, key); err != nil {
		return err
	}
	return c.first.Set(ctx, key, value, ttl)
}

func (c *LevelCache) MSet(ctx context.Context, kv map[string]any, ttl any) error {
	if err := c.second.MSet(ctx, kv, ttl); err != nil {
		return err
	}
	keys := make([]string, 0)
	for k := range kv {
		keys = append(keys, k)
	}
	if err := c.first.Del(ctx, keys...); err != nil {
		return err
	}
	return c.first.MSet(ctx, kv, ttl)
}

func (c *LevelCache) Del(ctx context.Context, keys ...string) error {
	if err := c.second.Del(ctx, keys...); err != nil {
		return err
	}
	return c.first.Del(ctx, keys...)
}

func NewDistributeCache(rdb *redis.Client) Cache {
	return &DistributeCache{rdb: rdb}
}

func NewLocalCache(size int, expireTopic string, pubsub *Pubsub) Cache {
	c := &LocalCache{pubsub: pubsub, cache: freecache.NewCache(1024 * 1024 * size), instanceId: idx.SnowId(), expireTopic: expireTopic}
	go c.subExpireEvent()
	return c
}

func NewLevelCache(first Cache, second Cache) Cache {
	return &LevelCache{first: first, second: second, defaultTTL: RandTTL(3600)}
}

func NewContextCache() Cache {
	return &ContextCache{}
}

func WithContextCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextCacheKey{}, &sync.Map{})
}

type contextCacheKey struct {
}

type ContextCache struct {
}

func (c *ContextCache) data(ctx context.Context) *sync.Map {
	val, ok := ctx.Value(contextCacheKey{}).(*sync.Map)
	if ok {
		return val
	}
	return nil
}

func (c *ContextCache) Get(ctx context.Context, key string, value any) error {
	data := c.data(ctx)
	if data == nil {
		return KeyNotFoundError(key)
	}
	bs, ok := data.Load(key)
	if !ok {
		return KeyNotFoundError(key)
	}
	return json.Unmarshal(bs.([]byte), value)
}

func (c *ContextCache) MGet(ctx context.Context, kv map[string]any) (map[string]any, map[string]any, error) {
	hit := make(map[string]any)
	miss := make(map[string]any)
	data := c.data(ctx)
	if data == nil {
		return hit, kv, nil
	}
	for k, v := range kv {
		bs, ok := data.Load(k)
		if !ok {
			miss[k] = v
			continue
		}
		err := json.Unmarshal(bs.([]byte), v)
		if err != nil {
			return nil, nil, err
		}
		hit[k] = v
	}
	return hit, miss, nil
}

func (c *ContextCache) Set(ctx context.Context, key string, value any, ttl any) error {
	data := c.data(ctx)
	if data == nil {
		return nil
	}
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data.Store(key, bs)
	return nil
}

func (c *ContextCache) MSet(ctx context.Context, kv map[string]any, ttl any) error {
	data := c.data(ctx)
	if data == nil {
		return nil
	}
	for k, v := range kv {
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data.Store(k, bs)
	}
	return nil
}

func (c *ContextCache) Del(ctx context.Context, keys ...string) error {
	data := c.data(ctx)
	if data == nil {
		return nil
	}
	for _, key := range keys {
		data.Delete(key)
	}
	return nil
}
