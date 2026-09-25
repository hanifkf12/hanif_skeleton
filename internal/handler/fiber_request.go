package handler

import (
	"context"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

// fiberRequest adalah satu-satunya tempat di codebase yang menerjemahkan
// *fiber.Ctx menjadi appctx.Request. Semua detail Fiber berhenti di sini.
type fiberRequest struct {
	ctx *fiber.Ctx
}

// NewFiberRequest membungkus *fiber.Ctx sebagai appctx.Request.
func NewFiberRequest(ctx *fiber.Ctx) appctx.Request {
	return &fiberRequest{ctx: ctx}
}

// Context memakai UserContext(), BUKAN Context().
//
// fiber.Ctx.Context() mengembalikan *fasthttp.RequestCtx — bukan context.Context
// yang membawa span OpenTelemetry yang dipasang TraceMiddleware — sehingga
// memakainya akan diam-diam menghilangkan trace. UserContext() tidak pernah nil
// (Fiber mengisinya dengan context.Background() bila belum di-set), jadi
// telemetry.StartSpan tidak bisa panic.
func (r *fiberRequest) Context() context.Context {
	return r.ctx.UserContext()
}

func (r *fiberRequest) Param(key string) string  { return r.ctx.Params(key) }
func (r *fiberRequest) Query(key string) string  { return r.ctx.Query(key) }
func (r *fiberRequest) Header(key string) string { return r.ctx.Get(key) }

func (r *fiberRequest) Body(dst any) error { return r.ctx.BodyParser(dst) }

func (r *fiberRequest) FormFile(key string) (*multipart.FileHeader, error) {
	return r.ctx.FormFile(key)
}

// Principal membaca identitas dari Locals yang di-set middleware JWT.
// Mengembalikan ok=false kalau request tidak terautentikasi.
func (r *fiberRequest) Principal() (entity.Actor, bool) {
	userID, ok := r.ctx.Locals(appctx.LocalsUserID).(int64)
	if !ok || userID <= 0 {
		return entity.Actor{}, false
	}

	// Role diambil sebagai best-effort: claim yang absen menghasilkan role
	// kosong, yang tidak pernah cocok dengan RoleAdmin maupun RoleSuperAdmin.
	role, _ := r.ctx.Locals(appctx.LocalsRole).(string)

	return entity.NewActor(userID, role), true
}

var _ appctx.Request = (*fiberRequest)(nil)
