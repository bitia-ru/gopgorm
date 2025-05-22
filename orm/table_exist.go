package pgorm

import (
	"github.com/Masterminds/squirrel"
	"reflect"
)

func (db *Db) TableExists(name string) (bool, error) {
	q := db.
		Select(
			reflect.TypeOf(
				struct {
					TableName string `json:"table_name"`
				}{},
			),
			"table_name",
		).
		From("information_schema.tables").
		Where(squirrel.Eq{"table_name": name})

	values, err := q.Values()

	if err != nil {
		return false, err
	}

	return len(values) > 0, nil
}
