package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx/appctxtest"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	jwtpkg "github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestLoginRejectsIncorrectPassword(t *testing.T) {
	hasher := crypto.NewBcryptHasher(4)
	passwordHash, err := hasher.HashPassword("correct-password")
	require.NoError(t, err)
	service, err := jwtpkg.NewJWT(jwtpkg.Config{SecretKey: "test-secret"})
	require.NoError(t, err)
	login := NewLogin(&fakeUserRepository{user: &entity.User{
		ID: 1, Username: "alice", PasswordHash: passwordHash, Role: "user",
	}}, hasher, service)

	resp := login.Serve(appctx.Data{Request: appctxtest.New().
		WithJSONBody(map[string]string{"username": "alice", "password": "wrong-password"})})

	require.Equal(t, appctx.StatusUnauthorized, resp.Code)
	require.Nil(t, resp.Data)
}

func TestCreateUserHashesPasswordBeforePersistence(t *testing.T) {
	repo := &fakeUserRepository{createID: 42}
	hasher := crypto.NewBcryptHasher(4)
	createUser := NewCreateUser(repo, hasher)

	resp := createUser.Serve(appctx.Data{Request: appctxtest.New().
		WithJSONBody(map[string]string{
			"username": "alice",
			"email":    "alice@example.com",
			"password": "plain-secret",
		})})

	require.Equal(t, appctx.StatusCreated, resp.Code)
	require.NotNil(t, repo.createdUser)
	require.NotEqual(t, "plain-secret", repo.createdUser.PasswordHash)
	require.True(t, hasher.ComparePassword("plain-secret", repo.createdUser.PasswordHash))
}

// UserInput tidak punya field plaintext sama sekali, jadi password mentah
// secara struktural tidak bisa sampai ke repository.
func TestCreateUserRejectsPasswordBeyondBcryptLimit(t *testing.T) {
	repo := &fakeUserRepository{createID: 42}
	createUser := NewCreateUser(repo, crypto.NewBcryptHasher(4))

	resp := createUser.Serve(appctx.Data{Request: appctxtest.New().
		WithJSONBody(map[string]string{
			"username": "alice",
			"email":    "alice@example.com",
			"password": string(make([]byte, 73)),
		})})

	require.Equal(t, appctx.StatusBadRequest, resp.Code)
	require.Nil(t, repo.createdUser)
}

func TestUserCreatedConsumerHashesPasswordBeforePersistence(t *testing.T) {
	repo := &fakeUserRepository{createID: 42}
	hasher := crypto.NewBcryptHasher(4)
	consumer := NewUserCreatedConsumer(repo, hasher)

	resp := consumer.Consume(appctx.PubSubData{
		Ctx: context.Background(),
		Message: &pubsub.Message{Data: []byte(
			`{"username":"alice","email":"alice@example.com","password":"plain-secret"}`,
		)},
	})

	require.True(t, resp.Success)
	require.NotNil(t, repo.createdUser)
	require.NotEqual(t, "plain-secret", repo.createdUser.PasswordHash)
	require.True(t, hasher.ComparePassword("plain-secret", repo.createdUser.PasswordHash))
}

// User yang tidak dikenal harus tetap dijawab 401 "Invalid credentials",
// bukan 404 — supaya keberadaan akun tidak terbocor lewat kode status.
func TestLoginDoesNotIssueTokenForUnknownUser(t *testing.T) {
	service, err := jwtpkg.NewJWT(jwtpkg.Config{SecretKey: "test-secret"})
	require.NoError(t, err)

	tests := []struct {
		name string
		err  error
	}{
		{"baris tidak ditemukan", sql.ErrNoRows},
		{"error terklasifikasi not found", apperror.NotFound("user not found")},
		{"error database tak terduga", errors.New("connection reset by peer")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			login := NewLogin(&fakeUserRepository{userErr: tt.err}, crypto.NewBcryptHasher(4), service)

			resp := login.Serve(appctx.Data{Request: appctxtest.New().
				WithJSONBody(map[string]string{"username": "unknown", "password": "anything"})})

			require.Equal(t, appctx.StatusUnauthorized, resp.Code)
			require.Equal(t, "Invalid credentials", resp.Errors)
			require.Nil(t, resp.Data)
		})
	}
}
