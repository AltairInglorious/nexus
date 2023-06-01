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

// SelectQuery represents a structure for generating SQL SELECT queries.
// TableName is the name of the table in the database.
// Fields is an optional slice of fields (columns) to be selected. If empty, all fields (*) are selected.
// Filter is an optional filter to be applied to the query as a WHERE clause.
type SelectQuery struct {
	TableName string
	Fields    []string
	Filter    any
}

// String method generates a SQL SELECT query string based on the SelectQuery values.
// It uses the UseFilter function to add any filter conditions to the query.
func (s SelectQuery) String() string {
	var q string
	if len(s.Fields) == 0 {
		q = fmt.Sprintf("SELECT * FROM %s", s.TableName)
	} else {
		q = fmt.Sprintf("SELECT %s FROM %s", strings.Join(s.Fields, ", "), s.TableName)
	}
	return UseFilter(s.Filter, q)
}

// NewSelectAll is a function that generates a new SelectQuery for selecting all fields from a specific table.
// It accepts the table name as an argument.
func NewSelectAll(t string) SelectQuery {
	return SelectQuery{
		TableName: t,
	}
}

// NewSelect is a function that generates a new SelectQuery for selecting specific fields from a specific table.
// It accepts the table name as the first argument and a variable number of string arguments for the fields.
func NewSelect(t string, f ...string) SelectQuery {
	return SelectQuery{
		TableName: t,
		Fields:    f,
	}
}

// WithFilter is a method that adds a filter to the SelectQuery and returns the updated SelectQuery.
// It accepts an interface{} as a filter which can be any type that is acceptable by the UseFilter function.
func (s SelectQuery) WithFilter(f any) SelectQuery {
	if f == nil {
		return s
	}
	s.Filter = f
	return s
}

// DB represents a wrapper over surrealdb.DB that includes a concurrent map for caching purposes.
type DB struct {
	s *surrealdb.DB
	c sync.Map
}

// New is a function that creates a new instance of DB.
// It establishes a connection to the SurrealDB with the provided URL and credentials,
// then switches to the specified namespace and database.
// If successful, it returns a pointer to the DB instance; otherwise, it returns an error.
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

// GeneralCreate is a generic function that handles the creation of a new record in the database.
// It takes a thing string which represents the table name, and a map of data for the record.
// After successfully creating the record, it clears the relevant cache.
// d: Pointer to DB instance
// thing: table name in the database
// data: map containing field-value pairs for the new record
// Returns a pointer to the created record of type T or an error.
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

// GeneralSelect is a generic function that handles querying of records from the database.
// It first checks if the query results are present in the cache. If not, it executes the query and
// stores the result in cache.
// d: Pointer to DB instance
// s: SelectQuery structure which encapsulates the SELECT query details
// Returns a slice of records of type T or an error.
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

// UseFilter takes an interface and a query string as input and adds WHERE and LIMIT clauses to the query
// based on the non-nil fields of the interface. It ignores the "limit" field while constructing WHERE clauses.
// f: Filter interface with optional fields
// q: Query string to which filters will be appended
// Returns the modified query string.
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
