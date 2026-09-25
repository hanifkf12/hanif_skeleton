package entity

// Peran pengguna. Nilai string-nya harus tetap sinkron dengan yang di-hardcode
// di pkg/jwt (Claims.IsAdmin) — pkg/jwt tidak boleh meng-import internal/entity,
// jadi keduanya terpisah dan harus diubah bersamaan.
const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superadmin"
)

// Actor adalah principal terautentikasi yang melakukan sebuah aksi. Dibangun
// dari klaim JWT, bukan dari baris database — lihat canAccessUser.
type Actor struct {
	UserID int64
	Role   string
}

func NewActor(userID int64, role string) Actor {
	return Actor{UserID: userID, Role: role}
}

// IsAdmin menganggap admin dan superadmin setara.
func (a Actor) IsAdmin() bool {
	return a.Role == RoleAdmin || a.Role == RoleSuperAdmin
}

// CanAccessUser menentukan apakah actor boleh membaca data targetUserID.
//
// Guard UserID <= 0 sengaja fail-closed: adapter yang buggy dan menghasilkan
// Actor{UserID: 0, Role: "admin"} akan mendapat 403, bukan lolos.
//
// Actor tidak boleh dibangun dari entity.User hasil query: GetUserByID dan
// GetUsers tidak men-SELECT kolom role, sehingga User.Role selalu kosong dan
// pemeriksaan berbasis role akan gagal-terbuka.
func (a Actor) CanAccessUser(targetUserID int64) bool {
	if a.UserID <= 0 || targetUserID <= 0 {
		return false
	}

	return a.IsAdmin() || a.UserID == targetUserID
}

// CanModifyUser menentukan apakah actor boleh mengubah data targetUserID.
//
// Saat ini identik dengan CanAccessUser; dipisah supaya keduanya bisa berbeda
// nanti (mis. "boleh baca semua, hanya boleh ubah sendiri") tanpa mengubah
// pemanggilnya.
func (a Actor) CanModifyUser(targetUserID int64) bool {
	return a.CanAccessUser(targetUserID)
}
