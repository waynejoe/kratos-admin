package adminmodel

// 软删除
const (
	NotDel = 0
)

// 团队 / 角色启用状态
const (
	StatusOff int32 = 1
	StatusOn  int32 = 2
)

// 后台账号状态
const (
	UserStatusNormal = 1
	UserStatusLogout = 2
	UserStatusBan    = 3
)
