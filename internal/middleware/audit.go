package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	http2 "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/helpx"
	"kratos-admin/pkg/toolbox/logx"

	"kratos-admin/internal/data/adminrepo"
	"kratos-admin/pkg/model/adminmodel"
)

var protoMarshalOptions = protojson.MarshalOptions{EmitUnpopulated: true}

type standardReply struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func AuditMiddleware(auditRepo *adminrepo.OperationLogRepo, adminUserRepo *adminrepo.AdminUserRepo) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(*http2.Transport); ok {
					method := ht.Request().Method
					if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete && method != http.MethodPatch {
						return handler(ctx, req)
					}

					reply, err := handler(ctx, req)
					if err != nil {
						return reply, err
					}

					stdReply := standardReply{
						Code:    0,
						Data:    extractArgs(reply),
						Message: "success",
					}

					go func() {
						newCtx := context.WithoutCancel(ctx)
						userId := helpx.GetUserId(newCtx)
						userName := ""
						userInfo, _ := adminUserRepo.Get(newCtx, userId)
						if userInfo != nil {
							userName = userInfo.Nickname
						}
						if err := auditRepo.Save(context.Background(), &adminmodel.OperationLog{
							Ip:         helpx.GetIP(newCtx),
							Method:     method,
							Url:        ht.Request().URL.Path,
							Params:     extractArgs(req),
							Response:   extractArgs(stdReply),
							OperatorId: userId,
							Operator:   userName,
						}); err != nil {
							logx.Errorf(newCtx, "failed to save audit log: %v", errorx.WithStack(err))
						}
					}()

					return reply, err
				}
			}
			return handler(ctx, req)
		}
	}
}

func extractArgs(data any) string {
	if p, ok := data.(proto.Message); ok {
		bs, _ := protoMarshalOptions.Marshal(p)
		return string(bs)
	}
	bs, _ := json.Marshal(data)
	return string(bs)
}
