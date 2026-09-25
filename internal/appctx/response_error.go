package appctx

import "github.com/hanifkf12/hanif_skeleton/pkg/apperror"

// statusFromKind memetakan klasifikasi error ke kode status. Kind yang tidak
// dikenal jatuh ke 500 — fail-closed, bukan fail-open.
func statusFromKind(kind apperror.Kind) int {
	switch kind {
	case apperror.KindInvalid:
		return StatusBadRequest
	case apperror.KindUnauthorized:
		return StatusUnauthorized
	case apperror.KindForbidden:
		return StatusForbidden
	case apperror.KindNotFound:
		return StatusNotFound
	case apperror.KindConflict:
		return StatusConflict
	case apperror.KindPaymentRequired:
		return StatusPaymentRequired
	case apperror.KindUnavailable:
		return StatusServiceUnavailable
	default:
		return StatusInternalServerError
	}
}

// ResponseFromError mengubah error apa pun menjadi envelope respons. Ini
// satu-satunya tempat yang memetakan error domain ke status HTTP.
//
// Hanya Message dari *apperror.Error yang sampai ke klien; error lain menjadi
// 500 generik, sehingga detail driver tidak pernah bocor.
//
// Alasan ditulis ke Errors dan bukan Message supaya bentuk wire-nya tetap sama
// dengan respons error yang sudah ada — semuanya menaruh alasan di "errors"
// dan membiarkan "message" kosong. Status juga sengaja tidak di-set (false +
// omitempty) agar "status" tetap absen dari body error.
func ResponseFromError(err error) *Response {
	if err == nil {
		return NewResponse().WithCode(StatusOK)
	}

	return NewResponse().
		WithCode(statusFromKind(apperror.KindOf(err))).
		WithErrors(apperror.MessageOf(err))
}
