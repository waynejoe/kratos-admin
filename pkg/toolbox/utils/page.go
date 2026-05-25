package utils

// Page 分页
type Page struct {
	pageIndex int32
	pageSize  int32
}

// NewPage 新建分页
func NewPage(pageIndex, pageSize int32) *Page {
	page := &Page{pageIndex: pageIndex, pageSize: pageSize}

	page.setDefault()

	return page
}

// GetOffset 获取偏移量
func (p *Page) GetOffset() int {
	p.setDefault()

	return int((p.pageIndex - 1) * p.pageSize)
}

// GetLimit 获取限制数量
func (p *Page) GetLimit() int {
	p.setDefault()

	return int(p.pageSize)
}

// setDefault 设置默认值
func (p *Page) setDefault() {
	if p.pageIndex < 1 {
		p.pageIndex = 1
	}

	if p.pageSize < 1 {
		p.pageSize = 2500
	}
}

// GetFrom 获取从第几个开始
func (p *Page) GetFrom() int {
	return p.GetOffset()
}

// GetSize 获取数量
func (p *Page) GetSize() int {
	return p.GetLimit()
}
