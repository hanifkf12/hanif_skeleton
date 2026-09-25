package appctx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseFromErrorMapsKindToStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid", apperror.Invalid("bad input"), appctx.StatusBadRequest},
		{"unauthorized", apperror.Unauthorized("nope"), appctx.StatusUnauthorized},
		{"forbidden", apperror.Forbidden("nope"), appctx.StatusForbidden},
		{"not found", apperror.NotFound("missing"), appctx.StatusNotFound},
		{"conflict", apperror.Conflict("dup"), appctx.StatusConflict},
		{"payment required", apperror.PaymentRequired("pay"), appctx.StatusPaymentRequired},
		{"unavailable", apperror.Unavailable("down"), appctx.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := appctx.ResponseFromError(tt.err)
			assert.Equal(t, tt.want, resp.Code)
		})
	}
}

// Error yang belum diklasifikasi harus jadi 500, bukan 400 — fail-closed.
func TestResponseFromErrorDefaultsToInternal(t *testing.T) {
	for _, err := range []error{
		errors.New("some driver error"),
		fmt.Errorf("wrapped: %w", errors.New("boom")),
	} {
		resp := appctx.ResponseFromError(err)
		assert.Equal(t, appctx.StatusInternalServerError, resp.Code)
		assert.Equal(t, "Internal server error", resp.Errors)
	}
}

// Ini alasan utama error infra ada: detail driver tidak boleh sampai ke klien.
func TestResponseFromInternalDoesNotLeakCause(t *testing.T) {
	driverErr := errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`)

	resp := appctx.ResponseFromError(apperror.Internal(driverErr))

	require.Equal(t, appctx.StatusInternalServerError, resp.Code)
	require.Equal(t, "Internal server error", resp.Errors)

	body := string(resp.Byte())
	assert.NotContains(t, body, "users_email_key", "nama constraint tidak boleh bocor")
	assert.NotContains(t, body, "pq:", "nama driver tidak boleh bocor")
}

// Bentuk wire harus sama dengan respons error yang sudah ada: tidak ada
// "message", tidak ada "status", dan "data" absen (bukan null).
func TestResponseFromErrorKeepsWireShape(t *testing.T) {
	resp := appctx.ResponseFromError(apperror.Forbidden("Cannot modify another user"))

	body := string(resp.Byte())
	assert.JSONEq(t, `{
		"code": 403,
		"timestamp": "`+resp.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00")+`",
		"errors": "Cannot modify another user"
	}`, body)

	assert.NotContains(t, body, `"status"`)
	assert.NotContains(t, body, `"message"`)
	assert.NotContains(t, body, `"data"`)
}

func TestResponseFromNilErrorIsOK(t *testing.T) {
	resp := appctx.ResponseFromError(nil)
	assert.Equal(t, appctx.StatusOK, resp.Code)
}
