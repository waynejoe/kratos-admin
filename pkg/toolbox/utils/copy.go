package utils

import (
	"github.com/jinzhu/copier"
	"kratos-admin/pkg/toolbox/errorx"
)

var (
	copyOption = copier.Option{DeepCopy: true}
)

func RawCopy(dst, src any) error {
	err := copier.CopyWithOption(dst, src, copyOption)
	if err != nil {
		return errorx.Wrap(err, "failed to copy dst %+v, src %+v", dst, src)
	}
	return nil
}

func Copy[T any](src any) T {
	var dst T
	err := copier.CopyWithOption(&dst, src, copyOption)
	if err != nil {
		panic(err)
	}
	return dst
}

func CopyPtr[T any](src any) *T {
	res := Copy[T](src)
	return &res
}

func CopySlice[T any](src any) []T {
	var dst = make([]T, 0)
	err := copier.CopyWithOption(&dst, src, copyOption)
	if err != nil {
		panic(err)
	}
	return dst
}
