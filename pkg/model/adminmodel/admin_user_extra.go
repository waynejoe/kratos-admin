package adminmodel

import (
	"encoding/json"

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
