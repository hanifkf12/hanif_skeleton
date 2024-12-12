package databasex

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/jmoiron/sqlx"
)

type mySql struct {
	db *sqlx.DB
}

func (m *mySql) Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return m.db.SelectContext(ctx, dst, query, args...)
}

func (m *mySql) Get(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return m.db.GetContext(ctx, dst, query, args...)
}

func (m *mySql) QueryX(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.QueryContext(
		ctx,
		query,
		args...)
}

func (m *mySql) QueryRowX(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.db.QueryRowContext(
		ctx,
		query, args...)
}

func (m *mySql) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.db.ExecContext(
		ctx,
		query,
		args...)
}

func NewMySql(cfg *config.Config) (Database, error) {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10) // Set connection pool limits if needed
	return &mySql{
		db: db,
	}, nil
}
