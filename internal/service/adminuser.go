package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"

	"kratos-admin/internal/biz/adminuserbiz"
	"kratos-admin/internal/biz/auditbiz"
	pb "kratos-admin/pb/admin/v1"
)

type AdminUserService struct {
	pb.UnimplementedAdminUserServer

	adminUserUsecase *adminuserbiz.AdminUserUsecase
	loginUsecase     *adminuserbiz.LoginUsecase
	auditUsecase     *auditbiz.AuditUsecase
}

func NewAdminUserService(
	adminUserUsecase *adminuserbiz.AdminUserUsecase,
	loginUsecase *adminuserbiz.LoginUsecase,
	auditUsecase *auditbiz.AuditUsecase,
) *AdminUserService {
	return &AdminUserService{
		adminUserUsecase: adminUserUsecase,
		loginUsecase:     loginUsecase,
		auditUsecase:     auditUsecase,
	}
}

func (s *AdminUserService) InitService(srv *http.Server) error {
	pb.RegisterAdminUserHTTPServer(srv, s)
	return nil
}

func (s *AdminUserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	token, err := s.loginUsecase.Login(ctx, req.Username, req.Password)
	return &pb.LoginReply{Token: token}, err
}

func (s *AdminUserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoReply, error) {
	return s.adminUserUsecase.GetUserInfo(ctx, req)
}

func (s *AdminUserService) ListAdminUser(ctx context.Context, req *pb.ListAdminUserRequest) (*pb.ListAdminUserReply, error) {
	return s.adminUserUsecase.ListAdminUser(ctx, req)
}

func (s *AdminUserService) SaveAdminUser(ctx context.Context, req *pb.SaveAdminUserRequest) (*pb.SaveAdminUserReply, error) {
	return s.adminUserUsecase.SaveAdminUser(ctx, req)
}

func (s *AdminUserService) ListOptimizer(ctx context.Context, req *pb.ListOptimizerRequest) (*pb.ListOptimizerReply, error) {
	return s.adminUserUsecase.ListOptimizer(ctx, req)
}

func (s *AdminUserService) ListOperationLog(ctx context.Context, req *pb.ListOperationLogRequest) (*pb.ListOperationLogReply, error) {
	return s.auditUsecase.ListOperationLog(ctx, req)
}
