package entity

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	// max=72 adalah batas keras bcrypt: x/crypto/bcrypt mengembalikan
	// ErrPasswordTooLong untuk input yang lebih panjang. Tanpa batas ini
	// password 73 byte menjadi error 500, padahal itu kesalahan input klien.
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type CreateUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
