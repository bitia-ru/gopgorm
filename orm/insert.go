package pgorm

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/bitia-ru/gopgorm/db"
	"reflect"
	"strings"
)

type InsertQuery struct {
	err              error
	builderType      sq.StatementBuilderType
	builder          sq.InsertBuilder
	returningColumns []string
	db               *Db
}

func (db *Db) InsertInto(tableName string) db.InsertQuery {
	q := InsertQuery{
		db:          db,
		builderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	q.builder = q.builderType.Insert(tableName)

	return q
}

func (q InsertQuery) Values(v any) db.InsertQuery {
	switch m := v.(type) {
	case map[string]any:
		var columns []string
		var values []any

		for key, value := range m {
			if key == "id" {
				q.err = fmt.Errorf("trying to insert id")
				return q
			}

			columns = append(columns, key)
			values = append(values, value)
		}
		q.builder = q.builder.Columns(columns...).Values(values...)

		return q
	}

	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	t := reflect.TypeOf(v)

	if isPtr {
		t = reflect.TypeOf(v).Elem()
	}

	var columns []string
	var values []any

	num := t.NumField()

	_ = num

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() || field.Tag.Get("serial") == "true" {
			continue
		}

		columnName := field.Tag.Get("json")

		if columnName == "" {
			if field.Tag.Get("base_struct") == "true" {
				fmt.Println("base struct")
			}
			continue
		}

		reflectValue := reflect.ValueOf(v)

		if isPtr {
			reflectValue = reflectValue.Elem()
		}

		columns = append(columns, columnName)
		values = append(values, reflectValue.Field(i).Interface())
	}

	q.builder = q.builder.Columns(columns...).Values(values...)

	return q
}

func (q InsertQuery) Returning(columns ...string) db.InsertQuery {
	if q.err != nil {
		return q
	}

	q.returningColumns = columns

	q.builder = q.builder.Suffix("RETURNING " + strings.Join(columns, ", "))

	return q
}

type ReturnedValues struct {
	Values map[string]any
}

func (rv ReturnedValues) Value(column string) (any, error) {
	r, ok := rv.Values[column]

	if !ok {
		return nil, fmt.Errorf("column %s not found in returned values", column)
	}

	return r, nil
}

func (q InsertQuery) Exec() (db.InsertResults, error) {
	if q.err != nil {
		return nil, q.err
	}

	sql, args, err := q.builder.ToSql()

	if len(q.returningColumns) > 0 {
		resRows, err := q.db.connection.Query(q.db.ctx, sql, args...)

		if err != nil {
			return nil, err
		}

		returnedValuesMap := make(map[string]any)

		if !resRows.Next() {
			return nil, fmt.Errorf("no result returened")
		}

		values, err := resRows.Values()

		if err != nil {
			return nil, err
		}

		for i, column := range q.returningColumns {
			returnedValuesMap[column] = values[i]
		}

		return ReturnedValues{
			Values: returnedValuesMap,
		}, nil
	}

	_, err = q.db.connection.Exec(q.db.ctx, sql, args...)

	return nil, err
}
