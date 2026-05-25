package data

import (
	"kratos-admin/pkg/toolbox/authz"
	"kratos-admin/pkg/toolbox/errorx"
	"github.com/redis/go-redis/v9"

	"kratos-admin/internal/conf"
)

func NewCasbinServer(c *conf.Data, data *Data) (*authz.CasbinServer, error) {
	casbinServer, err := authz.NewCasbinServer(
		data.G.AdminDB,
		&redis.Options{Addr: c.Global.Redis.Addr, Password: c.Global.Redis.Password, DB: int(c.Global.Redis.DB)},
	)
	if err != nil {
		return nil, errorx.WithStack(err)
	}
	return casbinServer.(*authz.CasbinServer), nil
}
