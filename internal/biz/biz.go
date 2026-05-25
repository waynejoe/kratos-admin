package biz

import (
	"github.com/google/wire"

	"kratos-admin/internal/biz/adminuserbiz"
	"kratos-admin/internal/biz/auditbiz"
	"kratos-admin/internal/biz/permissionbiz"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	adminuserbiz.NewAdminUserUsecase,
	adminuserbiz.NewLoginUsecase,
	permissionbiz.NewAdminGroupUsecase,
	permissionbiz.NewAdminRoleUsecase,
	permissionbiz.NewResourceUsecase,
	permissionbiz.NewUserPermissionUsecase,
	auditbiz.NewAuditUsecase,
)
