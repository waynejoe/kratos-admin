//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"kratos-admin/internal/authz"
	"kratos-admin/internal/biz"
	"kratos-admin/internal/conf"
	"kratos-admin/internal/data"
	"kratos-admin/internal/data/adminrepo"
	"kratos-admin/internal/server"
	"kratos-admin/internal/service"
)

func wireApp(
	*conf.Server,
	*conf.Data,
	*conf.Security,
	*conf.Permission,
	log.Logger,
) (*kratos.App, error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		authz.ProviderSet,
		adminrepo.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
