package valueobject

// ResourceType 资源类型
type ResourceType string

const (
	MenuResourceType   ResourceType = "menu"   // 菜单
	PageResourceType   ResourceType = "page"   // 页面
	ButtonResourceType ResourceType = "button" // 按钮
	APIResourceType    ResourceType = "api"    // 接口
)

func (r ResourceType) String() string {
	return string(r)
}

func (r ResourceType) IsMenu() bool {
	return r == MenuResourceType
}

func (r ResourceType) IsPage() bool {
	return r == PageResourceType
}

func (r ResourceType) IsButton() bool {
	return r == ButtonResourceType
}

func (r ResourceType) IsAPI() bool {
	return r == APIResourceType
}
