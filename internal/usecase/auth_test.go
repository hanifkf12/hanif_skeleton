package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	jwtpkg "github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/stretchr/testify/require"
)

type authUserRepository struct {
	user        *entity.User
	err         error
	createdUser *entity.CreateUserRequest
	createID    int64
}

func (r *authUserRepository) GetUserByUsername(context.Context, string) (*entity.User, error) {
	return r.user, r.err
}
func (r *authUserRepository) GetUsers(context.Context) ([]entity.User, error) { return nil, nil }
func (r *authUserRepository) GetUserByID(context.Context, int64) (*entity.User, error) {
	return nil, nil
}
func (r *authUserRepository) CreateUser(_ context.Context, user entity.CreateUserRequest) (int64, error) {
	r.createdUser = &user
	return r.createID, r.err
}

func TestLoginRejectsIncorrectPassword(t *testing.T) {
	hasher := crypto.NewBcryptHasher(4)
	passwordHash, err := hasher.HashPassword("correct-password")
	require.NoError(t, err)
	service, err := jwtpkg.NewJWT(jwtpkg.Config{SecretKey: "test-secret"})
	require.NoError(t, err)
	login := NewLogin(&authUserRepository{user: &entity.User{
		ID: 1, Username: "alice", PasswordHash: passwordHash, Role: "user",
	}}, hasher, service)

	app := fiber.New()
	app.Post("/login", func(ctx *fiber.Ctx) error {
		resp := login.Serve(appctx.Data{FiberCtx: ctx})
		return ctx.Status(resp.Code).JSON(resp)
	})
	req := httptest.NewRequest(fiber.MethodPost, "/login",
		strings.NewReader(`{"username":"alice","password":"wrong-password"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateUserHashesPasswordBeforePersistence(t *testing.T) {
	repo := &authUserRepository{createID: 42}
	hasher := crypto.NewBcryptHasher(4)
	createUser := NewCreateUser(repo, hasher)

	app := fiber.New()
	app.Post("/users", func(ctx *fiber.Ctx) error {
		resp := createUser.Serve(appctx.Data{FiberCtx: ctx})
		return ctx.Status(resp.Code).JSON(resp)
	})
	req := httptest.NewRequest(fiber.MethodPost, "/users",
		strings.NewReader(`{"username":"alice","email":"alice@example.com","password":"plain-secret"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	require.NotNil(t, repo.createdUser)
	require.NotEqual(t, "plain-secret", repo.createdUser.Password)
	require.True(t, hasher.ComparePassword("plain-secret", repo.createdUser.Password))
}

func TestUserCreatedConsumerHashesPasswordBeforePersistence(t *testing.T) {
	repo := &authUserRepository{createID: 42}
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
	require.NotEqual(t, "plain-secret", repo.createdUser.Password)
	require.True(t, hasher.ComparePassword("plain-secret", repo.createdUser.Password))
}
func (r *authUserRepository) UpdateUser(context.Context, entity.UpdateUserRequest) error { return nil }
func (r *authUserRepository) DeleteUser(context.Context, int64) error                    { return nil }

func TestLoginDoesNotIssueTokenForUnknownUser(t *testing.T) {
	service, err := jwtpkg.NewJWT(jwtpkg.Config{SecretKey: "test-secret"})
	require.NoError(t, err)
	login := NewLogin(&authUserRepository{err: sql.ErrNoRows}, crypto.NewBcryptHasher(4), service)

	app := fiber.New()
	app.Post("/login", func(ctx *fiber.Ctx) error {
		resp := login.Serve(appctx.Data{FiberCtx: ctx})
		return ctx.Status(resp.Code).JSON(resp)
	})
	req := httptest.NewRequest(fiber.MethodPost, "/login",
		strings.NewReader(`{"username":"unknown","password":"anything"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var body appctx.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Nil(t, body.Data)
}
