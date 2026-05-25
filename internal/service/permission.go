package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"

	"kratos-admin/internal/biz/permissionbiz"
	pb "kratos-admin/pb/admin/v1"
)

type PermissionService struct {
	pb.UnimplementedPermissionServer

	adminGroupUsecase *permissionbiz.AdminGroupUsecase
	adminRoleUsecase  *permissionbiz.AdminRoleUsecase
	resourceUsecase   *permissionbiz.ResourceUsecase
}

func NewPermissionService(
	adminGroupUsecase *permissionbiz.AdminGroupUsecase,
	adminRoleUsecase *permissionbiz.AdminRoleUsecase,
	resourceUsecase *permissionbiz.ResourceUsecase,
) *PermissionService {
	return &PermissionService{
		adminGroupUsecase: adminGroupUsecase,
		adminRoleUsecase:  adminRoleUsecase,
		resourceUsecase:   resourceUsecase,
	}
}

func (s *PermissionService) InitService(srv *http.Server) error {
	pb.RegisterPermissionHTTPServer(srv, s)
	return nil
}

func (s *PermissionService) UnInitService() error {
	return nil
}

func (s *PermissionService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	return s.adminGroupUsecase.ListGroup(ctx, req)
}

func (s *PermissionService) SaveGroup(ctx context.Context, req *pb.SaveGroupRequest) (*pb.SaveGroupReply, error) {
	return s.adminGroupUsecase.SaveGroup(ctx, req)
}

func (s *PermissionService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.adminGroupUsecase.DeleteGroup(ctx, req.Id)
}

func (s *PermissionService) ListGroupMember(ctx context.Context, req *pb.ListGroupMemberRequest) (*pb.ListGroupMemberReply, error) {
	return s.adminGroupUsecase.ListGroupMember(ctx, req)
}

func (s *PermissionService) SaveGroupMember(ctx context.Context, req *pb.SaveGroupMemberRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.adminGroupUsecase.SaveGroupMember(ctx, req)
}

func (s *PermissionService) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	return s.adminRoleUsecase.ListRole(ctx, req)
}

func (s *PermissionService) SaveRole(ctx context.Context, req *pb.SaveRoleRequest) (*pb.SaveRoleReply, error) {
	return s.adminRoleUsecase.SaveRole(ctx, req)
}

func (s *PermissionService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.adminRoleUsecase.DeleteRole(ctx, req.Id)
}

func (s *PermissionService) ListRolePermission(ctx context.Context, req *pb.ListRolePermissionRequest) (*pb.ListRolePermissionReply, error) {
	return s.adminRoleUsecase.ListRolePermission(ctx, req)
}

func (s *PermissionService) SaveRolePermission(ctx context.Context, req *pb.SaveRolePermissionRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.adminRoleUsecase.SaveRolePermission(ctx, req)
}

func (s *PermissionService) ListRoleMember(ctx context.Context, req *pb.ListRoleMemberRequest) (*pb.ListRoleMemberReply, error) {
	return s.adminRoleUsecase.ListRoleMember(ctx, req)
}

func (s *PermissionService) ListResource(ctx context.Context, req *pb.ListResourceRequest) (*pb.ListResourceReply, error) {
	return s.resourceUsecase.ListResource(ctx)
}

func (s *PermissionService) SaveResource(ctx context.Context, req *pb.SaveResourceRequest) (*pb.SaveResourceReply, error) {
	return s.resourceUsecase.SaveResource(ctx, req)
}

func (s *PermissionService) DeleteResource(ctx context.Context, req *pb.DeleteResourceRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.resourceUsecase.DeleteResource(ctx, req.Id)
}

func (s *PermissionService) ImportResource(ctx context.Context, req *pb.ImportResourceRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, s.resourceUsecase.ImportResource(ctx, req)
}
