package entity

import (
	"errors"
	"strings"
)

var (
	ErrUserMissingUsername = errors.New("user: username is required")
	ErrUserMissingEmail    = errors.New("user: email is required")
	ErrUserMissingPassword = errors.New("user: password hash is required")
)

// UserInput adalah input domain untuk membuat user.
//
// PasswordHash HARUS sudah berupa hash — tipe ini tidak punya cara untuk
// merepresentasikan password plaintext, dan itu memang tujuannya: repository
// tidak bisa lagi menerima password mentah dari layer transport.
type UserInput struct {
	Username     string
	Email        string
	PasswordHash string
}

// NewUserInput menormalkan input dan menegakkan invariant domain.
// Kolom `name` tidak ada di sini karena repository mengisinya dari Username.
func NewUserInput(username, email, passwordHash string) (UserInput, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	switch {
	case username == "":
		return UserInput{}, ErrUserMissingUsername
	case email == "":
		return UserInput{}, ErrUserMissingEmail
	case passwordHash == "":
		return UserInput{}, ErrUserMissingPassword
	}

	return UserInput{Username: username, Email: email, PasswordHash: passwordHash}, nil
}

// UserUpdate merepresentasikan perubahan parsial terhadap user.
//
// Field bernilai nil berarti "jangan diubah". Pointer dipilih supaya repository
// tetap hanya menulis kolom yang benar-benar di-set — memakai *User utuh akan
// mengosongkan role, name, dan password_hash pada update parsial.
type UserUpdate struct {
	Username     *string
	Email        *string
	PasswordHash *string
}

// IsEmpty melaporkan apakah tidak ada satu pun kolom yang akan diubah.
func (u UserUpdate) IsEmpty() bool {
	return u.Username == nil && u.Email == nil && u.PasswordHash == nil
}
