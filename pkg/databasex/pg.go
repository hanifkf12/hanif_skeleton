package databasex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db   *sqlx.DB
	tx   *sqlx.Tx
	conn *sqlx.Conn // the Conn of the Tx, when tx != nil
}

func (p *Postgres) Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if p.tx != nil {
		return p.tx.SelectContext(ctx, dst, query, args...)
	}
	return p.db.SelectContext(ctx, dst, query, args...)
}

func (p *Postgres) Get(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if p.tx != nil {
		return p.tx.GetContext(ctx, dst, query, args...)
	}
	return p.db.GetContext(ctx, dst, query, args...)
}

func (p *Postgres) QueryX(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.tx != nil {
		return p.tx.QueryContext(ctx, query, args...)
	}
	return p.db.QueryContext(ctx, query, args...)
}

func (p *Postgres) QueryRowX(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if p.tx != nil {
		return p.tx.QueryRowContext(ctx, query, args...)
	}
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *Postgres) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if p.tx != nil {
		return p.tx.ExecContext(ctx, query, args...)
	}
	return p.db.ExecContext(ctx, query, args...)
}

func (p *Postgres) Transact(ctx context.Context, iso sql.IsolationLevel, txFunc func(database Database) error) (err error) {
	opts := &sql.TxOptions{Isolation: iso}
	return p.transact(ctx, opts, txFunc)
}

func (p *Postgres) transact(ctx context.Context, opts *sql.TxOptions, txFunc func(database Database) error) (err error) {
	if p.InTransaction() {
		return errors.New("db transact function was called on a DB already in a transaction")
	}

	conn, err := p.db.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}

	pg := &Postgres{
		db:   p.db,
		tx:   tx,
		conn: conn,
	}

	if err := txFunc(pg); err != nil {
		tx.Rollback()
		return fmt.Errorf("fn(tx): %w", err)
	}

	return tx.Commit()
}

func (p *Postgres) InTransaction() bool {
	return p.tx != nil
}

func NewPostgres(cfg *config.Config) (Database, error) {
	// PostgreSQL connection string format
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(300) // 5 minutes

	return &Postgres{
		db: db,
	}, nil
}
