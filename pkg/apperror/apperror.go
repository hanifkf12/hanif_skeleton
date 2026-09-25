// Package apperror menyediakan error terklasifikasi yang memisahkan pesan aman
// untuk klien dari penyebab asli yang hanya boleh masuk log dan trace.
//
// Paket ini sengaja hanya bergantung pada stdlib supaya bisa dipakai di layer
// mana pun — termasuk entity, yang tidak boleh meng-import internal/.
package apperror

import (
	"errors"
	"fmt"
)

// Kind mengklasifikasikan error untuk dipetakan ke status HTTP. Nilai Kind
// adalah kosakata aplikasi, bukan detail transport.
type Kind string

const (
	KindInvalid         Kind = "invalid"
	KindUnauthorized    Kind = "unauthorized"
	KindForbidden       Kind = "forbidden"
	KindNotFound        Kind = "not_found"
	KindConflict        Kind = "conflict"
	KindPaymentRequired Kind = "payment_required"
	KindUnavailable     Kind = "unavailable"
	KindInternal        Kind = "internal"
)

// internalMessage adalah satu-satunya pesan yang pernah dilihat klien untuk
// error tak terklasifikasi. Menjaga error driver (mis. *pq.Error yang memuat
// nama tabel dan constraint) agar tidak pernah bocor ke respons HTTP.
const internalMessage = "Internal server error"

// Error selalu dipakai sebagai *Error.
//
//	Message -> aman dilihat klien.
//	cause   -> TIDAK PERNAH dilihat klien; hanya untuk log dan trace.
type Error struct {
	Kind    Kind
	Code    string // kode bisnis yang stabil, mis. "user_not_found"
	Message string // aman untuk klien
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap membuat errors.Is/errors.As menembus ke cause asli, sehingga
// pemanggil masih bisa memeriksa sentinel error dari layer bawah.
func (e *Error) Unwrap() error { return e.cause }

// WithCode menempelkan kode bisnis yang stabil dan mengembalikan error yang sama.
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// New membuat error tanpa cause. Pesan dianggap aman untuk klien.
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

// Newf seperti New dengan format string.
func Newf(kind Kind, format string, a ...any) *Error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, a...)}
}

// Wrap melampirkan cause dan mengklasifikasikan error.
//
// Idempoten: kalau err sudah *Error, Kind/Code/Message-nya dipertahankan apa
// adanya dan hanya cause-nya yang diperbarui. Ini membuat Wrap aman dipanggil
// di setiap layer tanpa menghapus klasifikasi dari layer bawah.
func Wrap(err error, kind Kind, message string) *Error {
	if err == nil {
		return nil
	}

	var existing *Error
	if errors.As(err, &existing) {
		return &Error{
			Kind:    existing.Kind,
			Code:    existing.Code,
			Message: existing.Message,
			cause:   err,
		}
	}

	return &Error{Kind: kind, Message: message, cause: err}
}

// Wrapf seperti Wrap dengan format string.
func Wrapf(err error, kind Kind, format string, a ...any) *Error {
	return Wrap(err, kind, fmt.Sprintf(format, a...))
}

func Invalid(message string) *Error         { return New(KindInvalid, message) }
func Unauthorized(message string) *Error    { return New(KindUnauthorized, message) }
func Forbidden(message string) *Error       { return New(KindForbidden, message) }
func NotFound(message string) *Error        { return New(KindNotFound, message) }
func Conflict(message string) *Error        { return New(KindConflict, message) }
func Unavailable(message string) *Error     { return New(KindUnavailable, message) }
func PaymentRequired(message string) *Error { return New(KindPaymentRequired, message) }

// Internal membungkus error tak terduga. Message-nya SELALU pesan generik,
// apa pun isi cause — supaya detail internal tidak pernah sampai ke klien.
func Internal(cause error) *Error {
	return &Error{Kind: KindInternal, Message: internalMessage, cause: cause}
}

// KindOf mengembalikan Kind dari err. Apa pun yang bukan *Error dianggap
// KindInternal, sehingga error driver yang belum di-wrap tidak akan pernah
// disalahartikan sebagai error klien (fail-closed).
func KindOf(err error) Kind {
	if err == nil {
		return KindInternal
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}

	return KindInternal
}

// CodeOf mengembalikan kode bisnis, atau string kosong kalau tidak ada.
func CodeOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// MessageOf mengembalikan pesan yang aman dilihat klien. Error yang bukan
// *Error selalu menjadi pesan generik.
func MessageOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return internalMessage
}
