package master

import (
	"context"
	"database/sql"

	"github.com/kemnaker/perjadin-backend/internal/db"
)

type Repository interface {
	GetProvinces(ctx context.Context) ([]db.Province, error)
	GetSBMRates(ctx context.Context) ([]db.SbmRate, error)
}

type repository struct {
	q *db.Queries
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		q: db.New(sqlDB),
	}
}

func (r *repository) GetProvinces(ctx context.Context) ([]db.Province, error) {
	return r.q.GetProvinces(ctx)
}

func (r *repository) GetSBMRates(ctx context.Context) ([]db.SbmRate, error) {
	return r.q.GetSBMRates(ctx)
}
