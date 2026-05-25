package utils

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"kratos-admin/pkg/toolbox/logx"
)

// 拷贝透传context(防超时处理)
func CopyContext(src context.Context, dest context.Context) context.Context {
	ctx := dest
	if tr, ok := transport.FromServerContext(src); ok {
		ctx = transport.NewServerContext(ctx, tr)
	}
	if md, ok := metadata.FromServerContext(src); ok {
		ctx = metadata.NewServerContext(ctx, md)
	}

	// 从源上下文提取追踪信息并注入到目标上下文中
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(src, carrier)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	return ctx
}

// RunGo 单协程
// @description 使用自定义go，统一拦截panic，使用注意上层传参，如果是引用类型，可能出现传参不达预期的情况
// @param ctx
// @param f 函数
// @param flag
func RunGo(ctx context.Context, f func(ctx context.Context)) {
	newCtx := CopyContext(ctx, context.Background())
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				logx.Errorf(ctx, "RunGo.panic||err=%+v||stack=%s", err, string(debug.Stack()))
			}
		}()
		f(ctx)
	}(newCtx)
}

// ExecutorSync 多协程运行wg封装体(封装了panic和context超时处理)
type ExecutorSync struct {
	wg sync.WaitGroup
}

// NewExecutorSync 创建 sync.WaitGroup 封装体
func NewExecutorSync() *ExecutorSync {
	return &ExecutorSync{
		wg: sync.WaitGroup{},
	}
}

// RunGo 运行一个wg协程
func (e *ExecutorSync) RunGo(ctx context.Context, fn func(ctx context.Context)) {
	e.wg.Add(1)
	newCtx := CopyContext(ctx, context.Background())
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				logx.Errorf(ctx, "RunGo.panic||err=%+v||stack=%s", err, string(debug.Stack()))
			}
			e.wg.Done()
		}()
		fn(ctx)
	}(newCtx)
}

// Wait 等待wg所有协程结束
func (e *ExecutorSync) Wait() {
	e.wg.Wait()
}
