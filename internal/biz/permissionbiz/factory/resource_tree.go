package factory

import (
	"context"
	"sort"

	"kratos-admin/pkg/toolbox/logx"
	"kratos-admin/pkg/toolbox/utils"

	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

var defaultAPIResource = &pb.ResourceInfo{Name: "API资源【内部使用】", Type: "api", Path: "/api"}

// BuildResourceInfos 将扁平资源列表组装为树形 pb 视图
func BuildResourceInfos(ctx context.Context, parentID int64, data []*adminmodel.Resource) []*pb.ResourceInfo {
	var (
		resourceList = make([]*pb.ResourceInfo, 0)
		apiResource  = utils.CopyPtr[pb.ResourceInfo](defaultAPIResource)
		children     = utils.Filter(data, func(item *adminmodel.Resource) bool { return item.ParentId == parentID })
	)

	for _, item := range children {
		resource := utils.CopyPtr[pb.ResourceInfo](item)

		apis, err := utils.UnmarshalList[string](item.Apis)
		if err != nil {
			logx.Errorf(ctx, "unmarshal apis failed: %+v", err)
		}

		resource.Apis = apis

		if item.Type == "api" {
			apiResource.Children = append(apiResource.Children, resource)
		} else {
			resource.Children = BuildResourceInfos(ctx, item.Id, data)
			resourceList = append(resourceList, resource)
		}
	}

	sort.Slice(resourceList, func(i, j int) bool {
		return resourceList[i].Weight > resourceList[j].Weight
	})

	if len(apiResource.Children) > 0 {
		resourceList = append(resourceList, apiResource)
	}

	return resourceList
}
