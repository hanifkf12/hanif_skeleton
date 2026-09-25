package repository

import (
	"database/sql"
	"errors"

	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/lib/pq"
)

// Kode error PostgreSQL yang punya arti di level aplikasi.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// MapSQLError mengubah error driver menjadi error aplikasi yang terklasifikasi.
//
// Ini satu-satunya tempat di codebase yang tahu soal lib/pq, sehingga
// *pq.Error — yang memuat nama tabel dan constraint — tidak pernah bisa
// mencapai layer atas maupun klien.
//
// notFoundMessage dipakai untuk sql.ErrNoRows, supaya pesan yang dilihat klien
// bisa spesifik per pemanggil (mis. "user not found").
func MapSQLError(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return apperror.Wrap(err, apperror.KindNotFound, notFoundMessage)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch string(pqErr.Code) {
		case pgUniqueViolation:
			return apperror.Wrap(err, apperror.KindConflict, "Resource already exists")
		case pgForeignKeyViolation:
			return apperror.Wrap(err, apperror.KindConflict, "Resource is still referenced by other data")
		}
	}

	// Apa pun yang tidak dikenali menjadi error internal dengan pesan generik.
	return apperror.Internal(err)
}
