package appctx

import (
	"context"
	"mime/multipart"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

// Request adalah pandangan transport-agnostic atas HTTP request yang masuk.
// Ini satu-satunya cara usecase layer boleh menyentuh request, sehingga
// mengganti framework HTTP tidak menyentuh business layer.
//
// Implementasi: handler.fiberRequest (produksi), appctxtest.Request (test).
//
// Interface ini tinggal di appctx, bukan di usecase/contract, karena
// usecase/contract meng-import appctx untuk appctx.Data — menaruh Request di
// sana akan membentuk siklus appctx -> contract -> appctx.
type Request interface {
	// Context mengembalikan context request, dan tidak pernah nil.
	//
	// Sengaja hanya ada satu accessor: Fiber punya Context() (yang mengembalikan
	// *fasthttp.RequestCtx) dan UserContext() (yang mengembalikan context.Context
	// berisi span OpenTelemetry). Menyatukannya di sini menghilangkan seluruh
	// kelas bug di mana span hilang karena accessor yang salah terpakai.
	Context() context.Context

	// Param mengembalikan path parameter, mis. "id" untuk /users/:id.
	Param(key string) string

	// Query mengembalikan query string parameter.
	Query(key string) string

	// Header mengembalikan header request.
	Header(key string) string

	// Body men-decode body request ke dst. Bentuk decoding-nya ditentukan
	// implementasi (Fiber memakai BodyParser yang sadar Content-Type).
	Body(dst any) error

	// FormFile mengembalikan file dari multipart form.
	FormFile(key string) (*multipart.FileHeader, error)

	// Principal mengembalikan aktor terautentikasi. Ok bernilai false kalau
	// request tidak terautentikasi. Mengembalikan value, bukan pointer, supaya
	// tidak ada jalur nil-dereference.
	Principal() (actor entity.Actor, ok bool)
}
