package utils

import (
	"encoding/json"
	"reflect"

	"kratos-admin/pkg/toolbox/errorx"
)

func ToStruct[T any](m any) (*T, error) {
	bs, err := json.Marshal(m)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to marshal")
	}
	var entity T
	err = json.Unmarshal(bs, &entity)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to unmarshal json")
	}
	return &entity, nil
}

// MergeStructs 合并两个相同类型的结构体，保留非零字段，优先使用a中的值
func MergeStructs(new, old interface{}) (interface{}, error) {
	valA := reflect.Indirect(reflect.ValueOf(new))
	valB := reflect.Indirect(reflect.ValueOf(old))

	if valA.Type() != valB.Type() {
		return nil, errorx.New("structs have different types")
	}

	result := reflect.New(valA.Type()).Elem()

	for i := 0; i < valA.NumField(); i++ {
		fieldType := valA.Type().Field(i)
		if !fieldType.IsExported() {
			continue // 跳过不可导出的字段
		}

		fieldA := valA.Field(i)
		fieldB := valB.Field(i)
		resultField := result.Field(i)

		if !fieldA.IsZero() {
			resultField.Set(fieldA)
		} else if !fieldB.IsZero() {
			resultField.Set(fieldB)
		}
	}

	return result.Interface(), nil
}
