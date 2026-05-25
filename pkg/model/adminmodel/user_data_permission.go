package adminmodel

// UserDataPermission 用户数据权限
type UserDataPermission struct {
	HasDataPermission bool
	UserIds           []int64
}
