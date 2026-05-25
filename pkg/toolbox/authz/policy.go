package authz

const rbacModels = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

const TableNameCasbinRule = "casbin_rule"

// CasbinRule 自定义策略表结构
type CasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:256"`
	V0    string `gorm:"size:256"`
	V1    string `gorm:"size:256"`
	V2    string `gorm:"size:256"`
	V3    string `gorm:"size:256"`
	V4    string `gorm:"size:256"`
	V5    string `gorm:"size:256"`
}

func (CasbinRule) TableName() string {
	return TableNameCasbinRule
}
