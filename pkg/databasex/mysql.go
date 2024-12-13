package databasex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/jmoiron/sqlx"
)

type MySql struct {
	db   *sqlx.DB
	tx   *sqlx.Tx
	conn *sqlx.Conn // the Conn of the Tx, when tx != nil
}

func (m *MySql) Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.tx != nil {
		return m.tx.SelectContext(ctx, dst, query, args...)
	}
	return m.db.SelectContext(ctx, dst, query, args...)
}

func (m *MySql) Get(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.tx != nil {
		return m.tx.GetContext(ctx, dst, query, args...)
	}
	return m.db.GetContext(ctx, dst, query, args...)
}

func (m *MySql) QueryX(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.tx != nil {
		return m.tx.QueryContext(ctx, query, args...)
	}
	return m.db.QueryContext(
		ctx,
		query,
		args...)
}

func (m *MySql) QueryRowX(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.tx != nil {
		return m.tx.QueryRowContext(ctx, query, args...)
	}
	return m.db.QueryRowContext(
		ctx,
		query, args...)
}

func (m *MySql) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.tx != nil {
		return m.tx.ExecContext(ctx, query, args...)
	}
	return m.db.ExecContext(
		ctx,
		query,
		args...)
}

func (m *MySql) Transact(ctx context.Context, iso sql.IsolationLevel, txFunc func(database Database) error) (err error) {

	// For the levels which require retry, see
	// https://www.postgresql.org/docs/11/transaction-iso.html.
	opts := &sql.TxOptions{Isolation: iso}

	return m.transact(ctx, opts, txFunc)
}

func (m *MySql) transact(ctx context.Context, opts *sql.TxOptions, txFunc func(database Database) error) (err error) {
	if m.InTransaction() {
		return errors.New("db transact function was called on a DB already in a transaction")
	}

	conn, err := m.db.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}

	//defer func() {
	//	if p := recover(); p != nil {
	//		tx.Rollback()
	//	} else if err != nil {
	//		tx.Rollback()
	//	} else {
	//		if txErr := tx.Commit(); txErr != nil {
	//			err = fmt.Errorf("tx commit: %w", txErr)
	//		}
	//	}
	//}()

	mysql := &MySql{
		db:   m.db,
		tx:   tx,
		conn: conn,
	}
	//dbtx.opts = *opts

	if err := txFunc(mysql); err != nil {
		tx.Rollback()
		return fmt.Errorf("fn(tx): %w", err)
	}

	return tx.Commit()
}

func (m *MySql) InTransaction() bool {
	return m.tx != nil
}

func NewMySql(cfg *config.Config) (Database, error) {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10) // Set connection pool limits if needed
	return &MySql{
		db: db,
	}, nil
}
