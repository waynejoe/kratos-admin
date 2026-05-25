package datax

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/hints"
)

var (
	cachePrefix         = "ec"
	cacheVersion        = "v1"
	cacheTTL        any = RandTTL(3600 * 24)
	ForceMasterHint     = hints.CommentBefore("SELECT", "FORCE_MASTER")
)

type RepoConfig[T Entity] struct {
	model         T
	tableName     string
	readonly      bool
	serialVersion string
	forceMaster   bool
	visible       func(*T) bool
	cachePrefix   string
	cacheVersion  string
	cacheTTL      any
	cacheOrigin   bool
}

type forceMasterContextKey struct {
}

type txContextKey struct {
}

func NewRepoConfig[T Entity](opts ...func(c *RepoConfig[T])) *RepoConfig[T] {
	visible := func(entity *T) bool {
		return true
	}
	model := *new(T)
	config := &RepoConfig[T]{
		model:         model,
		tableName:     model.TableName(),
		readonly:      false,
		serialVersion: GetSerialVersion(model),
		visible:       visible,
		cachePrefix:   cachePrefix,
		cacheVersion:  cacheVersion,
		cacheTTL:      cacheTTL,
		cacheOrigin:   true,
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func WithReadOnly[T Entity]() func(c *RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.readonly = true
	}
}

func WithForceMaster[T Entity]() func(c *RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.forceMaster = true
	}
}

func WithVisible[T Entity](visible func(*T) bool) func(*RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.visible = visible
	}
}

func WithCachePrefix[T Entity](prefix string) func(*RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.cachePrefix = prefix
	}
}

func WithCacheVersion[T Entity](version string) func(*RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.cacheVersion = version
	}
}

func WithCacheTTL[T Entity](ttl any) func(*RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.cacheTTL = ttl
	}
}

// WithCacheOriginDisable 缓存miss禁止回源
func WithCacheOriginDisable[T Entity]() func(c *RepoConfig[T]) {
	return func(c *RepoConfig[T]) {
		c.cacheOrigin = false
	}
}

func SetCacheVersion(version string) {
	cacheVersion = version
}

func SetCachePrefix(prefix string) {
	cachePrefix = prefix
}

func SetCacheTTL(ttl any) {
	cacheTTL = ttl
}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func ForceMaster(ctx context.Context) context.Context {
	return context.WithValue(ctx, forceMasterContextKey{}, true)
}
