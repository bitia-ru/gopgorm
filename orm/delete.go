package pgorm

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/bitia-ru/gopgorm/db"
)

type deleteQuery struct {
	err         error
	tableName   string
	builderType sq.StatementBuilderType
	builder     sq.DeleteBuilder
	db          *Db
}

func (db *Db) DeleteFrom(tableName string) db.DeleteQuery {
	q := deleteQuery{
		db:          db,
		tableName:   tableName,
		builderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	q.builder = q.builderType.Delete(tableName)

	return q
}

func (q deleteQuery) Where(filter map[string]any) db.DeleteQuery {
	if q.err != nil {
		return q
	}

	q.builder = q.builder.Where(filter)

	return q
}

func (q deleteQuery) Exec() error {
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
