package datax

import (
	"reflect"

	"kratos-admin/pkg/toolbox/utils"
)

func isNil(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}

	return false
}

func New[T Entity](src any) *T {
	var entity T
	err := utils.RawCopy(&entity, src)
	if err != nil {
		panic(err)
	}
	return &entity
}
