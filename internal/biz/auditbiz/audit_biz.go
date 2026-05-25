package auditbiz

import (
	"context"

	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
)

type AuditUsecase struct {
	operationLogRepo *adminrepo.OperationLogRepo
}

func NewAuditUsecase(operationLogRepo *adminrepo.OperationLogRepo) *AuditUsecase {
	return &AuditUsecase{operationLogRepo: operationLogRepo}
}

func (uc *AuditUsecase) ListOperationLog(ctx context.Context, req *pb.ListOperationLogRequest) (*pb.ListOperationLogReply, error) {
	res, total, err := uc.operationLogRepo.GetOperationLogList(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.ListOperationLogReply{
		List: utils.CopySlice[*pb.OperationLog](res),
		Page: &pb.PageInfo{
			Total: int32(total),
			Index: req.GetPageIndex(),
			Size:  req.GetPageSize(),
			IsEnd: total <= int64(req.GetPageIndex()*req.GetPageSize()),
		},
	}, nil
}
