package utils

import (
	"cmp"
	"slices"
)

func Filter[T any](list []T, f func(T) bool) []T {
	if len(list) == 0 {
		return list
	}
	res := make([]T, 0)
	for _, item := range list {
		if f(item) {
			res = append(res, item)
		}
	}
	return res
}

func Cut[T any](list []T, offset, limit int32) []T {
	if offset < 0 || limit <= 0 || len(list) <= int(offset) {
		return make([]T, 0)
	}
	return list[offset:min(len(list), int(offset+limit))]
}

func CutPage[T any](list []T, pageIndex, pageSize int32) []T {
	offset := (pageIndex - 1) * pageSize
	return Cut(list, offset, pageSize)
}

func Map[K, V any](list []K, mapFunc func(K) V) []V {
	res := make([]V, 0)
	for _, item := range list {
		v := mapFunc(item)
		res = append(res, v)
	}
	return res
}

// MapWithErr 将切片转换为另一个切片，支持错误处理
func MapWithErr[K, V any](list []K, mapFunc func(K) (V, error)) ([]V, error) {
	res := make([]V, 0)

	for _, item := range list {
		v, err := mapFunc(item)
		if err != nil {
			return nil, err
		}

		res = append(res, v)
	}

	return res, nil
}

// ToHashMap 将切片转换为 map
func ToHashMap[K comparable, V any](list []V, mapFunc func(V) K) map[K]V {
	res := make(map[K]V)

	for _, item := range list {
		k := mapFunc(item)

		res[k] = item
	}

	return res
}

// ForEach 遍历切片
func ForEach[T any](list []T, f func(T)) {
	for _, item := range list {
		f(item)
	}
}

func FindFirst[T any](list []T, f func(T) bool) (T, bool) {
	for _, item := range list {
		if f(item) {
			return item, true
		}
	}
	return *new(T), false
}

func Any[T any](list []T, f func(T) bool) bool {
	for _, item := range list {
		if f(item) {
			return true
		}
	}
	return false
}

func All[T any](list []T, f func(T) bool) bool {
	for _, item := range list {
		if !f(item) {
			return false
		}
	}
	return true
}

func Distinct[T any, K comparable](elements []T, keySelector func(item T) K) []T {
	var (
		seenKeys      = make(map[K]bool)
		distinctItems = make([]T, 0, len(elements))
	)

	for _, item := range elements {
		key := keySelector(item)
		if !seenKeys[key] {
			seenKeys[key] = true

			distinctItems = append(distinctItems, item)
		}
	}

	return distinctItems
}

func SortDesc[T any, K cmp.Ordered](list []T, fieldExtractor func(entity T) []K) {
	slices.SortStableFunc(list, func(a, b T) int {
		fieldsA := fieldExtractor(a)
		fieldsB := fieldExtractor(b)
		if len(fieldsA) != len(fieldsB) {
			return 0
		}
		for i := 0; i < len(fieldsA); i++ {
			// 直接使用cmp.Compare，然后取反实现降序
			if result := cmp.Compare(fieldsA[i], fieldsB[i]); result != 0 {
				return -result // 取反实现降序
			}
		}
		return 0
	})
}
