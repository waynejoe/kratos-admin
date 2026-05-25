package server

import (
	validate "github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/biz/permissionbiz"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data"
	"kratos-admin/internal/data/adminrepo"
	"kratos-admin/internal/middleware"
	"kratos-admin/internal/service"
)

func NewHTTPServer(
	c *conf.Server,
	security *conf.Security,
	permissionConfig *conf.Permission,
	data *data.Data,
	operationLogRepo *adminrepo.OperationLogRepo,
	adminUserRepo *adminrepo.AdminUserRepo,
	authzStore *authz.Store,
	userPermissionUsecase *permissionbiz.UserPermissionUsecase,
	permissionService *service.PermissionService,
	adminUserService *service.AdminUserService,
) *http.Server {
	opts := []http.ServerOption{
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
			handlers.AllowedHeaders([]string{
				"Content-Type",
				"Authorization",
				"X-Requested-With",
				"X-Trace-ID",
				"Origin",
				"Accept",
				"X-OptimizerId",
			}),
			handlers.AllowCredentials(),
		)),
		http.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			validate.ProtoValidate(),
			middleware.CacheMiddleware,
			middleware.AuthMiddleware(security, permissionConfig, data, authzStore, userPermissionUsecase),
			middleware.AuditMiddleware(operationLogRepo, adminUserRepo),
		),
	}
	if c.HTTP.Addr != "" {
		opts = append(opts, http.Address(c.HTTP.Addr))
	}
	srv := http.NewServer(opts...)
	_ = permissionService.InitService(srv)
	_ = adminUserService.InitService(srv)
	return srv
}
