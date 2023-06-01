package db

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/surrealdb/surrealdb.go"
)

type CacheKey struct {
	TableName string
	Query     string
}

type SelectQuery struct {
	TableName string
	Fields    []string
	Filter    any
}

func (s SelectQuery) String() string {
	var q string
	if len(s.Fields) == 0 {
		q = fmt.Sprintf("SELECT * FROM %s", s.TableName)
	} else {
		q = fmt.Sprintf("SELECT %s FROM %s", strings.Join(s.Fields, ", "), s.TableName)
	}
	return UseFilter(s.Filter, q)
}

func NewSelectAll(t string) SelectQuery {
	return SelectQuery{
		TableName: t,
	}
}
func NewSelect(t string, f ...string) SelectQuery {
	return SelectQuery{
		TableName: t,
		Fields:    f,
	}
}
func (s SelectQuery) WithFilter(f any) SelectQuery {
	if f == nil {
		return s
	}
	s.Filter = f
	return s
}

type DB struct {
	s *surrealdb.DB
	c sync.Map
}

func New(url, user, pass, ns, db string) (*DB, error) {
	s, err := surrealdb.New(url)
	if err != nil {
		return nil, err
	}

	if _, err = s.Signin(map[string]interface{}{
		"user": user,
		"pass": pass,
	}); err != nil {
		return nil, err
	}

	if _, err = s.Use(ns, db); err != nil {
		return nil, err
	}

	return &DB{
		s: s,
		c: sync.Map{},
	}, nil
}

func (d *DB) putQueryToCache(s SelectQuery, value any) {
	d.c.Store(CacheKey{
		TableName: s.TableName,
		Query:     s.String(),
	}, value)
}

func (d *DB) getQueryFromCache(s SelectQuery) (any, error) {
	if v, ok := d.c.Load(CacheKey{
		TableName: s.TableName,
		Query:     s.String(),
	}); ok {
		return v, nil
	}

	return nil, fmt.Errorf("not found in cache")
}

func (d *DB) clearCache(t string) {
	d.c.Range(func(k, v interface{}) bool {
		if k.(CacheKey).TableName == t {
			d.c.Delete(k)
		}
		return true
	})
}

func GeneralCreate[T any](d *DB, thing string, data map[string]interface{}) (*T, error) {
	pr, err := d.s.Create(thing, data)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	p := make([]T, 1)
	if err = surrealdb.Unmarshal(pr, &p); err != nil {
		return nil, err
	}
	d.clearCache(thing)
	return &p[0], nil
}

func GeneralSelect[T any](d *DB, s SelectQuery) ([]T, error) {
	cv, err := d.getQueryFromCache(s)
	if err == nil {
		p, ok := cv.([]T)
		if ok {
			return p, nil
		}
	}

	r, err := d.s.Query(s.String(), nil)
	if err != nil {
		return nil, err
	}

	var p []T
	ok, err := surrealdb.UnmarshalRaw(r, &p)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []T{}, nil
	}
	d.putQueryToCache(s, p)
	return p, nil
}

func UseFilter(f interface{}, q string) string {
	if reflect.ValueOf(f).IsNil() {
		return q
	}

	v := reflect.ValueOf(f).Elem()
	typeOfT := v.Type()

	var w []string

	for i := 0; i < v.NumField(); i++ {
		fl := v.Field(i)
		if fl.Kind() == reflect.Ptr && !fl.IsNil() {
			flv := reflect.Indirect(fl).Interface()
			tag := typeOfT.Field(i).Tag.Get("json")
			tagParts := strings.Split(tag, ",")
			fln := tagParts[0]

			if fln == "limit" {
				continue
			}

			switch v := flv.(type) {
			case string:
				w = append(w, fmt.Sprintf("%s = '%s'", fln, v))
			case bool:
				w = append(w, fmt.Sprintf("%s = %t", fln, v))
			case int:
				w = append(w, fmt.Sprintf("%s = %d", fln, v))
			}
		}
	}

	if len(w) > 0 {
		q += " WHERE " + strings.Join(w, " AND ")
	}

	if fll := v.FieldByName("Limit"); fll.IsValid() && !fll.IsNil() {
		q += fmt.Sprintf(" LIMIT %d", fll.Elem().Interface())
	}

	return q
}
