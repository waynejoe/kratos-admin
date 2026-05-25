package authz

import "github.com/google/wire"

// ProviderSet authz providers.
var ProviderSet = wire.NewSet(NewStore)
