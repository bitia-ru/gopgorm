package db

import "reflect"

type SelectQuery interface {
	From(table string) SelectQuery
	Where(filter map[string]any) SelectQuery
	OrderBy(order string) SelectQuery
	Limit(limit uint64) SelectQuery
	Values() ([]any, error)
}

type InsertQuery interface {
	Values(values any) InsertQuery
	Returning(columns ...string) InsertQuery
	Exec() (InsertResults, error)
}

type InsertResults interface {
	Value(column string) (any, error)
}

type DeleteQuery interface {
	Where(filter map[string]any) DeleteQuery
	Exec() error
}

type UpdateQuery interface {
	Where(filter map[string]any) UpdateQuery
	Values(values any) UpdateQuery
	Limit(limit uint64) UpdateQuery
	Exec() error
}

type Db interface {
	Select(rowType reflect.Type, fieldNames ...string) SelectQuery
	InsertInto(tableName string) InsertQuery
	DeleteFrom(tableName string) DeleteQuery
	Update(tableName string) UpdateQuery

	CreateTable(rowType reflect.Value, tableName string) error
	TableExists(name string) (bool, error)
}
