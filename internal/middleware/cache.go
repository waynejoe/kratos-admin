package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"kratos-admin/pkg/toolbox/datax"
)

func CacheMiddleware(handler middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return handler(datax.WithContextCache(ctx), req)
	}
}
