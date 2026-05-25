package valueobject

// GroupDataIsolation 分组数据隔离级别
type GroupDataIsolation int32

const (
	CloseDataIsolation GroupDataIsolation = 1
	OpenDataIsolation  GroupDataIsolation = 2
)

// IsClose 关闭数据隔离
func (g GroupDataIsolation) IsClose() bool {
	return g == CloseDataIsolation
}

// IsOpen 开启数据隔离
func (g GroupDataIsolation) IsOpen() bool {
	return g == OpenDataIsolation
}
