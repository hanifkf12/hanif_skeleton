package databasex

import (
	"context"
	"database/sql"
)

const (
	constStringMysqlFormat = ""
)

type Database interface {
	QueryX(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowX(ctx context.Context, query string, args ...interface{}) *sql.Row
	Get(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Transact(ctx context.Context, iso sql.IsolationLevel, txFunc func(database Database) error) (err error)
	InTransaction() bool
}
