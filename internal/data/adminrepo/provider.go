package adminrepo

import "github.com/google/wire"

// ProviderSet is admin repository providers.
var ProviderSet = wire.NewSet(
	NewAdminUserRepo,
	NewAdminGroupRepo,
	NewAdminRoleRepo,
	NewResourceRepo,
	NewOperationLogRepo,
)
