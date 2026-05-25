package data

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"kratos-admin/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewDataFromConfig,
	NewCasbinServer,
)

func NewDataFromConfig(c *conf.Data) (*Data, error) {
	return NewData(
		c.Global.AdminDatabase.Source,
		&redis.Options{
			Addr:     c.Global.Redis.Addr,
			Password: c.Global.Redis.Password,
			DB:       int(c.Global.Redis.DB),
		},
	)
}
