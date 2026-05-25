package permissionbiz

import (
	"context"
	"sort"
	"sync"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/logx"
	"kratos-admin/pkg/toolbox/utils"

	factory "kratos-admin/internal/biz/permissionbiz/factory"
	"kratos-admin/internal/biz/permissionbiz/valueobject"
	"kratos-admin/internal/data/adminrepo"
	pb "kratos-admin/pb/admin/v1"
	"kratos-admin/pkg/model/adminmodel"
)

type ResourceUsecase struct {
	resourceRepo *adminrepo.ResourceRepo
}

func NewResourceUsecase(resourceRepo *adminrepo.ResourceRepo) *ResourceUsecase {
	return &ResourceUsecase{resourceRepo: resourceRepo}
}

func (uc *ResourceUsecase) ListResource(ctx context.Context) (*pb.ListResourceReply, error) {
	raw, err := uc.resourceRepo.GetAllResource(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListResourceReply{List: factory.BuildResourceInfos(ctx, 0, raw)}, nil
}

func (uc *ResourceUsecase) SaveResource(ctx context.Context, req *pb.SaveResourceRequest) (*pb.SaveResourceReply, error) {
	id, err := uc.saveResource(ctx, req)
	return &pb.SaveResourceReply{Id: id}, err
}

func (uc *ResourceUsecase) DeleteResource(ctx context.Context, id int64) error {
	if id == 0 {
		return pb.ErrorPermissionInvalidParams("该资源不允许删除")
	}

	deleteIds, err := uc.resourceRepo.GetAllChildren(ctx, id)
	if err != nil {
		return err
	}

	deleteIds = append(deleteIds, id)

	return uc.resourceRepo.UpdateResources(ctx, nil, deleteIds)
}

func (uc *ResourceUsecase) ImportResource(ctx context.Context, req *pb.ImportResourceRequest) error {
	return uc.importResource(ctx, req.Data)
}

func (uc *ResourceUsecase) saveResource(ctx context.Context, req *pb.SaveResourceRequest) (int64, error) {
	oldResource, err := uc.resourceRepo.GetResourceByPath(ctx, req.Path)
	if err != nil {
		return 0, err
	}

	if oldResource != nil {
		if req.Id == 0 || (req.Id > 0 && oldResource.Id != req.Id) {
			return 0, pb.ErrorPermissionInvalidParams("资源路径已存在")
		}
	}

	newResource := factory.NewResourceModel(ctx, req)

	siblings, err := uc.resourceRepo.GetResourcesByParentId(ctx, req.ParentId)
	if err != nil {
		return 0, err
	}

	siblings = utils.Filter(siblings, func(in *adminmodel.Resource) bool { return in.Id != req.Id })
	siblings = append(siblings, newResource)

	sortResourceSiblings(siblings)

	if err := uc.resourceRepo.UpdateResources(ctx, siblings, nil); err != nil {
		return 0, err
	}

	return newResource.Id, nil
}

func (uc *ResourceUsecase) importResource(ctx context.Context, data []*pb.ResourceInfo) error {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		globalErr error
		weight    int32 = 1000
	)

	if err := validateImportResources(data, nil); err != nil {
		return err
	}

	for _, item := range data {
		item.Weight = weight
		weight--

		resource, err := uc.resourceRepo.GetResourceByPath(ctx, item.Path)
		if err != nil {
			return err
		}

		if resource != nil {
			item.Id = resource.Id
		}

		id, err := uc.saveResource(ctx, utils.CopyPtr[pb.SaveResourceRequest](item))
		if err != nil {
			return errorx.WithStack(err)
		}

		wg.Add(1)

		utils.RunGo(ctx, func(ctx context.Context) {
			defer wg.Done()

			if err := uc.batchImportResource(ctx, id, item.Children); err != nil {
				logx.Errorf(ctx, "batchImportResource failed: %+v", err)
				mu.Lock()
				globalErr = err
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	return globalErr
}

func validateImportResources(resources []*pb.ResourceInfo, visited map[string]bool) error {
	if len(resources) == 0 {
		return nil
	}

	if len(visited) == 0 {
		visited = make(map[string]bool)
	}

	for _, item := range resources {
		if visited[item.Path] {
			return pb.ErrorPermissionInvalidParams("资源路径重复: %s", item.Path)
		}

		visited[item.Path] = true

		if len(item.Children) == 0 {
			continue
		}

		if err := validateImportResources(item.Children, visited); err != nil {
			return err
		}
	}

	return nil
}

func (uc *ResourceUsecase) batchImportResource(ctx context.Context, parentId int64, resources []*pb.ResourceInfo) error {
	if len(resources) == 0 {
		return nil
	}

	data := utils.Distinct(resources, func(item *pb.ResourceInfo) string { return item.Path })
	if len(data) != len(resources) {
		return errorx.New("资源路径重复: %v", utils.Map(resources, func(item *pb.ResourceInfo) string { return item.Path }))
	}

	childrenIds, err := uc.resourceRepo.GetAllChildren(ctx, parentId)
	if err != nil {
		return errorx.WithStack(err)
	}

	newResources := utils.Map(data, func(item *pb.ResourceInfo) *adminmodel.Resource {
		item.ParentId = parentId
		return factory.NewResourceModel(ctx, utils.CopyPtr[pb.SaveResourceRequest](item))
	})

	assignResourceWeights(newResources)

	if err := uc.resourceRepo.UpdateResources(ctx, newResources, childrenIds); err != nil {
		return errorx.WithStack(err)
	}

	for _, item := range data {
		if len(item.Children) == 0 {
			continue
		}

		parent, ok := utils.FindFirst(newResources, func(v *adminmodel.Resource) bool { return v.Path == item.Path })
		if !ok {
			return errorx.New("资源路径(%s)保存失败", item.Path)
		}

		if err := uc.batchImportResource(ctx, parent.Id, item.Children); err != nil {
			return errorx.WithStack(err)
		}
	}

	return nil
}

func sortResourceSiblings(data []*adminmodel.Resource) {
	sort.Slice(data, func(i, j int) bool {
		if data[i].Weight == data[j].Weight {
			if data[i].CreateTime.IsZero() {
				return true
			}
		}

		return data[i].Weight > data[j].Weight
	})

	assignResourceWeights(data)
}

func assignResourceWeights(data []*adminmodel.Resource) {
	var weight int32 = 1000

	for _, v := range data {
		if valueobject.ResourceType(v.Type).IsAPI() {
			v.Weight = 0
			v.ParentId = 0
			continue
		}

		v.Weight = weight
		weight--
	}
}
