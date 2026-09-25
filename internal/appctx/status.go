package appctx

// Konstanta status ini mencerminkan kode status HTTP tanpa bergantung pada
// framework HTTP mana pun. Usecase, repository, dan entity hanya boleh memakai
// konstanta ini — bukan fiber.StatusX — supaya business layer tidak terikat
// pada Fiber (Clean Architecture Dependency Rule).
//
// Nilainya sengaja sama persis dengan kode HTTP karena Response.Code dikirim
// apa adanya sebagai status code oleh router.
const (
	StatusOK                   = 200
	StatusCreated              = 201
	StatusBadRequest           = 400
	StatusUnauthorized         = 401
	StatusPaymentRequired      = 402
	StatusForbidden            = 403
	StatusNotFound             = 404
	StatusConflict             = 409
	StatusUnsupportedMediaType = 415
	StatusTooManyRequests      = 429
	StatusInternalServerError  = 500
	StatusServiceUnavailable   = 503
)
