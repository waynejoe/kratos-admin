package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kratos-admin/pkg/toolbox/errorx"
)

func CutIds(sep string, args ...any) []int64 {
	ids := make([]int64, 0)
	for _, arg := range args {
		if id, ok := arg.(int64); ok {
			ids = append(ids, id)
		} else if idStr, ok := arg.(string); ok {
			ids = append(ids, CutIdStr(idStr, sep)...)
		} else if id, ok := arg.([]int64); ok {
			ids = append(ids, id...)
		} else if id, ok := arg.([]string); ok {
			for _, idStr := range id {
				ids = append(ids, CutIdStr(idStr, sep)...)
			}
		} else if id, ok := arg.(int); ok {
			ids = append(ids, int64(id))
		} else if id, ok := arg.(int32); ok {
			ids = append(ids, int64(id))
		}
	}
	return ids
}

func CutIdStr(s string, sep string) []int64 {
	if sep == "" {
		num, _ := strconv.ParseInt(s, 10, 64)
		return []int64{num}
	}
	parts := strings.Split(s, sep)
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		if part == "" { // 处理空字符串
			continue
		}
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, num)
	}
	return result
}

func JoinIds(sep string, args ...any) string {
	parts := make([]string, 0)
	for _, arg := range args {
		if id, ok := arg.(int64); ok {
			parts = append(parts, strconv.FormatInt(id, 10))
		} else if id, ok := arg.(int); ok {
			parts = append(parts, strconv.Itoa(id))
		} else if id, ok := arg.(int32); ok {
			parts = append(parts, strconv.FormatInt(int64(id), 10))
		} else if idStr, ok := arg.(string); ok {
			parts = append(parts, idStr)
		} else if idSlice, ok := arg.([]int64); ok {
			for _, id := range idSlice {
				parts = append(parts, strconv.FormatInt(id, 10))
			}
		} else if idSlice, ok := arg.([]int); ok {
			for _, id := range idSlice {
				parts = append(parts, strconv.Itoa(id))
			}
		} else if idSlice, ok := arg.([]string); ok {
			parts = append(parts, idSlice...)
		}
	}
	return strings.Join(parts, sep)
}

func GetRandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// Deprecated: This function simply calls [utils.Distinct].
func RemoveDuplicates[T comparable](ids []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0, len(ids))

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func CamelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// UnmarshalList 字符串转换为列表
func UnmarshalList[T any](s string) ([]T, error) {
	var data []T

	if s == "" {
		return data, nil
	}

	err := json.Unmarshal([]byte(s), &data)

	return data, errorx.WithStack(err)
}

// MarshalList 列表转换为字符串
func MarshalList[T any](data []T) (string, error) {
	if len(data) == 0 {
		return "[]", nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "", errorx.WithStack(err)
	}

	return string(b), nil
}

func Unmarshal[T any](s string) (*T, error) {
	var data T
	err := json.Unmarshal([]byte(s), &data)
	return &data, err
}

// 1.2.3 -> 可排序字符串
func VersionToSortableString(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", errorx.New("invalid version")
	}

	buf := make([]string, 0, len(parts))
	for _, p := range parts {
		num, err := strconv.Atoi(p)
		if err != nil {
			return "", errorx.WithStack(err)
		}
		buf = append(buf, fmt.Sprintf("%05d", num)) // 每段固定5位
	}
	return strings.Join(buf, "."), nil
}

// 可排序字符串 -> 1.2.3
func SortableStringToVersion(s string) string {
	parts := strings.Split(s, ".")
	buf := make([]string, 0, len(parts))
	for _, p := range parts {
		num, _ := strconv.Atoi(p)
		buf = append(buf, fmt.Sprintf("%d", num))
	}
	return strings.Join(buf, ".")
}

var (
	ShouldRetryError = errors.New("should retry")
)

func RetryWithTimeout[T any](ctx context.Context, timeout time.Duration, attempts int32, f func(ctx context.Context) (T, error)) (T, error) {
	var (
		err    error
		i      int32
		result T
	)
	for i = 0; i < attempts; i++ {
		fnCtx, cancel := context.WithTimeout(ctx, timeout)
		result, err = f(fnCtx)
		cancel()
		if err == nil {
			return result, nil
		}

		// 超时和 ShouldRetryError 进行重试
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ShouldRetryError) {
			continue
		}
		return result, err
	}

	return result, err
}
