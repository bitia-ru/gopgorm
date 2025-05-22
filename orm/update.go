package pgorm

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/bitia-ru/gopgorm/db"
	"reflect"
)

type updateQuery struct {
	err         error
	builderType sq.StatementBuilderType
	builder     sq.UpdateBuilder
	db          *Db
}

func (db *Db) Update(tableName string) db.UpdateQuery {
	q := updateQuery{
		db:          db,
		builderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	q.builder = q.builderType.Update(tableName)

	return q
}

func (q updateQuery) Where(filter map[string]any) db.UpdateQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.Where(filter)

	return q
}

func (q updateQuery) Limit(limit uint64) db.UpdateQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.Limit(limit)

	return q
}

func (q updateQuery) Values(v any) db.UpdateQuery {
	switch m := v.(type) {
	case map[string]any:
		q.builder = q.builder.SetMap(m)

		return q
	}

	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	t := reflect.TypeOf(v)

	if isPtr {
		t = reflect.TypeOf(v).Elem()
	}

	num := t.NumField()

	_ = num

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		reflectValue := reflect.ValueOf(v)

		if field.Tag.Get("serial") == "true" {
			q.builder = q.builder.Where(map[string]any{"id": reflectValue.Field(i).Interface()})
			continue
		}

		if !field.IsExported() {
			continue
		}

		columnName := field.Tag.Get("json")

		if columnName == "" {
			if field.Tag.Get("base_struct") == "true" {
				fmt.Println("base struct")
			}
			continue
		}

		if isPtr {
			reflectValue = reflectValue.Elem()
		}

		q.builder = q.builder.Set(columnName, reflectValue.Field(i).Interface())
	}

	return q
}

func (q updateQuery) Exec() error {
	if q.err != nil {
		return q.err
	}

	sql, args, err := q.builder.ToSql()

	if err != nil {
		q.err = err
		return q.err
	}

	if _, err := q.db.connection.Exec(q.db.ctx, sql, args...); err != nil {
		q.err = err
		return q.err
	}

	return nil
}
