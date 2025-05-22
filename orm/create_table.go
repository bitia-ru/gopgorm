package pgorm

import (
	"fmt"
	"reflect"
	"strings"
)

type columnDescription struct {
	name       string
	dataType   string
	primaryKey bool
}

func (db *Db) CreateTable(rowType reflect.Value, tableName string) error {
	query := fmt.Sprintf("CREATE TABLE %s", tableName)

	_ = query

	v := rowType

	var columns []columnDescription

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := v.Type().Field(i)

		if !fieldType.IsExported() {
			continue
		}

		var column columnDescription

		column.name = fieldType.Tag.Get("json")

		if column.name == "" {
			return fmt.Errorf("field %s has no json tag", fieldType.Name)
		}

		columnType := fieldValue.Kind().String()

		if fieldType.Tag.Get("serial") == "true" {
			if columnType != "int" && columnType != "int64" {
				return fmt.Errorf("serial tag is only supported for int and int64 types")
			}

			column.dataType = "SERIAL"
		} else {
			switch columnType {
			case "string":
				column.dataType = "VARCHAR(255)"
			case "text":
				column.dataType = "TEXT"
			case "int64":
				column.dataType = "BIGINT"
			case "int":
				column.dataType = "INTEGER"
			case "bool":
				column.dataType = "BOOLEAN"
			default:
				return fmt.Errorf("unsupported type %s for field %s", columnType, fieldType.Name)
			}
		}

		if fieldType.Tag.Get("primary_key") == "true" {
			column.primaryKey = true
		}

		columns = append(columns, column)
	}

	var columnsStrParts []string

	for _, column := range columns {
		columnsStrParts = append(columnsStrParts, fmt.Sprintf("%s %s", column.name, column.dataType))
	}

	query += fmt.Sprintf(" (%s)", strings.Join(columnsStrParts, ", "))

	_, err := db.connection.Exec(db.ctx, query)

	return err
}
