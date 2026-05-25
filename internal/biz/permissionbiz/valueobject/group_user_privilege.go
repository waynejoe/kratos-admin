package valueobject

type PrivilegeLevel int32

const (
	MemberPrivilegeLevel PrivilegeLevel = 1 // 普通成员, 仅查看个人数据
	LeaderPrivilegeLevel PrivilegeLevel = 2 // 组长, 可查看团队数据
)

// GroupUserPrivilege 团队用户权限
type GroupUserPrivilege struct {
	UserId         int64          `json:"userId"`
	PrivilegeLevel PrivilegeLevel `json:"privilegeLevel"`
}

// NewGroupUserPrivilege 创建团队用户权限
func NewGroupUserPrivilege(userId int64, privilegeLevel int32) *GroupUserPrivilege {
	return &GroupUserPrivilege{
		UserId:         userId,
		PrivilegeLevel: PrivilegeLevel(privilegeLevel),
	}
}

// IsLeader 判断权限是否为组长
func (p PrivilegeLevel) IsLeader() bool {
	return p == LeaderPrivilegeLevel
}
