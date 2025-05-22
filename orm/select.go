package pgorm

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/bitia-ru/gopgorm/db"
	"reflect"
)

type SelectQuery struct {
	err              error
	rowType          reflect.Type
	structFieldNames []string
	builderType      sq.StatementBuilderType
	builder          sq.SelectBuilder
	db               *Db
}

func (db *Db) Select(rowType reflect.Type, columnNames ...string) db.SelectQuery {
	q := SelectQuery{
		db:          db,
		rowType:     rowType,
		builderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	var columns []string

	if reflect.DeepEqual(columnNames, []string{"*"}) {
		for i := 0; i < q.rowType.NumField(); i++ {
			field := q.rowType.Field(i)
			column := field.Tag.Get("json")

			if column == "" {
				q.err = fmt.Errorf("field %s has no json tag", field.Name)
				break
			}

			columns = append(columns, column)
			q.structFieldNames = append(q.structFieldNames, field.Name)
		}
	} else {
		q.structFieldNames = make([]string, 0)

		for _, columnName := range columnNames {
			var field *reflect.StructField = nil

			for i := 0; i < q.rowType.NumField(); i++ {
				f := q.rowType.Field(i)

				if !f.IsExported() {
					continue
				}

				if f.Tag.Get("json") == columnName {
					field = &f
					break
				}
			}

			if field == nil {
				q.err = fmt.Errorf("column %s not found in row type", columnName)
				return q
			}

			columns = append(columns, columnName)

			q.structFieldNames = append(q.structFieldNames, field.Name)
		}
	}

	q.builder = q.builderType.Select(columns...)

	return q
}

func (q SelectQuery) From(table string) db.SelectQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.From(table)

	return q
}

func (q SelectQuery) Where(filter map[string]any) db.SelectQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.Where(filter)

	return q
}

func (q SelectQuery) OrderBy(order string) db.SelectQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.OrderBy(order)

	return q
}

func (q SelectQuery) Limit(limit uint64) db.SelectQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.Limit(limit)

	return q
}

func (q SelectQuery) Values() ([]any, error) {
	if q.err != nil {
		return nil, q.err
	}

	sql, args, err := q.builder.ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := q.db.connection.Query(q.db.ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []any

	for rows.Next() {
		values, err := rows.Values()

		if err != nil {
			return nil, err
		}

		if len(values) != len(q.structFieldNames) {
			return nil, fmt.Errorf("invalid number of values in a row")
		}

		row := reflect.New(q.rowType).Elem()

		for i, fieldName := range q.structFieldNames {
			field, ok := q.rowType.FieldByName(fieldName)

			if !ok {
				return nil, fmt.Errorf("field %s not found in row type", fieldName)
			}

			value := values[i]

			if value != nil {
				structField := row.FieldByName(field.Name)

				if structField.CanSet() {
					structField.Set(reflect.ValueOf(value).Convert(field.Type))
				}
			}
		}

		result = append(result, row.Interface())
	}

	return result, nil
}
