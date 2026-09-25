// Package appctxtest menyediakan implementasi appctx.Request in-memory untuk
// unit test, sehingga usecase bisa diuji tanpa Fiber dan tanpa HTTP.
//
// Ini paket biasa, bukan file _test.go, supaya bisa dipakai bersama oleh test
// di paket lain — pola yang sama dengan net/http/httptest. Paket ini tidak ikut
// ke binary produksi karena linker Go hanya menyertakan paket yang reachable
// dari main.
package appctxtest

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

// ErrEmptyBody dikembalikan Body kalau tidak ada body yang di-set.
var ErrEmptyBody = errors.New("appctxtest: request has no body")

// Request adalah appctx.Request in-memory. Semua field opsional; nilai nol
// menghasilkan request kosong yang tidak terautentikasi.
type Request struct {
	Ctx         context.Context
	Params      map[string]string
	QueryParams map[string]string
	Headers     map[string]string
	BodyBytes []byte
	File      *multipart.FileHeader
	FileErr   error
	Actor     entity.Actor
	Auth      bool
}

var _ appctx.Request = (*Request)(nil)

// New membuat Request kosong. Map-nya sudah terinisialisasi supaya builder
// bisa dirantai tanpa nil-check.
func New() *Request {
	return &Request{
		Params:      map[string]string{},
		QueryParams: map[string]string{},
		Headers:     map[string]string{},
	}
}

// WithJSONBody men-serialize v menjadi body JSON. Karena signature-nya
// mengembalikan *Request, error marshal dilaporkan saat Body() dipanggil.
func (r *Request) WithJSONBody(v any) *Request {
	bytes, err := json.Marshal(v)
	if err != nil {
		r.BodyBytes = nil
		r.FileErr = err
		return r
	}

	r.BodyBytes = bytes
	return r
}

// WithRawBody mengeset body mentah, berguna untuk menguji body yang tidak valid.
func (r *Request) WithRawBody(b []byte) *Request {
	r.BodyBytes = b
	return r
}

// WithPrincipal menandai request sebagai terautentikasi sebagai actor.
func (r *Request) WithPrincipal(actor entity.Actor) *Request {
	r.Actor = actor
	r.Auth = true
	return r
}

func (r *Request) WithParam(key, value string) *Request {
	r.Params[key] = value
	return r
}

func (r *Request) WithQuery(key, value string) *Request {
	r.QueryParams[key] = value
	return r
}

func (r *Request) WithHeader(key, value string) *Request {
	r.Headers[key] = value
	return r
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Ctx = ctx
	return r
}

// WithFormFile mengeset file yang dikembalikan FormFile.
func (r *Request) WithFormFile(file *multipart.FileHeader) *Request {
	r.File = file
	return r
}

// WithFormFileError membuat FormFile mengembalikan err.
func (r *Request) WithFormFileError(err error) *Request {
	r.FileErr = err
	return r
}

// Context tidak pernah nil — kontrak appctx.Request mensyaratkan itu.
func (r *Request) Context() context.Context {
	if r.Ctx == nil {
		return context.Background()
	}
	return r.Ctx
}

func (r *Request) Param(key string) string { return r.Params[key] }
func (r *Request) Query(key string) string { return r.QueryParams[key] }

func (r *Request) Header(key string) string { return r.Headers[key] }

func (r *Request) Body(dst any) error {
	if r.FileErr != nil && len(r.BodyBytes) == 0 {
		return r.FileErr
	}
	if len(r.BodyBytes) == 0 {
		return ErrEmptyBody
	}
	return json.Unmarshal(r.BodyBytes, dst)
}

func (r *Request) FormFile(key string) (*multipart.FileHeader, error) {
	if r.FileErr != nil {
		return nil, r.FileErr
	}
	if r.File == nil {
		return nil, errors.New("appctxtest: no form file set")
	}
	return r.File, nil
}

func (r *Request) Principal() (entity.Actor, bool) {
	if !r.Auth {
		return entity.Actor{}, false
	}
	return r.Actor, true
}
