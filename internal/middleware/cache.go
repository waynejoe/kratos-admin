package middleware

import (
	"context"

	"kratos-admin/pkg/toolbox/datax"
	"github.com/go-kratos/kratos/v2/middleware"
)

func CacheMiddleware(handler middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return handler(datax.WithContextCache(ctx), req)
	}
}
