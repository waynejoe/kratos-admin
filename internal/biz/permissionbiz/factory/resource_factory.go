package factory

import (
	"context"
	"fmt"

	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/utils"
)

// newPermissionObject 新权资源值
func newPermissionObject(in *pb.ResourceInfo) string {
	return fmt.Sprintf("%s:%s", in.Type, in.Path)
}

// NewResourceModel 创建资源模型
func NewResourceModel(ctx context.Context, req *pb.SaveResourceRequest) *adminmodel.Resource {
	data := utils.CopyPtr[adminmodel.Resource](req)

	data.OperatorId = helpx.GetUserId(ctx)

	data.Apis, _ = utils.MarshalList(utils.Filter(req.Apis, func(s string) bool { return s != "" }))

	return data
}

// NewResourceFromModel 从模型创建资源
func NewResourceFromModel(req *adminmodel.Resource) (*pb.ResourceInfo, error) {
	data := utils.CopyPtr[pb.ResourceInfo](req)

	apis, err := utils.UnmarshalList[string](req.Apis)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	data.Apis = apis

	return data, nil
}
