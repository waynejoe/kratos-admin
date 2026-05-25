package datax

import (
	"context"
)

func ScanRepo[T Entity](ctx context.Context, repo IRepo[T], batchSize int, fn func(*T) error) error {
	if batchSize <= 10 {
		batchSize = 10
	}
	db := repo.DB(ctx)
	var lastId int64
	for {
		list := make([]*T, 0)
		err := db.Where("`id` > ?", lastId).Order("`id` ASC").Limit(batchSize).Find(&list).Error
		if err != nil {
			return err
		}
		for _, entity := range list {
			if err := fn(entity); err != nil {
				return err
			}
		}
		if len(list) < batchSize {
			return nil
		}
		lastId = any(list[len(list)-1]).(Entity).PKVal()
	}
}
