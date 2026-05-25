package middleware

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"

	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/logx"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
	"kratos-admin/pkg/toolbox/claim"
)

type permissionServer interface {
	GetUserDataPermission(ctx context.Context, userId int64) (*adminmodel.UserDataPermission, error)
}

func AuthMiddleware(
	security *conf.Security,
	permissionConfig *conf.Permission,
	data *data.Data,
	authzStore *authz.Store,
	permissionSrv permissionServer,
) middleware.Middleware {
	whitePaths := utils.NewSet(security.JwtSkipPaths...)
	whitePaths.Add(security.ThirdPaths...)
	permissionWhitePaths := utils.NewSet(permissionConfig.SkipPaths...)
	datapermPaths := utils.NewSet(permissionConfig.DataPermissionPaths...)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			httpReq, ok := http.RequestFromServerContext(ctx)
			if !ok || whitePaths.Has(fmt.Sprintf("%s:%s", httpReq.Method, httpReq.URL.Path)) {
				return handler(ctx, req)
			}
			header, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, pb.ErrorUnauthorized("wrong context")
			}

			claims := &claim.AdminUserClaim{}
			token := strings.TrimPrefix(header.RequestHeader().Get("Authorization"), "Bearer ")
			if token == "" {
				return nil, pb.ErrorUnauthorized("invalid access token")
			}

			if _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(security.JwtSecret), nil
			}); err != nil {
				return nil, pb.ErrorUnauthorized("invalid access token")
			}

			var dbAccessToken string
			if err := data.G.LocalCache.Get(ctx, fmt.Sprintf("admin:auth:access:%d", claims.UserId), &dbAccessToken); err != nil {
				return nil, pb.ErrorUnauthorized("invalid access token")
			}
			if token != dbAccessToken {
				return nil, pb.ErrorUnauthorized("invalid access token")
			}

			helpx.SetUserId(ctx, claims.UserId)

			if permissionConfig.ClosePermission || permissionWhitePaths.Has(fmt.Sprintf("%s:%s", httpReq.Method, httpReq.URL.Path)) {
				SetDataPermission(ctx, nil)
				return handler(ctx, req)
			}

			ok, err := authzStore.CheckAPIPermission(claims.UserId, httpReq.URL.Path)
			if err != nil {
				logx.Errorf(ctx, "CheckPermission failed: %+v", err)
				return nil, err
			}
			if !ok {
				return nil, pb.ErrorForbidden("no permission")
			}

			if !datapermPaths.Has(fmt.Sprintf("%s:%s", httpReq.Method, httpReq.URL.Path)) {
				return handler(ctx, req)
			}

			dataPermission, err := permissionSrv.GetUserDataPermission(ctx, claims.UserId)
			if err != nil {
				logx.Errorf(ctx, "GetUserDataPermission failed: %+v", err)
				return handler(ctx, req)
			}

			if !dataPermission.HasDataPermission {
				return nil, pb.ErrorForbidden("no data permission")
			}

			optimizerIdStr := strings.TrimSpace(header.RequestHeader().Get("X-OptimizerId"))
			optimizerId, _ := strconv.ParseInt(optimizerIdStr, 10, 64)

			if optimizerId > searchPrivateData && (slices.Contains(dataPermission.UserIds, optimizerId) || len(dataPermission.UserIds) == 0) {
				SetDataPermission(ctx, dataPermission.UserIds)
				return handler(ctx, req)
			}
			if optimizerId == searchAllData {
				SetDataPermission(ctx, dataPermission.UserIds)
				return handler(ctx, req)
			}

			return nil, pb.ErrorForbidden("no data permission")
		}
	}
}

const (
	searchAllData     = -1
	searchPrivateData = 0
)
