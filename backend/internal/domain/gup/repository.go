package gup

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

type Repository interface {
	GetTransactions(ctx context.Context) ([]models.GUPTransaction, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.GUPTransaction, error)
	CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error
	UpdateTransaction(ctx context.Context, trx *models.GUPTransaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error

	GetBudgets(ctx context.Context, year int16) ([]models.Budget, error)
	GetMonthlyLS(ctx context.Context, year int16) ([]models.MonthlyLS, error)
	SaveMonthlyLS(ctx context.Context, items []models.MonthlyLS) error
	GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error)
	GetLaporanRows(ctx context.Context, year int16, month int16) ([]models.LaporanRow, error)
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
			gt.receipt_date, gt.recipient, gt.pum, gt.document_file, gt.created_at, gt.updated_at,
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
		var docFile sql.NullString
		
		err := rows.Scan(
			&trx.ID, &trx.BusinessID, &trx.PaymentDescription, &trx.ProcurementTypeID,
			&fundingSourceID, &trx.ValueAmount, &trx.PaidAmount, &trx.TaxAmount,
			&trx.ReceiptDate, &trx.Recipient, &trx.Pum, &docFile, &trx.CreatedAt, &trx.UpdatedAt,
			&trx.ProcurementTypeName,
		)
		if err != nil {
			return nil, err
		}
		
		if docFile.Valid {
			trx.DocumentFile = []byte(docFile.String)
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
			gt.receipt_date, gt.recipient, gt.pum, gt.document_file, gt.created_at, gt.updated_at,
			pt.name as procurement_type_name
		FROM gup_transactions gt
		JOIN procurement_types pt ON gt.procurement_type_id = pt.id
		WHERE gt.id = $1
	`
	row := r.d.QueryRowContext(ctx, query, id)
	
	var trx models.GUPTransaction
	var fundingSourceID sql.NullString
	var docFile sql.NullString
	
	err := row.Scan(
		&trx.ID, &trx.BusinessID, &trx.PaymentDescription, &trx.ProcurementTypeID,
		&fundingSourceID, &trx.ValueAmount, &trx.PaidAmount, &trx.TaxAmount,
		&trx.ReceiptDate, &trx.Recipient, &trx.Pum, &docFile, &trx.CreatedAt, &trx.UpdatedAt,
		&trx.ProcurementTypeName,
	)
	if err != nil {
		return nil, err
	}
	
	if docFile.Valid {
		trx.DocumentFile = []byte(docFile.String)
	}
	
	if fundingSourceID.Valid {
		if id, err := uuid.Parse(fundingSourceID.String); err == nil {
			trx.FundingSourceID = &id
		}
	}
	
	return &trx, nil
}

func (r *repository) CreateTransaction(ctx context.Context, trx *models.GUPTransaction) error {
	// 1. Math Validation
	if math.Abs(trx.ValueAmount - (trx.PaidAmount + trx.TaxAmount)) > 0.01 {
		return fmt.Errorf("validasi gagal: nilai (%.2f) harus sama dengan jumlah dibayarkan (%.2f) + pajak (%.2f)", trx.ValueAmount, trx.PaidAmount, trx.TaxAmount)
	}

	// 2. Budget Validation
	year := trx.ReceiptDate.Year()
	var budgetAmount float64
	var totalUsedBudget float64
	
	err := r.d.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM budgets WHERE year = $1 AND procurement_type_id = $2", year, trx.ProcurementTypeID).Scan(&budgetAmount)
	if err != nil {
		return err
	}
	
	err = r.d.QueryRowContext(ctx, "SELECT COALESCE(SUM(paid_amount), 0) FROM gup_transactions WHERE EXTRACT(YEAR FROM receipt_date) = $1 AND procurement_type_id = $2", year, trx.ProcurementTypeID).Scan(&totalUsedBudget)
	if err != nil {
		return err
	}
	
	if budgetAmount - totalUsedBudget < trx.PaidAmount {
		return fmt.Errorf("validasi gagal: saldo pagu anggaran untuk jenis pengadaan ini tidak mencukupi (Sisa: %.2f, Diajukan: %.2f)", budgetAmount - totalUsedBudget, trx.PaidAmount)
	}

	// 3. Funding Source Validation
	if trx.FundingSourceID != nil {
		var fsAmount float64
		var fsUsed float64
		
		err = r.d.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM monthly_ls WHERE id = $1", *trx.FundingSourceID).Scan(&fsAmount)
		if err != nil {
			return err
		}
		
		err = r.d.QueryRowContext(ctx, "SELECT COALESCE(SUM(paid_amount), 0) FROM gup_transactions WHERE funding_source_id = $1", *trx.FundingSourceID).Scan(&fsUsed)
		if err != nil {
			return err
		}
		
		if fsAmount - fsUsed < trx.PaidAmount {
			return fmt.Errorf("validasi gagal: saldo dompet GUP yang dipilih tidak mencukupi (Sisa: %.2f, Diajukan: %.2f)", fsAmount - fsUsed, trx.PaidAmount)
		}
	}

	if trx.ID == uuid.Nil {
		trx.ID = uuid.New()
	}
	
	query := `
		INSERT INTO gup_transactions (
			id, business_id, payment_description, procurement_type_id, 
			funding_source_id, value_amount, paid_amount, tax_amount, 
			receipt_date, recipient, pum, document_file, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`
	
	now := time.Now()
	trx.CreatedAt = now
	trx.UpdatedAt = now
	
	var fsID interface{}
	if trx.FundingSourceID != nil {
		fsID = *trx.FundingSourceID
	}
	
	_, err = r.d.ExecContext(ctx, query,
		trx.ID, trx.BusinessID, trx.PaymentDescription, trx.ProcurementTypeID,
		fsID, trx.ValueAmount, trx.PaidAmount, trx.TaxAmount,
		trx.ReceiptDate, trx.Recipient, trx.Pum, trx.DocumentFile, trx.CreatedAt, trx.UpdatedAt,
	)
	return err
}

func (r *repository) UpdateTransaction(ctx context.Context, trx *models.GUPTransaction) error {
	// 1. Math Validation
	if math.Abs(trx.ValueAmount - (trx.PaidAmount + trx.TaxAmount)) > 0.01 {
		return fmt.Errorf("validasi gagal: nilai (%.2f) harus sama dengan jumlah dibayarkan (%.2f) + pajak (%.2f)", trx.ValueAmount, trx.PaidAmount, trx.TaxAmount)
	}
	
	// Skip detailed budget checking on update for simplicity (or can be implemented later)
	query := `
		UPDATE gup_transactions SET
			business_id = $2, payment_description = $3, procurement_type_id = $4,
			funding_source_id = $5, value_amount = $6, paid_amount = $7, tax_amount = $8,
			receipt_date = $9, recipient = $10, pum = $11, document_file = $12, updated_at = $13
		WHERE id = $1
	`
	trx.UpdatedAt = time.Now()
	
	var fsID interface{}
	if trx.FundingSourceID != nil {
		fsID = *trx.FundingSourceID
	}
	
	_, err := r.d.ExecContext(ctx, query,
		trx.ID, trx.BusinessID, trx.PaymentDescription, trx.ProcurementTypeID,
		fsID, trx.ValueAmount, trx.PaidAmount, trx.TaxAmount,
		trx.ReceiptDate, trx.Recipient, trx.Pum, trx.DocumentFile, trx.UpdatedAt,
	)
	return err
}

func (r *repository) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	_, err := r.d.ExecContext(ctx, "DELETE FROM gup_transactions WHERE id = $1", id)
	return err
}

func (r *repository) GetNextBusinessID(ctx context.Context) (string, error) {
	var lastID string
	query := `
		SELECT business_id 
		FROM gup_transactions 
		WHERE business_id ILIKE 'gup_%' 
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err := r.d.QueryRowContext(ctx, query).Scan(&lastID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "gup_001", nil
		}
		return "", err
	}

	// Parse the number from lastID (e.g., "gup_010")
	var seq int
	_, err = fmt.Sscanf(lastID, "gup_%d", &seq)
	if err != nil {
		// Fallback if format is weird but matches ILIKE
		return "gup_001", nil
	}

	seq++
	return fmt.Sprintf("gup_%03d", seq), nil
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
			ls.id, ls.funding_source_id, ls.account_code_id, ls.amount, ls.created_at, ls.updated_at,
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
			&m.ID, &m.FundingSourceID, &m.AccountCodeID, &m.Amount, &m.CreatedAt, &m.UpdatedAt, 
			&m.MonthName, &m.GupLabel,
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *repository) SaveMonthlyLS(ctx context.Context, items []models.MonthlyLS) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO monthly_ls (id, funding_source_id, account_code_id, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (funding_source_id) 
		DO UPDATE SET amount = EXCLUDED.amount, account_code_id = EXCLUDED.account_code_id, updated_at = NOW()
	`
	for _, item := range items {
		id := item.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		if _, err := tx.ExecContext(ctx, query, id, item.FundingSourceID, item.AccountCodeID, item.Amount); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *repository) GetMasterData(ctx context.Context, year int16) (map[string]interface{}, error) {
	// 1. Get FundingSources with Remaining Budget and has_income
	fsQuery := `
		SELECT 
			fs.id, fs.year, fs.month_number, fs.month_name, fs.gup_label, fs.created_at, fs.updated_at,
			COALESCE(ml.amount, 0) - COALESCE(
				(SELECT SUM(paid_amount) FROM gup_transactions WHERE funding_source_id = fs.id), 0
			) as remaining_budget,
			CASE WHEN ml.amount IS NOT NULL AND ml.amount > 0 THEN true ELSE false END as has_income
		FROM funding_sources fs
		LEFT JOIN monthly_ls ml ON ml.funding_source_id = fs.id
		WHERE fs.year = $1 
		ORDER BY fs.month_number ASC
	`
	fsRows, err := r.d.QueryContext(ctx, fsQuery, year)
	if err != nil {
		return nil, err
	}
	defer fsRows.Close()

	var fundingSources []models.FundingSource
	for fsRows.Next() {
		var fs models.FundingSource
		if err := fsRows.Scan(&fs.ID, &fs.Year, &fs.MonthNumber, &fs.MonthName, &fs.GupLabel, &fs.CreatedAt, &fs.UpdatedAt, &fs.RemainingBudget, &fs.HasIncome); err != nil {
			return nil, err
		}
		fundingSources = append(fundingSources, fs)
	}

	// 2. Get ProcurementTypes with AccountCodes and budget info
	ptQuery := `
		SELECT pt.id, pt.account_code_id, pt.name, pt.is_active, pt.created_at, pt.updated_at,
		       ac.code, ac.mak,
		       CASE WHEN b.amount IS NOT NULL AND b.amount > 0 THEN true ELSE false END as has_budget
		FROM procurement_types pt
		JOIN account_codes ac ON pt.account_code_id = ac.id
		LEFT JOIN budgets b ON b.procurement_type_id = pt.id AND b.year = $1
		WHERE pt.is_active = true
		ORDER BY pt.name ASC
	`
	ptRows, err := r.d.QueryContext(ctx, ptQuery, year)
	if err != nil {
		return nil, err
	}
	defer ptRows.Close()

	var procurementTypes []models.ProcurementType
	for ptRows.Next() {
		var pt models.ProcurementType
		if err := ptRows.Scan(&pt.ID, &pt.AccountCodeID, &pt.Name, &pt.IsActive, &pt.CreatedAt, &pt.UpdatedAt, &pt.AccountCode, &pt.AccountMak, &pt.HasBudget); err != nil {
			return nil, err
		}
		procurementTypes = append(procurementTypes, pt)
	}

	// 3. Get AccountCodes
	acQuery := `SELECT id, code, mak, description, created_at, updated_at FROM account_codes ORDER BY code ASC`
	acRows, err := r.d.QueryContext(ctx, acQuery)
	if err != nil {
		return nil, err
	}
	defer acRows.Close()

	var accountCodes []models.AccountCode
	for acRows.Next() {
		var ac models.AccountCode
		if err := acRows.Scan(&ac.ID, &ac.Code, &ac.Mak, &ac.Description, &ac.CreatedAt, &ac.UpdatedAt); err != nil {
			return nil, err
		}
		accountCodes = append(accountCodes, ac)
	}

	return map[string]interface{}{
		"fundingSources":   fundingSources,
		"procurementTypes": procurementTypes,
		"accountCodes":     accountCodes,
	}, nil
}

func (r *repository) CreateProcurementType(ctx context.Context, pt *models.ProcurementType) error {
	query := `
		INSERT INTO procurement_types (account_code_id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.d.QueryRowContext(ctx, query, pt.AccountCodeID, pt.Name, pt.IsActive).
		Scan(&pt.ID, &pt.CreatedAt, &pt.UpdatedAt)
}

func (r *repository) UpdateProcurementType(ctx context.Context, pt *models.ProcurementType) error {
	query := `
		UPDATE procurement_types
		SET account_code_id = $1, name = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.d.QueryRowContext(ctx, query, pt.AccountCodeID, pt.Name, pt.IsActive, pt.ID).
		Scan(&pt.UpdatedAt)
}

func (r *repository) DeleteProcurementType(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM procurement_types WHERE id = $1`
	_, err := r.d.ExecContext(ctx, query, id)
	return err
}

func (r *repository) CreateAccountCode(ctx context.Context, ac *models.AccountCode) error {
	query := `
		INSERT INTO account_codes (code, mak, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.d.QueryRowContext(ctx, query, ac.Code, ac.Mak, ac.Description).
		Scan(&ac.ID, &ac.CreatedAt, &ac.UpdatedAt)
}

func (r *repository) UpdateAccountCode(ctx context.Context, ac *models.AccountCode) error {
	query := `
		UPDATE account_codes
		SET code = $1, mak = $2, description = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.d.QueryRowContext(ctx, query, ac.Code, ac.Mak, ac.Description, ac.ID).
		Scan(&ac.UpdatedAt)
}

func (r *repository) DeleteAccountCode(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM account_codes WHERE id = $1`
	_, err := r.d.ExecContext(ctx, query, id)
	return err
}

// GetLaporanRows menghasilkan rekapitulasi per Jenis Pengadaan:
// menggabungkan procurement_types, account_codes, gup_transactions, dan budgets
// dalam satu query agregasi untuk keperluan halaman Laporan.
func (r *repository) GetLaporanRows(ctx context.Context, year int16, month int16) ([]models.LaporanRow, error) {
	query := `
		SELECT
			pt.id,
			pt.name                                           AS jenis_pengadaan,
			ac.code                                           AS kode_akun,
			ac.mak,
			COUNT(gt.id)                                      AS jumlah_transaksi,
			COALESCE(SUM(gt.paid_amount), 0) + 
			CASE 
				WHEN pt.name ILIKE '%Luar Kota%' OR pt.name ILIKE '%Dalam Negeri%' THEN 
					(SELECT COALESCE(SUM(total_cost), 0) 
					 FROM travel_records 
					 WHERE EXTRACT(YEAR FROM COALESCE(start_date, created_at)) = $1 
					 AND ($2 = 0 OR EXTRACT(MONTH FROM COALESCE(start_date, created_at)) = $2)
					 AND deleted_at IS NULL
					 AND status IN ('Approved', 'Completed')
					 AND type = 'luar_kota')
				WHEN pt.name ILIKE '%Luar Negeri%' THEN 
					(SELECT COALESCE(SUM(total_cost), 0) 
					 FROM travel_records 
					 WHERE EXTRACT(YEAR FROM COALESCE(start_date, created_at)) = $1 
					 AND ($2 = 0 OR EXTRACT(MONTH FROM COALESCE(start_date, created_at)) = $2)
					 AND deleted_at IS NULL
					 AND status IN ('Approved', 'Completed')
					 AND type = 'luar_negeri')
				WHEN pt.name ILIKE '%Dalam Kota%' THEN 
					(SELECT COALESCE(SUM(da.spj_cost + da.actual_cost), 0) 
					 FROM dalkot_assignments da 
					 JOIN dalkot_records dr ON da.dalkot_record_id = dr.id 
					 WHERE EXTRACT(YEAR FROM dr.execution_date) = $1 
					 AND ($2 = 0 OR EXTRACT(MONTH FROM dr.execution_date) = $2)
					 AND dr.status IN ('Approved', 'Completed'))
				ELSE 0 
			END                                               AS realisasi,
			COALESCE(SUM(gt.value_amount), 0)                 AS nilai_pengajuan,
			COALESCE(SUM(gt.tax_amount), 0)                   AS total_pajak,
			COALESCE(b.amount, 0)                             AS anggaran
		FROM procurement_types pt
		JOIN account_codes ac ON pt.account_code_id = ac.id
		LEFT JOIN gup_transactions gt ON gt.procurement_type_id = pt.id AND EXTRACT(YEAR FROM gt.receipt_date) = $1 AND ($2 = 0 OR EXTRACT(MONTH FROM gt.receipt_date) = $2)
		LEFT JOIN (
			SELECT procurement_type_id, SUM(amount) as amount 
			FROM budgets 
			WHERE year = $1 AND ($2 = 0 OR month_number = $2)
			GROUP BY procurement_type_id
		) b ON b.procurement_type_id = pt.id
		WHERE pt.is_active = true
		GROUP BY pt.id, pt.name, ac.code, ac.mak, b.amount
		ORDER BY pt.name ASC
	`
	rows, err := r.d.QueryContext(ctx, query, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.LaporanRow
	for rows.Next() {
		var row models.LaporanRow
		if err := rows.Scan(
			&row.ProcurementTypeID,
			&row.JenisPengadaan,
			&row.KodeAkun,
			&row.Mak,
			&row.JumlahTransaksi,
			&row.Realisasi,
			&row.NilaiPengajuan,
			&row.TotalPajak,
			&row.Anggaran,
		); err != nil {
			return nil, err
		}
		// Hitung sisa dan persentase serapan
		row.SisaAnggaran = row.Anggaran - row.Realisasi
		if row.Anggaran > 0 {
			row.PersentaseSerapan = (row.Realisasi / row.Anggaran) * 100
		}
		results = append(results, row)
	}
	return results, nil
}

// SaveBudget menyimpan anggaran per jenis pengadaan (upsert berdasarkan year + month_number + procurement_type_id).
func (r *repository) SaveBudget(ctx context.Context, b *models.Budget) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	query := `
		INSERT INTO budgets (id, year, month_number, procurement_type_id, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (year, month_number, procurement_type_id)
		DO UPDATE SET amount = EXCLUDED.amount, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.d.QueryRowContext(ctx, query, b.ID, b.Year, b.MonthNumber, b.ProcurementTypeID, b.Amount).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func (r *repository) GetDashboardSummary(ctx context.Context, year int16) (*models.GupDashboardSummary, error) {
	summary := &models.GupDashboardSummary{
		MonthlyRealisasi:   make([]models.MonthlyChart, 0),
		CompositionUP:      make([]models.Composition, 0),
		RecentTransactions: make([]models.GUPTransaction, 0),
	}

	// 1. Total Pagu
	err := r.d.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM budgets WHERE year = $1", year).Scan(&summary.TotalPaguAnggaran)
	if err != nil {
		return nil, err
	}

	// 2. Total Realisasi GUP (Full year)
	rowsLaporan, err := r.GetLaporanRows(ctx, year, 0)
	if err != nil {
		return nil, err
	}
	var totalRealisasi float64
	for _, row := range rowsLaporan {
		totalRealisasi += row.Realisasi
	}
	summary.TotalRealisasiGUP = totalRealisasi
	summary.SisaSaldoUP = summary.TotalPaguAnggaran - summary.TotalRealisasiGUP

	// 3. Total GUP Bulan Ini
	currentMonth := time.Now().Month()
	currentYear := time.Now().Year()
	if int(year) != currentYear {
		summary.TotalGupBulanIni = 0
	} else {
		rowsLaporanBulan, err := r.GetLaporanRows(ctx, year, int16(currentMonth))
		if err != nil {
			return nil, err
		}
		var totalRealisasiBulan float64
		for _, row := range rowsLaporanBulan {
			totalRealisasiBulan += row.Realisasi
		}
		summary.TotalGupBulanIni = totalRealisasiBulan
	}

	// 4. Status Dalkot Pending
	err = r.d.QueryRowContext(ctx, "SELECT COUNT(*) FROM dalkot_records WHERE status IN ('Draft', 'Submitted')").Scan(&summary.StatusDalkotPending)
	if err != nil {
		return nil, err
	}

	// 5. Monthly Realisasi
	for m := int16(1); m <= 12; m++ {
		rowsM, err := r.GetLaporanRows(ctx, year, m)
		if err == nil {
			var totalM float64
			for _, rowM := range rowsM {
				totalM += rowM.Realisasi
			}
			if totalM > 0 {
				summary.MonthlyRealisasi = append(summary.MonthlyRealisasi, models.MonthlyChart{
					Month: int(m),
					Total: totalM,
				})
			}
		}
	}

	// 6. Komposisi UP
	compMap := make(map[string]float64)
	compCountMap := make(map[string]int)
	for _, row := range rowsLaporan {
		if row.Realisasi > 0 {
			compMap[row.KodeAkun] += row.Realisasi
			compCountMap[row.KodeAkun] += int(row.JumlahTransaksi)
		}
	}
	for code, total := range compMap {
		summary.CompositionUP = append(summary.CompositionUP, models.Composition{
			Label: code,
			Value: total,
			Count: int64(compCountMap[code]),
		})
	}

	// 7. Recent Transactions (limit 5)
	rRows, err := r.d.QueryContext(ctx, `
		SELECT 
			gt.id, gt.business_id, gt.payment_description, gt.procurement_type_id, 
			gt.funding_source_id, gt.value_amount, gt.paid_amount, gt.tax_amount, 
			gt.receipt_date, gt.recipient, gt.pum, gt.created_at, gt.updated_at,
			pt.name as procurement_type_name
		FROM gup_transactions gt
		JOIN procurement_types pt ON gt.procurement_type_id = pt.id
		ORDER BY gt.receipt_date DESC NULLS LAST, gt.created_at DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rRows.Close()
	for rRows.Next() {
		var gt models.GUPTransaction
		if err := rRows.Scan(
			&gt.ID, &gt.BusinessID, &gt.PaymentDescription, &gt.ProcurementTypeID,
			&gt.FundingSourceID, &gt.ValueAmount, &gt.PaidAmount, &gt.TaxAmount,
			&gt.ReceiptDate, &gt.Recipient, &gt.Pum, &gt.CreatedAt, &gt.UpdatedAt,
			&gt.ProcurementTypeName,
		); err == nil {
			summary.RecentTransactions = append(summary.RecentTransactions, gt)
		}
	}

	return summary, nil
}
