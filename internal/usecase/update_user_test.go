package usecase

import (
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx/appctxtest"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/stretchr/testify/require"
)

// TestUpdateUserAuthorization adalah test regresi untuk celah IDOR di
// PUT /users/:id: sebelumnya setiap pemegang token yang valid bisa mengubah
// email dan password user lain, termasuk admin.
//
// Assert utamanya adalah updateCalls == 0 pada kasus yang ditolak. Assert kode
// status saja tidak akan menangkap regresi di mana usecase menulis lebih dulu
// lalu melaporkan 403.
func TestUpdateUserAuthorization(t *testing.T) {
	const selfID = int64(7)
	const otherID = int64(9)

	tests := []struct {
		name          string
		actor         entity.Actor
		authenticated bool
		targetID      string
		wantCode      int
		wantWrites    int
	}{
		{
			name:          "user biasa mengubah akun sendiri",
			actor:         entity.Actor{UserID: selfID, Role: entity.RoleUser},
			authenticated: true,
			targetID:      "7",
			wantCode:      appctx.StatusOK,
			wantWrites:    1,
		},
		{
			name:          "user biasa mengubah akun orang lain",
			actor:         entity.Actor{UserID: selfID, Role: entity.RoleUser},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusForbidden,
			wantWrites:    0,
		},
		{
			name:          "admin mengubah akun orang lain",
			actor:         entity.Actor{UserID: 1, Role: entity.RoleAdmin},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusOK,
			wantWrites:    1,
		},
		{
			name:          "superadmin mengubah akun orang lain",
			actor:         entity.Actor{UserID: 2, Role: entity.RoleSuperAdmin},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusOK,
			wantWrites:    1,
		},
		{
			name:          "tanpa autentikasi",
			authenticated: false,
			targetID:      "9",
			wantCode:      appctx.StatusUnauthorized,
			wantWrites:    0,
		},
		{
			// Actor yang dibangun adapter yang buggy: user_id tidak terbaca
			// tetapi role-nya admin. Harus gagal-tertutup.
			name:          "admin tanpa user_id",
			actor:         entity.Actor{UserID: 0, Role: entity.RoleAdmin},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusForbidden,
			wantWrites:    0,
		},
		{
			name:          "id bukan angka",
			actor:         entity.Actor{UserID: selfID, Role: entity.RoleUser},
			authenticated: true,
			targetID:      "abc",
			wantCode:      appctx.StatusBadRequest,
			wantWrites:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{}
			service := NewUpdateUser(repo, crypto.NewBcryptHasher(4))

			req := appctxtest.New().
				WithParam("id", tt.targetID).
				WithJSONBody(map[string]string{"email": "changed@example.com"})
			if tt.authenticated {
				req = req.WithPrincipal(tt.actor)
			}

			resp := service.Serve(appctx.Data{Request: req})

			require.Equal(t, tt.wantCode, resp.Code)
			require.Equal(t, tt.wantWrites, repo.updateCalls,
				"percobaan yang ditolak tidak boleh menyentuh database")
		})
	}
}

func TestUpdateUserHashesPasswordBeforePersistence(t *testing.T) {
	repo := &fakeUserRepository{}
	hasher := crypto.NewBcryptHasher(4)
	service := NewUpdateUser(repo, hasher)

	resp := service.Serve(appctx.Data{Request: appctxtest.New().
		WithParam("id", "7").
		WithPrincipal(entity.Actor{UserID: 7, Role: entity.RoleUser}).
		WithJSONBody(map[string]string{"password": "new-plain-secret"})})

	require.Equal(t, appctx.StatusOK, resp.Code)
	require.Equal(t, 1, repo.updateCalls)
	require.NotNil(t, repo.updatedWith)
	require.NotNil(t, repo.updatedWith.PasswordHash)
	require.NotEqual(t, "new-plain-secret", *repo.updatedWith.PasswordHash)
	require.True(t, hasher.ComparePassword("new-plain-secret", *repo.updatedWith.PasswordHash))
	// Field yang tidak dikirim harus tetap nil, supaya repository tidak
	// menimpanya dengan string kosong.
	require.Nil(t, repo.updatedWith.Username)
	require.Nil(t, repo.updatedWith.Email)
}

// Body kosong tidak boleh menjadi UPDATE tanpa SET.
func TestUpdateUserRejectsEmptyUpdate(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUpdateUser(repo, crypto.NewBcryptHasher(4))

	resp := service.Serve(appctx.Data{Request: appctxtest.New().
		WithParam("id", "7").
		WithPrincipal(entity.Actor{UserID: 7, Role: entity.RoleUser}).
		WithJSONBody(map[string]string{})})

	require.Equal(t, appctx.StatusBadRequest, resp.Code)
	require.Equal(t, 0, repo.updateCalls)
}
