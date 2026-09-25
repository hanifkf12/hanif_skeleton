package appctx

// Key Locals yang di-set oleh middleware JWT dan dibaca oleh adapter request.
// Dikumpulkan di sini supaya middleware dan adapter tidak bisa drift satu sama
// lain. Nilainya string polos karena appctx tidak boleh meng-import framework
// HTTP mana pun.
const (
	LocalsUserID = "user_id"
	LocalsRole   = "role"
	LocalsClaims = "claims"
)
