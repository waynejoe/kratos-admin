package factory

import (
	"fmt"
	"strings"

	"kratos-admin/pkg/model/adminmodel"
	"kratos-admin/pkg/toolbox/utils"
)

// 历史 Casbin 策略（page:/system/permission/*）与当前资源 path（menu:team_management）不一致时的别名
var legacyPermissionAliases = map[string][]string{
	"menu:/system":                      {"menu:permission"},
	"menu:/system/permission":           {"menu:permission"},
	"page:/system/permission/team":      {"menu:team_management"},
	"page:/system/permission/role":      {"menu:role_management"},
	"page:/system/permission/user":      {"menu:user_management"},
	"page:/system/permission/menu":      {"menu:menu_management"},
}

// NormalizeUserPermissions 展开旧版策略键，并在检测到旧版超管页面权限时补齐全部资源策略
func NormalizeUserPermissions(perms []string, resources []*adminmodel.Resource) []string {
	set := make(map[string]struct{}, len(perms)+32)
	for _, p := range perms {
		set[p] = struct{}{}
		for _, alias := range legacyPermissionAliases[p] {
			set[alias] = struct{}{}
		}
	}

	if hasLegacySuperAdminPages(perms) {
		for _, r := range resources {
			set[fmt.Sprintf("%s:%s", r.Type, r.Path)] = struct{}{}
			apis, _ := utils.UnmarshalList[string](r.Apis)
			for _, api := range apis {
				if api == "" {
					continue
				}
				set[fmt.Sprintf("api:%s", api)] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	return out
}

func hasLegacySuperAdminPages(perms []string) bool {
	for _, p := range perms {
		if strings.HasPrefix(p, "page:/system/permission/") {
			return true
		}
	}
	return false
}
