package datax

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"kratos-admin/pkg/toolbox/utils"
)

var keyReg = regexp.MustCompile(`\$(\w+)`)

type Query struct {
	Query       string   `json:"query,omitempty"`
	Key         string   `json:"key,omitempty"`
	PositionKey string   `json:"positionKey,omitempty"`
	Fields      []string `json:"fields,omitempty"`
	Unify       string   `json:"-"`
	TTL         any      `json:"-"`
}

func (q *Query) KeyByArgs(args ...any) string {
	return fmt.Sprintf(q.PositionKey, args...)
}

func (q *Query) KeyByEntity(entity any) string {
	value := reflect.ValueOf(entity).Elem()
	args := make([]any, len(q.Fields))
	for idx, field := range q.Fields {
		args[idx] = value.FieldByName(field).Interface()
	}
	return fmt.Sprintf(q.PositionKey, args...)
}

func (q *Query) WithTTL(ttl any) *Query {
	q.TTL = ttl
	return q
}

func (q *Query) String() string {
	return q.Unify
}

func NewQuery(query, key string) *Query {
	q := &Query{Query: query, Key: key}
	q.PositionKey = keyReg.ReplaceAllString(key, "%v")
	q.Fields = utils.Map(keyReg.FindAllString(key, -1), func(s string) string {
		return strings.TrimPrefix(s, "$")
	})
	q.Unify = fmt.Sprintf("/*cache_enable*/ %s", query)
	return q
}
