package campaign

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
	"github.com/stretchr/testify/require"
)

// recordingDB captures the SQL a repository sends, so tests check what Postgres would receive.
type recordingDB struct {
	query    string
	args     []interface{}
	affected int64
}

type result int64

func (r result) LastInsertId() (int64, error) { return 0, nil }
func (r result) RowsAffected() (int64, error) { return int64(r), nil }

func (d *recordingDB) record(query string, args []interface{}) { d.query, d.args = query, args }

func (d *recordingDB) QueryX(_ context.Context, q string, a ...interface{}) (*sql.Rows, error) {
	d.record(q, a)
	return nil, nil
}
func (d *recordingDB) QueryRowX(_ context.Context, q string, a ...interface{}) *sql.Row {
	d.record(q, a)
	return nil
}
func (d *recordingDB) Get(_ context.Context, _ interface{}, q string, a ...interface{}) error {
	d.record(q, a)
	return sql.ErrNoRows
}
func (d *recordingDB) Select(_ context.Context, _ interface{}, q string, a ...interface{}) error {
	d.record(q, a)
	return nil
}
func (d *recordingDB) Exec(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
	d.record(q, a)
	return result(d.affected), nil
}
func (d *recordingDB) Transact(context.Context, sql.IsolationLevel, func(databasex.Database) error) error {
	return nil
}
func (d *recordingDB) InTransaction() bool { return false }

func newRepo(t *testing.T, affected int64) (*recordingDB, *campaignRepository) {
	t.Helper()
	// Same setting bootstrap.RegistryDatabase applies in production.
	sqlbuilder.SetDefaultDriver(sqlbuilder.DriverPostgreSQL)
	db := &recordingDB{affected: affected}
	return db, &campaignRepository{db: db}
}

func TestCreateSendsGeneratedIDWithPostgresPlaceholders(t *testing.T) {
	db, repo := newRepo(t, 1)
	c := &entity.Campaign{Name: "x", TargetDonation: 10}

	require.NoError(t, repo.Create(context.Background(), c))

	require.NotEmpty(t, c.ID)
	require.Contains(t, db.query, "id")
	require.Contains(t, db.query, "$1")
	require.NotContains(t, db.query, "?")
	require.Contains(t, db.args, c.ID)
}

func TestGetByIDUsesPostgresPlaceholderAndMapsNoRows(t *testing.T) {
	db, repo := newRepo(t, 0)

	_, err := repo.GetByID(context.Background(), "missing")

	require.Contains(t, db.query, "id = $1")
	require.Equal(t, apperror.KindNotFound, apperror.KindOf(err))
}

func TestUpdateAndDeleteReportMissingRowAsNotFound(t *testing.T) {
	_, repo := newRepo(t, 0)
	ctx := context.Background()

	err := repo.Update(ctx, &entity.Campaign{ID: "missing", Name: "x"})
	require.Equal(t, apperror.KindNotFound, apperror.KindOf(err))

	err = repo.Delete(ctx, "missing")
	require.Equal(t, apperror.KindNotFound, apperror.KindOf(err))
}
