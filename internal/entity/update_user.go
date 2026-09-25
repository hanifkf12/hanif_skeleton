package entity

// UpdateUserRequest adalah bentuk wire body untuk PUT /users/:id.
// Tidak ada field ID di sini dengan sengaja: id datang dari path, dan otorisasi
// dibandingkan terhadap path itu. Field ID di body hanya akan menciptakan
// dua sumber kebenaran yang bisa berbeda.
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	// Lihat catatan batas bcrypt di CreateUserRequest.Password.
	Password string `json:"password,omitempty" validate:"omitempty,min=6,max=72"`
}

type UpdateUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
