package datax

import (
	"context"

	"gorm.io/gorm"
)

type IRepo[T Entity] interface {
	DB(context.Context) *gorm.DB
}

type ISyncRepo interface {
	SyncAll(ctx context.Context) error
	SyncById(ctx context.Context, id int64) error
}
