package main

import (
	"flag"
	"log"

	"github.com/go-kratos/kratos/v2"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"kratos-admin/internal/conf"
	"kratos-admin/pkg/toolbox/authz"
)

var (
	Name    = "kratos-admin"
	Version = "dev"
)

func main() {
	confPath := flag.String("conf", "../../configs/config.yaml", "config path")
	flag.Parse()

	bc, err := conf.Load(*confPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := wireApp(&bc.Server, &bc.Data, &bc.Security, &bc.Permission, klog.DefaultLogger)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func newApp(logger klog.Logger, hs *http.Server, gs *grpc.Server, cs *authz.CasbinServer) *kratos.App {
	return kratos.New(
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Logger(logger),
		kratos.Server(hs, gs, cs),
	)
}
