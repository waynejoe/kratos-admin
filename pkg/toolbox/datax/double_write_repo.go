package datax

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"kratos-admin/pkg/toolbox/logx"
)

// DoubleWriteRepo 数据库&缓存双写
type DoubleWriteRepo[T Entity] struct {
	Repo[T]
	kv KVRepo[T]
}

func NewDoubleWriteRepo[T Entity](db *gorm.DB, cache Cache, opts ...func(config *RepoConfig[T])) DoubleWriteRepo[T] {
	config := NewRepoConfig[T](opts...)
	repo := DoubleWriteRepo[T]{
		Repo: NewRepo[T](db, opts...),
		kv:   NewKVRepo[T](cache, config),
	}
	return repo
}

func (repo *DoubleWriteRepo[T]) Create(ctx context.Context, entity *T) error {
	err := repo.Repo.Create(ctx, entity)
	if err != nil {
		return err
	}
	return repo.SyncById(ctx, any(entity).(Entity).PKVal())
}

func (repo *DoubleWriteRepo[T]) Update(ctx context.Context, entity *T, columns ...string) error {
	err := repo.Repo.Update(ctx, entity, columns...)
	if err != nil {
		return err
	}
	return repo.SyncById(ctx, any(entity).(Entity).PKVal())
}

func (repo *DoubleWriteRepo[T]) Save(ctx context.Context, entity *T) error {
	if any(entity).(Entity).PKVal() != 0 {
		return repo.Update(ctx, entity)
	} else {
		return repo.Create(ctx, entity)
	}
}

func (repo *DoubleWriteRepo[T]) DelById(ctx context.Context, id int64) error {
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return err
	}
	return repo.Del(ctx, entity)
}

func (repo *DoubleWriteRepo[T]) Del(ctx context.Context, entity *T) error {
	err := repo.Repo.Del(ctx, entity)
	if err != nil {
		return err
	}
	return repo.kv.Del(ctx, entity)
}

// Upsert 唯一索引/主键冲突执行update
func (repo *DoubleWriteRepo[T]) Upsert(ctx context.Context, entity *T, columns ...string) error {
	err := repo.Repo.Upsert(ctx, entity, columns...)
	if err != nil {
		return err
	}
	return repo.SyncById(ctx, any(entity).(Entity).PKVal())
}

// InsertIgnore 唯一索引/主键冲突执行insert ignore
func (repo *DoubleWriteRepo[T]) InsertIgnore(ctx context.Context, entity *T) error {
	err := repo.Repo.InsertIgnore(ctx, entity)
	if err != nil {
		return err
	}
	return repo.SyncById(ctx, any(entity).(Entity).PKVal())
}

func (repo *DoubleWriteRepo[T]) SyncAll(ctx context.Context) error {
	// db & kv 数据同步
	return ScanRepo(ctx, repo, 100, func(entity *T) error {
		return repo.kv.Set(ctx, any(entity).(Entity).PKVal(), entity)
	})
}

func (repo *DoubleWriteRepo[T]) SyncById(ctx context.Context, id int64) error {
	if _, ok := repo.Tx(ctx); ok {
		// 事务容错
		// 补充超时机制，防止泄露
		timeout, _ := context.WithTimeout(ctx, 3*time.Second)
		go func() {
			syncCtx := context.Background()
			<-timeout.Done()
			entity, err := repo.RawGet(syncCtx, id)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 事务提交失败，记录不存在，删除数据
				_ = repo.kv.DelById(syncCtx, id)
			} else if err != nil {
				logx.Errorf(ctx, "sync error %s %d", repo.config.tableName, id)
			} else {
				// 事务有可能成功也有可能失败，刷最新数据
				_ = repo.kv.Set(syncCtx, id, entity)
			}
		}()
	}
	entity, err := repo.RawGet(ctx, id)
	if err != nil {
		return err
	}
	return repo.kv.Set(ctx, id, entity)
}
