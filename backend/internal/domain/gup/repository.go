package gup

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

type Repository interface {
	GetTransactions(ctx context.Context) ([]models.GUPTransaction, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.GUPTransaction, error)
	CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error
	
	GetBudgets(ctx context.Context, year int16) ([]models.Budget, error)
	GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error)
	GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error)
}

type repository struct {
	d *sql.DB
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		d: sqlDB,
	}
}

func (r *repository) GetTransactions(ctx context.Context) ([]models.GUPTransaction, error) {
	query := `
		SELECT 
			gt.id, gt.business_id, gt.payment_description, gt.procurement_type_id, 
			gt.funding_source_id, gt.value_amount, gt.paid_amount, gt.tax_amount, 
			gt.receipt_date, gt.recipient, gt.pum, gt.created_at, gt.updated_at,
			pt.name as procurement_type_name
		FROM gup_transactions gt
		JOIN procurement_types pt ON gt.procurement_type_id = pt.id
		ORDER BY gt.created_at DESC
	`
	rows, err := r.d.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.GUPTransaction
	for rows.Next() {
		var trx models.GUPTransaction
		var fundingSourceID sql.NullString
		
		err := rows.Scan(
			&trx.ID, &trx.BusinessID, &trx.PaymentDescription, &trx.ProcurementTypeID,
			&fundingSourceID, &trx.ValueAmount, &trx.PaidAmount, &trx.TaxAmount,
			&trx.ReceiptDate, &trx.Recipient, &trx.Pum, &trx.CreatedAt, &trx.UpdatedAt,
			&trx.ProcurementTypeName,
		)
		if err != nil {
			return nil, err
		}
		
		if fundingSourceID.Valid {
			if id, err := uuid.Parse(fundingSourceID.String); err == nil {
				trx.FundingSourceID = &id
			}
		}
		
		results = append(results, trx)
	}
	return results, nil
}

func (r *repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.GUPTransaction, error) {
	query := `
		SELECT 
			gt.id, gt.business_id, gt.payment_description, gt.procurement_type_id, 
			gt.funding_source_id, gt.value_amount, gt.paid_amount, gt.tax_amount, 
			gt.receipt_date, gt.recipient, gt.pum, gt.created_at, gt.updated_at,
			pt.name as procurement_type_name
		FROM gup_transactions gt
		JOIN procurement_types pt ON gt.procurement_type_id = pt.id
		WHERE gt.id = $1
	`
	row := r.d.QueryRowContext(ctx, query, id)
	
	var trx models.GUPTransaction
	var fundingSourceID sql.NullString
	
	err := row.Scan(
		&trx.ID, &trx.BusinessID, &trx.PaymentDescription, &trx.ProcurementTypeID,
		&fundingSourceID, &trx.ValueAmount, &trx.PaidAmount, &trx.TaxAmount,
		&trx.ReceiptDate, &trx.Recipient, &trx.Pum, &trx.CreatedAt, &trx.UpdatedAt,
		&trx.ProcurementTypeName,
	)
	if err != nil {
		return nil, err
	}
	
	if fundingSourceID.Valid {
		if id, err := uuid.Parse(fundingSourceID.String); err == nil {
			trx.FundingSourceID = &id
		}
	}
	
	return &trx, nil
}

func (r *repository) CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error {
	if trx.ID == uuid.Nil {
		trx.ID = uuid.New()
	}
	
	query := `
		INSERT INTO gup_transactions (
			id, business_id, payment_description, procurement_type_id, 
			funding_source_id, value_amount, paid_amount, tax_amount, 
			receipt_date, recipient, pum, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	
	now := time.Now()
	trx.CreatedAt = now
	trx.UpdatedAt = now
	
	var fsID interface{}
	if trx.FundingSourceID != nil {
		fsID = *trx.FundingSourceID
	}
	
	_, err := r.d.ExecContext(ctx, query,
		trx.ID, trx.BusinessID, trx.PaymentDescription, trx.ProcurementTypeID,
		fsID, trx.ValueAmount, trx.PaidAmount, trx.TaxAmount,
		trx.ReceiptDate, trx.Recipient, trx.Pum, trx.CreatedAt, trx.UpdatedAt,
	)
	return err
}

func (r *repository) GetBudgets(ctx context.Context, year int16) ([]models.Budget, error) {
	query := `
		SELECT b.id, b.year, b.procurement_type_id, b.amount, b.created_at, b.updated_at, pt.name
		FROM budgets b
		JOIN procurement_types pt ON b.procurement_type_id = pt.id
		WHERE b.year = $1
		ORDER BY pt.name ASC
	`
	rows, err := r.d.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Budget
	for rows.Next() {
		var b models.Budget
		if err := rows.Scan(
			&b.ID, &b.Year, &b.ProcurementTypeID, &b.Amount, &b.CreatedAt, &b.UpdatedAt, &b.ProcurementTypeName,
		); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, nil
}

func (r *repository) GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error) {
	query := `
		SELECT 
			ls.id, ls.funding_source_id, ls.amount, ls.created_at, ls.updated_at,
			fs.month_name, fs.gup_label
		FROM monthly_ls ls
		JOIN funding_sources fs ON ls.funding_source_id = fs.id
		WHERE fs.year = $1
		ORDER BY fs.month_number ASC
	`
	rows, err := r.d.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.MonthlyLS
	for rows.Next() {
		var m models.MonthlyLS
		if err := rows.Scan(
			&m.ID, &m.FundingSourceID, &m.Amount, &m.CreatedAt, &m.UpdatedAt, 
			&m.MonthName, &m.GupLabel,
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *repository) GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error) {
	// 1. Get FundingSources
	fsQuery := `SELECT id, year, month_number, month_name, gup_label, created_at, updated_at FROM funding_sources WHERE year = $1 ORDER BY month_number ASC`
	fsRows, err := r.d.QueryContext(ctx, fsQuery, year)
	if err != nil {
		return nil, err
	}
	defer fsRows.Close()

	var fundingSources []models.FundingSource
	for fsRows.Next() {
		var fs models.FundingSource
		if err := fsRows.Scan(&fs.ID, &fs.Year, &fs.MonthNumber, &fs.MonthName, &fs.GupLabel, &fs.CreatedAt, &fs.UpdatedAt); err != nil {
			return nil, err
		}
		fundingSources = append(fundingSources, fs)
	}

	// 2. Get ProcurementTypes with AccountCodes
	ptQuery := `
		SELECT pt.id, pt.account_code_id, pt.name, pt.is_active, pt.created_at, pt.updated_at,
		       ac.code, ac.mak
		FROM procurement_types pt
		JOIN account_codes ac ON pt.account_code_id = ac.id
		WHERE pt.is_active = true
		ORDER BY pt.name ASC
	`
	ptRows, err := r.d.QueryContext(ctx, ptQuery)
	if err != nil {
		return nil, err
	}
	defer ptRows.Close()

	var procurementTypes []models.ProcurementType
	for ptRows.Next() {
		var pt models.ProcurementType
		if err := ptRows.Scan(&pt.ID, &pt.AccountCodeID, &pt.Name, &pt.IsActive, &pt.CreatedAt, &pt.UpdatedAt, &pt.AccountCode, &pt.AccountMak); err != nil {
			return nil, err
		}
		procurementTypes = append(procurementTypes, pt)
	}

	return map[string]interface{}{
		"fundingSources":   fundingSources,
		"procurementTypes": procurementTypes,
	}, nil
}
