package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"kratos-admin/internal/conf"
	"kratos-admin/internal/service"
	pb "kratos-admin/pb/admin/v1"
)

func NewGRPCServer(
	c *conf.Server,
	permissionService *service.PermissionService,
	adminUserService *service.AdminUserService,
) *grpc.Server {
	opts := make([]grpc.ServerOption, 0, 2)
	if c.GRPC.Addr != "" {
		opts = append(opts, grpc.Address(c.GRPC.Addr))
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterPermissionServer(srv, permissionService)
	pb.RegisterAdminUserServer(srv, adminUserService)
	return srv
}
