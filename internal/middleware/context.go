package middleware

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/metadata"

	"kratos-admin/pkg/toolbox/logx"
)

const dataPermissionKey = "user.dataPermission"

func SetDataPermission(ctx context.Context, data []int64) {
	if len(data) == 0 {
		return
	}
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		logx.Warnf(ctx, "SetDataPermission failed, metadata not found")
		return
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		logx.Errorf(ctx, "SetDataPermission failed, marshal data failed, err: %v", err)
		return
	}
	md.Set(dataPermissionKey, string(dataStr))
}

func GetDataPermission(ctx context.Context) []int64 {
	var res []int64
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		return res
	}
	data := md.Get(dataPermissionKey)
	if data == "" {
		return res
	}
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		logx.Errorf(ctx, "GetDataPermission failed, unmarshal data failed, err: %v", err)
	}
	return res
}
