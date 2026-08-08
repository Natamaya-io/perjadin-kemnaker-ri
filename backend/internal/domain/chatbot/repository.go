package chatbot

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetGupSnapshot(ctx context.Context, year int) (GupSnapshot, error)
	GetDalkotSnapshot(ctx context.Context, year int) (DalkotSnapshot, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type GupSnapshot struct {
	TotalTransaksi int
	TotalPagu      int64
	TotalRealisasi int64
	TotalSisa      int64
	Menus          []GupMenuStat
}

type GupMenuStat struct {
	Name           string
	TotalTransaksi int
	Pagu           int64
	Realisasi      int64
	Sisa           int64
}

type DalkotSnapshot struct {
	JumlahPengajuan int
	TotalSPJ        int64
	TotalRiil       int64
	TotalBiaya      int64
	Draft           int
	Pending         int
	Selesai         int
	PetugasSPJ      []PetugasStat
	PetugasRiil     []PetugasStat
}

type PetugasStat struct {
	Nama        string
	JumlahTugas int
	TotalBiaya  int64
	Pending     int
	Selesai     int
}

func (r *repository) GetGupSnapshot(ctx context.Context, year int) (GupSnapshot, error) {
	// A simple query to get GUP stats per menu
	query := `
		SELECT 
			pt.name,
			COUNT(gt.id) as total_transaksi,
			COALESCE(MAX(b.amount), 0) as pagu,
			COALESCE(SUM(gt.paid_amount), 0) as realisasi
		FROM procurement_types pt
		LEFT JOIN budgets b ON b.procurement_type_id = pt.id AND b.year = $1
		LEFT JOIN gup_transactions gt ON gt.procurement_type_id = pt.id 
		GROUP BY pt.name
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return GupSnapshot{}, err
	}
	defer rows.Close()

	var snap GupSnapshot
	var totalPagu, totalRealisasi int64

	for rows.Next() {
		var m GupMenuStat
		var fPagu, fRealisasi float64
		if err := rows.Scan(&m.Name, &m.TotalTransaksi, &fPagu, &fRealisasi); err != nil {
			return GupSnapshot{}, err
		}
		m.Pagu = int64(fPagu)
		m.Realisasi = int64(fRealisasi)
		m.Sisa = m.Pagu - m.Realisasi
		snap.Menus = append(snap.Menus, m)
		snap.TotalTransaksi += m.TotalTransaksi
		totalPagu += m.Pagu
		totalRealisasi += m.Realisasi
	}

	snap.TotalPagu = totalPagu
	snap.TotalRealisasi = totalRealisasi
	snap.TotalSisa = totalPagu - totalRealisasi

	return snap, nil
}

func (r *repository) GetDalkotSnapshot(ctx context.Context, year int) (DalkotSnapshot, error) {
	// Dalkot overall stats
	queryStats := `
		SELECT 
			COUNT(id) as jumlah_pengajuan,
			COALESCE(SUM(total_spj_cost), 0) as total_spj,
			COALESCE(SUM(total_actual_cost), 0) as total_riil,
			COUNT(CASE WHEN status = 'Draft' THEN 1 END) as draft_count,
			COUNT(CASE WHEN status = 'Pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'Selesai' THEN 1 END) as selesai_count
		FROM dalkot_records
		WHERE EXTRACT(YEAR FROM execution_date) = $1
	`
	var snap DalkotSnapshot
	var fTotalSPJ, fTotalRiil float64
	err := r.db.QueryRowContext(ctx, queryStats, year).Scan(
		&snap.JumlahPengajuan, &fTotalSPJ, &fTotalRiil,
		&snap.Draft, &snap.Pending, &snap.Selesai,
	)
	if err != nil {
		return DalkotSnapshot{}, err
	}
	snap.TotalSPJ = int64(fTotalSPJ)
	snap.TotalRiil = int64(fTotalRiil)
	snap.TotalBiaya = snap.TotalSPJ + snap.TotalRiil

	// Petugas stats (SPJ vs Riil is based on dalkot_assignments.assignment_type)
	// We'll join with users to get names.
	queryPetugas := `
		SELECT 
			u.name,
			da.assignment_type,
			COUNT(da.id) as jumlah_tugas,
			COALESCE(SUM(da.spj_cost + da.actual_cost), 0) as total_biaya,
			COUNT(CASE WHEN da.status = 'Pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN da.status = 'Selesai' THEN 1 END) as selesai_count
		FROM dalkot_assignments da
		JOIN users u ON u.id = da.user_id
		JOIN dalkot_records dr ON dr.id = da.dalkot_record_id
		WHERE EXTRACT(YEAR FROM dr.execution_date) = $1
		GROUP BY u.name, da.assignment_type
	`
	rows, err := r.db.QueryContext(ctx, queryPetugas, year)
	if err != nil {
		return DalkotSnapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var p PetugasStat
		var aType string
		var fTotalBiaya float64
		if err := rows.Scan(&p.Nama, &aType, &p.JumlahTugas, &fTotalBiaya, &p.Pending, &p.Selesai); err != nil {
			return DalkotSnapshot{}, err
		}
		p.TotalBiaya = int64(fTotalBiaya)
		if aType == "SPJ RIIL" || aType == "SPJ" {
			snap.PetugasSPJ = append(snap.PetugasSPJ, p)
		} else {
			snap.PetugasRiil = append(snap.PetugasRiil, p)
		}
	}

	return snap, nil
}
