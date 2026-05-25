package adminrepo

import (
	"context"
	"fmt"
	"time"

	"kratos-admin/pkg/toolbox/datax"
	"kratos-admin/pkg/toolbox/utils"

	"kratos-admin/internal/data"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type OperationLogRepo struct {
	datax.Repo[adminmodel.OperationLog]
}

func NewOperationLogRepo(data *data.Data) *OperationLogRepo {
	return &OperationLogRepo{
		Repo: datax.NewRepo[adminmodel.OperationLog](data.G.AdminDB),
	}
}

func (a *OperationLogRepo) GetOperationLogList(ctx context.Context, req *pb.ListOperationLogRequest) ([]*adminmodel.OperationLog, int64, error) {
	var (
		res   []*adminmodel.OperationLog
		total int64
		page  = utils.NewPage(req.PageIndex, req.PageSize)
	)
	db := a.DB(ctx)
	if req.Url != "" {
		db = db.Where("url like ?", fmt.Sprintf("%%%s%%", req.Url))
	}
	if req.Method != "" {
		db = db.Where("method = ?", req.Method)
	}
	if req.StartTime != 0 {
		startTime := time.UnixMilli(req.StartTime).Format(time.DateTime)
		db = db.Where("create_time >= ?", startTime)
	}
	if req.EndTime != 0 {
		endTime := time.UnixMilli(req.EndTime).Format(time.DateTime)
		db = db.Where("create_time <= ?", endTime)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("create_time DESC").Limit(page.GetLimit()).Offset(page.GetOffset()).Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, total, nil
}
