package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AccountCode struct {
	Base
	Code        string `json:"code"`
	Mak         string `json:"mak"`
	Description string `json:"description,omitempty"`
}

type ProcurementType struct {
	Base
	AccountCodeID uuid.UUID `json:"accountCodeId"`
	Name          string    `json:"name"`
	IsActive      bool      `json:"isActive"`
	// Joined fields
	AccountCode string `json:"accountCode,omitempty"`
	AccountMak  string `json:"accountMak,omitempty"`
}

type FundingSource struct {
	Base
	Year            int16   `json:"year"`
	MonthNumber     int16   `json:"monthNumber"`
	MonthName       string  `json:"monthName"`
	GupLabel        string  `json:"gupLabel"`
	RemainingBudget float64 `json:"remainingBudget"`
}

type Budget struct {
	Base
	Year              int16     `json:"year"`
	ProcurementTypeID uuid.UUID `json:"procurementTypeId"`
	Amount            float64   `json:"amount"`
	// Joined fields
	ProcurementTypeName string `json:"procurementTypeName,omitempty"`
}

type MonthlyLS struct {
	Base
	FundingSourceID uuid.UUID  `json:"fundingSourceId"`
	AccountCodeID   *uuid.UUID `json:"accountCodeId,omitempty"`
	Amount          float64    `json:"amount"`
	// Joined fields
	MonthName string `json:"monthName,omitempty"`
	GupLabel  string `json:"gupLabel,omitempty"`
}

type GUPTransaction struct {
	Base
	BusinessID         string          `json:"businessId"`
	PaymentDescription string          `json:"paymentDescription"`
	ProcurementTypeID  uuid.UUID       `json:"procurementTypeId"`
	FundingSourceID    *uuid.UUID      `json:"fundingSourceId,omitempty"`
	ValueAmount        float64         `json:"valueAmount"`
	PaidAmount         float64         `json:"paidAmount"`
	TaxAmount          float64         `json:"taxAmount"`
	ReceiptDate        *time.Time      `json:"receiptDate,omitempty"`
	Recipient          string          `json:"recipient,omitempty"`
	Pum                string          `json:"pum,omitempty"`
	DocumentFile       json.RawMessage `json:"documentFile,omitempty"`
	
	// Joined fields
	ProcurementTypeName string `json:"procurementTypeName,omitempty"`
	FundingSourceLabel  string `json:"fundingSourceLabel,omitempty"`
}

// LaporanRow adalah hasil agregasi per Jenis Pengadaan untuk halaman Laporan.
// Nilai ini dihitung langsung di query SQL (GROUP BY procurement_type_id).
type LaporanRow struct {
	ProcurementTypeID uuid.UUID `json:"procurementTypeId"`
	JenisPengadaan    string    `json:"jenisPengadaan"`
	KodeAkun          string    `json:"kodeAkun"`
	Mak               string    `json:"mak"`
	JumlahTransaksi   int64     `json:"jumlahTransaksi"`
	Realisasi         float64   `json:"realisasi"`
	NilaiPengajuan    float64   `json:"nilaiPengajuan"`
	TotalPajak        float64   `json:"totalPajak"`
	Anggaran          float64   `json:"anggaran"`
	SisaAnggaran      float64   `json:"sisaAnggaran"`
	PersentaseSerapan float64   `json:"persentaseSerapan"`
}

// LaporanSummary adalah ringkasan total dari semua LaporanRow.
type LaporanSummary struct {
	TotalAnggaran      float64 `json:"totalAnggaran"`
	TotalRealisasi     float64 `json:"totalRealisasi"`
	SisaAnggaran       float64 `json:"sisaAnggaran"`
	TotalTransaksi     int64   `json:"totalTransaksi"`
	TotalJenisPengadaan int    `json:"totalJenisPengadaan"`
	PersentaseSerapan  float64 `json:"persentaseSerapan"`
}

// LaporanResponse adalah response lengkap untuk endpoint GET /gup/laporan.
type LaporanResponse struct {
	Summary LaporanSummary `json:"summary"`
	Rows    []LaporanRow   `json:"rows"`
}

type GupDashboardSummary struct {
	TotalPaguAnggaran   float64          `json:"totalPaguAnggaran"`
	TotalRealisasiGUP   float64          `json:"totalRealisasiGup"`
	SisaSaldoUP         float64          `json:"sisaSaldoUp"`
	TotalGupBulanIni    float64          `json:"totalGupBulanIni"`
	StatusDalkotPending int64            `json:"statusDalkotPending"`
	MonthlyRealisasi    []MonthlyChart   `json:"monthlyRealisasi"`
	CompositionUP       []Composition    `json:"compositionUp"`
	RecentTransactions  []GUPTransaction `json:"recentTransactions"`
}

type MonthlyChart struct {
	Month int     `json:"month"`
	Total float64 `json:"total"`
}

type Composition struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}
