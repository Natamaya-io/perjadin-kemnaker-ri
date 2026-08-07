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
	UpdateTransaction(ctx context.Context, trx *models.GUPTransaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error

	GetBudgets(ctx context.Context, year int16) ([]models.Budget, error)
	GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error)
	SaveMonthlyLS(ctx context.Context, items []models.MonthlyLS) error
	GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error)
	GetLaporan(ctx context.Context, year int16) (*models.LaporanResponse, error)
	GetDashboardSummary(ctx context.Context, year int16) (*models.GupDashboardSummary, error)
	GetNextBusinessID(ctx context.Context) (string, error)
	SaveBudget(ctx context.Context, b *models.Budget) error
	CreateProcurementType(ctx context.Context, pt *models.ProcurementType) error
	UpdateProcurementType(ctx context.Context, pt *models.ProcurementType) error
	DeleteProcurementType(ctx context.Context, id uuid.UUID) error
	CreateAccountCode(ctx context.Context, ac *models.AccountCode) error
	UpdateAccountCode(ctx context.Context, ac *models.AccountCode) error
	DeleteAccountCode(ctx context.Context, id uuid.UUID) error
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

func (s *service) UpdateTransaction(ctx context.Context, trx *models.GUPTransaction) error {
	return s.repo.UpdateTransaction(ctx, trx)
}

func (s *service) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTransaction(ctx, id)
}

func (s *service) GetBudgets(ctx context.Context, year int16) ([]models.Budget, error) {
	return s.repo.GetBudgets(ctx, year)
}

func (s *service) GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error) {
	return s.repo.GetMonthlyLS(ctx, year)
}

func (s *service) SaveMonthlyLS(ctx context.Context, items []models.MonthlyLS) error {
	return s.repo.SaveMonthlyLS(ctx, items)
}

func (s *service) GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error) {
	return s.repo.GetMasterData(ctx, year)
}

func (s *service) CreateProcurementType(ctx context.Context, pt *models.ProcurementType) error {
	return s.repo.CreateProcurementType(ctx, pt)
}

func (s *service) UpdateProcurementType(ctx context.Context, pt *models.ProcurementType) error {
	return s.repo.UpdateProcurementType(ctx, pt)
}

func (s *service) DeleteProcurementType(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProcurementType(ctx, id)
}

func (s *service) CreateAccountCode(ctx context.Context, ac *models.AccountCode) error {
	return s.repo.CreateAccountCode(ctx, ac)
}

func (s *service) UpdateAccountCode(ctx context.Context, ac *models.AccountCode) error {
	return s.repo.UpdateAccountCode(ctx, ac)
}

func (s *service) DeleteAccountCode(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAccountCode(ctx, id)
}

// GetLaporan mengambil data rekapitulasi per Jenis Pengadaan beserta ringkasan total.
func (s *service) GetLaporan(ctx context.Context, year int16) (*models.LaporanResponse, error) {
	rows, err := s.repo.GetLaporanRows(ctx, year)
	if err != nil {
		return nil, err
	}

	summary := models.LaporanSummary{}
	uniqueProcurements := make(map[uuid.UUID]bool)

	for _, row := range rows {
		summary.TotalAnggaran += row.Anggaran
		summary.TotalRealisasi += row.Realisasi
		summary.TotalTransaksi += row.JumlahTransaksi
		uniqueProcurements[row.ProcurementTypeID] = true
	}

	summary.SisaAnggaran = summary.TotalAnggaran - summary.TotalRealisasi
	if summary.TotalAnggaran > 0 {
		summary.PersentaseSerapan = (summary.TotalRealisasi / summary.TotalAnggaran) * 100
	}
	summary.TotalJenisPengadaan = len(uniqueProcurements)

	return &models.LaporanResponse{
		Rows:    rows,
		Summary: summary,
	}, nil
}

func (s *service) GetDashboardSummary(ctx context.Context, year int16) (*models.GupDashboardSummary, error) {
	return s.repo.GetDashboardSummary(ctx, year)
}

func (s *service) GetNextBusinessID(ctx context.Context) (string, error) {
	return s.repo.GetNextBusinessID(ctx)
}

// SaveBudget menyimpan atau memperbarui anggaran untuk satu jenis pengadaan.
func (s *service) SaveBudget(ctx context.Context, b *models.Budget) error {
	return s.repo.SaveBudget(ctx, b)
}
