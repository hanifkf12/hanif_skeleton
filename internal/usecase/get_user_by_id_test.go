package usecase

import (
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx/appctxtest"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/stretchr/testify/require"
)

// Jalur baca dari celah yang sama: GET /users/:id tidak punya pemeriksaan
// otorisasi sama sekali, sehingga setiap pemegang token bisa membaca data
// user lain.
func TestGetUserByIDAuthorization(t *testing.T) {
	tests := []struct {
		name          string
		actor         entity.Actor
		authenticated bool
		targetID      string
		wantCode      int
		wantReads     int
	}{
		{
			name:          "user biasa membaca dirinya sendiri",
			actor:         entity.Actor{UserID: 7, Role: entity.RoleUser},
			authenticated: true,
			targetID:      "7",
			wantCode:      appctx.StatusOK,
			wantReads:     1,
		},
		{
			name:          "user biasa membaca user lain",
			actor:         entity.Actor{UserID: 7, Role: entity.RoleUser},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusForbidden,
			wantReads:     0,
		},
		{
			name:          "admin membaca user lain",
			actor:         entity.Actor{UserID: 1, Role: entity.RoleAdmin},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusOK,
			wantReads:     1,
		},
		{
			name:          "tanpa autentikasi",
			authenticated: false,
			targetID:      "9",
			wantCode:      appctx.StatusUnauthorized,
			wantReads:     0,
		},
		{
			name:          "admin tanpa user_id",
			actor:         entity.Actor{UserID: 0, Role: entity.RoleAdmin},
			authenticated: true,
			targetID:      "9",
			wantCode:      appctx.StatusForbidden,
			wantReads:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{byID: &entity.User{ID: 9, Username: "bob"}}
			service := NewGetUserByID(repo)

			req := appctxtest.New().WithParam("id", tt.targetID)
			if tt.authenticated {
				req = req.WithPrincipal(tt.actor)
			}

			resp := service.Serve(appctx.Data{Request: req})

			require.Equal(t, tt.wantCode, resp.Code)
			require.Equal(t, tt.wantReads, repo.getByIDCalls,
				"percobaan yang ditolak tidak boleh menyentuh database")
		})
	}
}

func TestGetUserByIDMapsNotFound(t *testing.T) {
	repo := &fakeUserRepository{byIDErr: apperror.NotFound("User not found")}
	service := NewGetUserByID(repo)

	resp := service.Serve(appctx.Data{Request: appctxtest.New().
		WithParam("id", "7").
		WithPrincipal(entity.Actor{UserID: 7, Role: entity.RoleUser})})

	require.Equal(t, appctx.StatusNotFound, resp.Code)
	require.Equal(t, "User not found", resp.Errors)
}
