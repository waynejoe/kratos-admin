package logx

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

func LogValue[T any](fn func(ctx context.Context) T) log.Valuer {
	return func(ctx context.Context) any {
		value := fn(ctx)
		return value
	}
}
