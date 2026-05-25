package authz

import (
	"context"
	"strconv"
	"strings"

	toolboxauthz "kratos-admin/pkg/toolbox/authz"
	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/utils"
)

// Store Casbin 授权存储（团队/角色/用户关系与策略）
type Store struct {
	*toolboxauthz.CasbinServer
}

func NewStore(server *toolboxauthz.CasbinServer) *Store {
	return &Store{CasbinServer: server}
}

func (s *Store) GetUserPermissions(_ context.Context, userID int64) []string {
	return s.GetPermissionsForUserInDomain(strconv.FormatInt(userID, 10), Domain)
}

func (s *Store) UpdateUserPermissions(_ context.Context, userID int64, permissions []string) error {
	user := strconv.FormatInt(userID, 10)
	if _, err := s.DeleteRolesForUserInDomain(user, Domain); err != nil {
		return err
	}
	_, err := s.AddRolesForUserInDomain(user, permissions, Domain)
	return err
}

func (s *Store) GetUsersForGroup(groupSub string) []int64 {
	return parseUserIDs(s.GetUsersForRoleInDomain(groupSub, Domain))
}

func (s *Store) GetUsersForRole(roleSub string) []int64 {
	return parseUserIDs(s.GetUsersForRoleInDomain(roleSub, Domain))
}

func (s *Store) DeleteGroup(ctx context.Context, groupSub string) error {
	users := s.GetUsersForRoleInDomain(groupSub, Domain)
	_, err := s.DeleteUsersForRoleInDomain(groupSub, users, Domain)
	return err
}

func (s *Store) UpdateUsersGroup(_ context.Context, newGroupSub, oldGroupSub string, userIDs []int64) error {
	users := utils.Map(userIDs, func(id int64) string { return strconv.FormatInt(id, 10) })
	if _, err := s.DeleteUsersForRoleInDomain(oldGroupSub, users, Domain); err != nil {
		return errorx.WithStack(err)
	}
	_, err := s.AddUsersForRoleInDomain(newGroupSub, users, Domain)
	return errorx.WithStack(err)
}

func (s *Store) GetUserGroups(_ context.Context, userIDs ...int64) (map[int64][]int64, error) {
	return s.parseEntityIDsForUsers(userIDs, GroupPermissionPrefix)
}

func (s *Store) GetUserRoles(_ context.Context, userIDs ...int64) (map[int64][]int64, error) {
	return s.parseEntityIDsForUsers(userIDs, RolePermissionPrefix)
}

func (s *Store) UpdateRolePermissions(_ context.Context, newData, deleteData []*toolboxauthz.Policy) error {
	if _, err := s.RemovePolicies(deleteData); err != nil {
		return err
	}
	_, err := s.AddPolicies(newData)
	return err
}

func (s *Store) DeleteRole(_ context.Context, roleSub string) error {
	users := s.GetUsersForRoleInDomain(roleSub, Domain)
	resources := s.GetPermissionsForUserInDomain(roleSub, Domain)
	if _, err := s.DeleteUsersForRoleInDomain(roleSub, users, Domain); err != nil {
		return err
	}
	_, err := s.RemovePolicies(utils.Map(resources, func(obj string) *toolboxauthz.Policy {
		return toolboxauthz.NewPolicy(roleSub, Domain, obj, Act)
	}))
	return err
}

func (s *Store) GetRolePermissions(roleSub string) []string {
	return s.GetPermissionsForUserInDomain(roleSub, Domain)
}

func (s *Store) CheckAPIPermission(userID int64, path string) (bool, error) {
	return s.CheckPermission(strconv.FormatInt(userID, 10), Domain, "api:"+path, Act)
}

func (s *Store) parseEntityIDsForUsers(userIDs []int64, prefix string) (map[int64][]int64, error) {
	out := make(map[int64][]int64)
	var entityIDs []int64
	for _, userID := range userIDs {
		roles := s.GetRolesForUserInDomain(strconv.FormatInt(userID, 10), Domain)
		for _, role := range roles {
			if !strings.HasPrefix(role, prefix) {
				continue
			}
			id, err := strconv.ParseInt(strings.TrimPrefix(role, prefix), 10, 64)
			if err != nil {
				continue
			}
			entityIDs = append(entityIDs, id)
			out[userID] = append(out[userID], id)
		}
	}
	return out, nil
}

func parseUserIDs(users []string) []int64 {
	return utils.Map(users, func(in string) int64 {
		id, _ := strconv.Atoi(in)
		return int64(id)
	})
}
