package adminmodel

import (
	"kratos-admin/pkg/toolbox/datax"
)

const TableNameOperationLog = "operation_log"

// OperationLog 审计日志记录表
type OperationLog struct {
	datax.BaseEntity
	OperatorId int64  `gorm:"column:operator_id;type:bigint(20);not null;comment:操作人id" json:"operatorId"`
	Operator   string `gorm:"column:operator;type:varchar(32);not null;comment:操作人" json:"operator"`
	Ip         string `gorm:"column:ip;type:varchar(64);not null;comment:IP地址" json:"ip"`
	Method     string `gorm:"column:method;type:varchar(32);not null;comment:请求方法" json:"method"`
	Url        string `gorm:"column:url;type:varchar(255);not null;comment:请求URL" json:"url"`
	Params     string `gorm:"column:params;type:varchar(1024);comment:请求参数" json:"params"`
	Response   string `gorm:"column:response;type:varchar(1024);comment:响应内容" json:"response"`
}

func (OperationLog) TableName() string {
	return TableNameOperationLog
}
