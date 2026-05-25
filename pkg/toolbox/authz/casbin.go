package authz

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/logx"
)

type ICasbinServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	CheckPermission(sub, dom, obj, act string) (bool, error)
	AddPolicy(sub, dom, obj, act string) (bool, error)
	AddPolicies(policies []*Policy) (bool, error)
	RemovePolicy(sub, dom, obj, act string) (bool, error)
	RemovePolicies(policies []*Policy) (bool, error)
	RemovePoliciesBySubject(sub string) (bool, error)
	AddRoleForUserInDomain(user, role, domain string) (bool, error)
	AddRolesForUserInDomain(user string, roles []string, dom string) (bool, error)
	DeleteRoleForUserInDomain(user, role, domain string) (bool, error)
	DeleteRolesForUserInDomain(user, domain string) (bool, error)
	GetRolesForUserInDomain(user, domain string) []string
	GetUsersForRoleInDomain(role, domain string) []string
	DeleteUsersForRoleInDomain(role string, users []string, dom string) (bool, error)
	GetPermissionsForUserInDomain(user, dom string) []string
	GetAllUsersByDomain(domain string) ([]string, error)
	AddUsersForRoleInDomain(role string, users []string, dom string) (bool, error)
}

var _ ICasbinServer = (*CasbinServer)(nil)

// CasbinServer 封装 Casbin 权限管理
type CasbinServer struct {
	enforcer *casbin.SyncedEnforcer
	watcher  persist.Watcher
	db       *gorm.DB
}

// NewCasbinServer 创建 Casbin 服务实例
func NewCasbinServer(db *gorm.DB, redisOpts *redis.Options) (ICasbinServer, error) {
	// 1. 初始化 model + adapter
	m, err := model.NewModelFromString(rbacModels)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &CasbinRule{}, TableNameCasbinRule)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	// 2. 初始化 enforcer
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	// 3. 初始化 watcher
	updateFn := func(msg string) {
		log.Infof("Policy update received from watcher: %s", msg)
		if err := enforcer.LoadPolicy(); err != nil {
			log.Errorf("Casbin reload failed: %v", err)
		} else {
			log.Info("Casbin policies reloaded successfully.")
		}
	}
	watcher, err := newRedisWatcher(redisOpts, updateFn)
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	if err := enforcer.SetWatcher(watcher); err != nil {
		return nil, errorx.WithStack(err)
	}

	// 4. 启用自动保存
	enforcer.EnableAutoSave(true)

	return &CasbinServer{
		enforcer: enforcer,
		watcher:  watcher,
		db:       db,
	}, nil
}

func newRedisWatcher(redisOpts *redis.Options, updateFn func(msg string)) (persist.Watcher, error) {
	watcher, err := rediswatcher.NewWatcher(redisOpts.Addr, rediswatcher.WatcherOptions{
		Options: *redisOpts,
		Channel: "/casbin-rbac",
	})
	if err != nil {
		return nil, errorx.WithStack(err)
	}

	if err := watcher.SetUpdateCallback(updateFn); err != nil {
		return nil, errorx.WithStack(err)
	}

	return watcher, nil
}

func (s *CasbinServer) Start(ctx context.Context) error {
	logx.Info(ctx, "Starting Casbin RBAC Server...")

	// 加载策略
	if err := s.enforcer.LoadPolicy(); err != nil {
		return errorx.WithStack(err)
	}
	policys, err := s.enforcer.GetPolicy()
	if err != nil {
		return errorx.WithStack(err)
	}

	logx.Infof(ctx, "Casbin initialized with %d policies", len(policys))

	return nil
}

func (s *CasbinServer) Stop(ctx context.Context) error {
	if s.watcher != nil {
		s.watcher.Close()
	}
	logx.Info(ctx, "Casbin server stopped")
	return nil
}
