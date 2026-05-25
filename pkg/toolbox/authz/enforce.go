package authz

import (
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"
)

type Policy struct {
	Sub string
	Dom string
	Obj string
	Act string
}

func NewPolicy(sub, dom, obj, act string) *Policy {
	return &Policy{Sub: sub, Dom: dom, Obj: obj, Act: act}
}

// ================= 基础权限检查 =================

// CheckPermission 检查用户是否有权限
func (s *CasbinServer) CheckPermission(sub, dom, obj, act string) (bool, error) {
	ok, err := s.enforcer.Enforce(sub, dom, obj, act)

	return ok, errorx.WithStack(err)
}

// ================= 策略管理 =================

// AddPolicy 添加直接权限
func (s *CasbinServer) AddPolicy(sub, dom, obj, act string) (bool, error) {
	ok, err := s.enforcer.AddPolicy(sub, dom, obj, act)

	return ok, errorx.WithStack(err)
}

// AddPolicies 批量添加权限
func (s *CasbinServer) AddPolicies(policies []*Policy) (bool, error) {
	if len(policies) == 0 {
		return true, nil
	}

	rules := utils.Map(policies, func(p *Policy) []string {
		return []string{p.Sub, p.Dom, p.Obj, p.Act}
	})

	ok, err := s.enforcer.AddPolicies(rules)

	return ok, errorx.WithStack(err)
}

// RemovePolicy 移除直接权限
func (s *CasbinServer) RemovePolicy(sub, dom, obj, act string) (bool, error) {
	ok, err := s.enforcer.RemovePolicy(sub, dom, obj, act)

	return ok, errorx.WithStack(err)
}

// RemovePolicies 批量移除权限
func (s *CasbinServer) RemovePolicies(policies []*Policy) (bool, error) {
	if len(policies) == 0 {
		return true, nil
	}

	rules := utils.Map(policies, func(p *Policy) []string {
		return []string{p.Sub, p.Dom, p.Obj, p.Act}
	})

	ok, err := s.enforcer.RemovePolicies(rules)

	return ok, errorx.WithStack(err)
}

// 删除某个 subject 的所有策略
func (s *CasbinServer) RemovePoliciesBySubject(sub string) (bool, error) {
	ok, err := s.enforcer.RemoveFilteredPolicy(0, sub)

	return ok, errorx.WithStack(err)
}

// ================= 角色管理 =================

func (s *CasbinServer) AddRoleForUserInDomain(user, role, dom string) (bool, error) {
	ok, err := s.enforcer.AddGroupingPolicy(user, role, dom)

	return ok, errorx.WithStack(err)
}

func (s *CasbinServer) GetRolesForUserInDomain(user, dom string) []string {
	return s.enforcer.GetRolesForUserInDomain(user, dom)
}

func (s *CasbinServer) DeleteRoleForUserInDomain(user, role, dom string) (bool, error) {
	ok, err := s.enforcer.DeleteRoleForUserInDomain(user, role, dom)

	return ok, errorx.WithStack(err)
}

func (s *CasbinServer) GetUsersForRoleInDomain(role, dom string) []string {
	return s.enforcer.GetUsersForRoleInDomain(role, dom)
}

func (s *CasbinServer) GetPermissionsForUserInDomain(user, dom string) []string {
	permissions := s.enforcer.GetPermissionsForUserInDomain(user, dom)

	return utils.Map(permissions, func(item []string) string { return item[2] })
}

func (s *CasbinServer) DeleteUsersForRoleInDomain(role string, users []string, dom string) (bool, error) {
	if len(users) == 0 {
		return true, nil
	}

	rules := utils.Map(users, func(user string) []string {
		return []string{user, role, dom}
	})

	ok, err := s.enforcer.RemoveGroupingPolicies(rules)

	return ok, errorx.WithStack(err)
}

func (s *CasbinServer) AddRolesForUserInDomain(user string, roles []string, dom string) (bool, error) {
	if len(roles) == 0 {
		return true, nil
	}

	rules := utils.Map(roles, func(role string) []string {
		return []string{user, role, dom}
	})

	ok, err := s.enforcer.AddGroupingPolicies(rules)

	return ok, errorx.WithStack(err)
}

func (s *CasbinServer) DeleteRolesForUserInDomain(user, domain string) (bool, error) {
	ok, err := s.enforcer.DeleteRolesForUserInDomain(user, domain)

	return ok, errorx.WithStack(err)
}

func (s *CasbinServer) GetAllUsersByDomain(domain string) ([]string, error) {
	users, err := s.enforcer.GetAllUsersByDomain(domain)

	return users, errorx.WithStack(err)
}

func (s *CasbinServer) AddUsersForRoleInDomain(role string, users []string, dom string) (bool, error) {
	if len(users) == 0 {
		return true, nil
	}

	rules := utils.Map(users, func(user string) []string {
		return []string{user, role, dom}
	})

	ok, err := s.enforcer.AddGroupingPolicies(rules)

	return ok, errorx.WithStack(err)
}
