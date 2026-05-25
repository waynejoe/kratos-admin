package datax

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm/schema"

	"kratos-admin/pkg/toolbox/utils"
)

var (
	omitColumns = []string{"update_time", "create_time"}
)

type Entity interface {
	TableName() string
	PKVal() int64
}

type BaseEntity struct {
	Id         int64     `gorm:"primaryKey;column:id" json:"id,omitempty"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
}

func (e BaseEntity) PKVal() int64 {
	return e.Id
}

func GetIds[T Entity](list []*T) []int64 {
	return utils.Map(list, func(e *T) int64 {
		return any(e).(Entity).PKVal()
	})
}

func JoinIds[T Entity](sep string, list []*T) string {
	return utils.JoinIds(sep, GetIds(list))
}

func GetSerialVersion(model any) string {
	s, _ := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	fields := make([]string, 0)
	for _, f := range s.Fields {
		tag := f.Tag.Get("json")
		tag = strings.Split(tag, ",")[0]
		fsign := fmt.Sprintf("%s.%s.%s", f.DBName, f.FieldType.String(), tag)
		fields = append(fields, fsign)
	}
	slices.Sort(fields)
	hash := md5.New()
	hash.Write([]byte(strings.Join(fields, ",")))
	h := hex.EncodeToString(hash.Sum(nil))
	return h[:6]
}
