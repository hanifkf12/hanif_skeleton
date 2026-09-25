package entity_test

import (
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestActorCanAccessUser(t *testing.T) {
	tests := []struct {
		name         string
		actor        entity.Actor
		targetUserID int64
		want         bool
	}{
		{
			name:         "user dapat mengakses dirinya sendiri",
			actor:        entity.NewActor(7, entity.RoleUser),
			targetUserID: 7,
			want:         true,
		},
		{
			name:         "user tidak dapat mengakses user lain",
			actor:        entity.NewActor(7, entity.RoleUser),
			targetUserID: 9,
			want:         false,
		},
		{
			name:         "admin dapat mengakses user lain",
			actor:        entity.NewActor(1, entity.RoleAdmin),
			targetUserID: 9,
			want:         true,
		},
		{
			name:         "superadmin dapat mengakses user lain",
			actor:        entity.NewActor(1, entity.RoleSuperAdmin),
			targetUserID: 9,
			want:         true,
		},
		{
			name:         "role tidak dikenal tidak mendapat akses istimewa",
			actor:        entity.NewActor(7, "manager"),
			targetUserID: 9,
			want:         false,
		},
		{
			name:         "role kosong diperlakukan sebagai user biasa",
			actor:        entity.NewActor(7, ""),
			targetUserID: 7,
			want:         true,
		},
		// Guard fail-closed: actor tanpa UserID valid tidak boleh lolos,
		// walaupun rolenya admin. Ini melindungi dari adapter yang buggy.
		{
			name:         "actor tanpa user id ditolak walau admin",
			actor:        entity.NewActor(0, entity.RoleAdmin),
			targetUserID: 9,
			want:         false,
		},
		{
			name:         "actor kosong ditolak",
			actor:        entity.Actor{},
			targetUserID: 9,
			want:         false,
		},
		{
			name:         "target user id tidak valid ditolak walau actor admin",
			actor:        entity.NewActor(1, entity.RoleAdmin),
			targetUserID: 0,
			want:         false,
		},
		{
			name:         "target user id negatif ditolak",
			actor:        entity.NewActor(1, entity.RoleAdmin),
			targetUserID: -3,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.actor.CanAccessUser(tt.targetUserID))
			assert.Equal(t, tt.want, tt.actor.CanModifyUser(tt.targetUserID),
				"CanModifyUser harus konsisten dengan CanAccessUser")
		})
	}
}
