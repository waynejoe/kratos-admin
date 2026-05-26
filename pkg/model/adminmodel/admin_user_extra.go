package adminmodel

import (
	"encoding/json"

	"gorm.io/gorm"

	"kratos-admin/pkg/toolbox/utils"
)

// AdminUserExtra 管理用户额外信息
type AdminUserExtra struct {
	OpenId  string `json:"openId"`
	UnionId string `json:"unionId"`
}

func NewAdminUserExtraFromModel(data string) *AdminUserExtra {
	if data == "" {
		return &AdminUserExtra{}
	}

	extra, _ := utils.Unmarshal[AdminUserExtra](data)

	return extra
}

func (a *AdminUserExtra) ToJson() string {
	if a == nil {
		return "{}"
	}

	data, _ := json.Marshal(a)
	return string(data)
}

// BeforeSave 保证 JSON 列写入合法值（MySQL 不接受空字符串）
func (u *AdminUser) BeforeSave(_ *gorm.DB) error {
	return u.normalizeExtra()
}

// BeforeUpdate 与 BeforeSave 相同；CacheRepo.Update 走 Updates，只会触发 BeforeUpdate
func (u *AdminUser) BeforeUpdate(_ *gorm.DB) error {
	return u.normalizeExtra()
}

func (u *AdminUser) normalizeExtra() error {
	if u.Extra == "" {
		u.Extra = "{}"
	}
	return nil
}
