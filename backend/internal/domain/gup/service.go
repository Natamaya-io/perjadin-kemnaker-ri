package gup

import (
	"context"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

type Service interface {
	GetTransactions(ctx context.Context) ([]models.GUPTransaction, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.GUPTransaction, error)
	CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error
	
	GetBudgets(ctx context.Context, year int16) ([]models.Budget, error)
	GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetTransactions(ctx context.Context) ([]models.GUPTransaction, error) {
	return s.repo.GetTransactions(ctx)
}

func (s *service) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.GUPTransaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

func (s *service) CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error {
	return s.repo.CreateTransaction(ctx, trx)
}

func (s *service) GetBudgets(ctx context.Context, year int16) ([]models.Budget, error) {
	return s.repo.GetBudgets(ctx, year)
}

func (s *service) GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error) {
	return s.repo.GetMonthlyLS(ctx, year)
}
