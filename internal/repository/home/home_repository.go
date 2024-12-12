package home

import (
	"context"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
)

type homeRepository struct {
	db databasex.Database
}

func (h homeRepository) GetAdmin(ctx context.Context, data any) ([]entity.Admin, error) {
	query := `SELECT * FROM Admin;`
	var result = make([]entity.Admin, 0)

	err := h.db.Select(ctx, &result, query)
	if err != nil {
		return result, err
	}
	return result, nil
}

func NewHomeRepository(db databasex.Database) repository.HomeRepository {
	return &homeRepository{db: db}
}
