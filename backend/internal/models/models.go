package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Base model for UUID support
type Base struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	Base
	Email        string `json:"email"`
	Password     string `json:"-"` // Never return password
	Name         string `json:"name"`
	Role         string `json:"role"` // super_admin, keuangan, kasubag, protokol
	NIP          string `json:"nip"`
	NomorHP      string `json:"nomorHp"`
	Pangkat      string `json:"pangkat"`
	Golongan     string `json:"golongan"`
	Jabatan      string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
	SessionID    string `json:"-"` // Tracks the current active session ID
	DemoPassword string `json:"-"` // Stores plain text password for Demo Banner (INSECURE - DEMO ONLY)
}

type TravelRecord struct {
	Base
	SPDNumber   string      `json:"spd"`
	EmployeeID  uuid.UUID   `json:"employeeId"`
	EmployeeIDs []uuid.UUID `json:"employeeIds,omitempty"` // For bulk creation
	Employee    User        `json:"employee"`              // Linked to User now
	CreatorID   uuid.UUID   `json:"creatorId"`             // User who created this
	Creator     User        `json:"creator"`
	StartDate   time.Time   `json:"startDate"`
	EndDate     time.Time   `json:"endDate"`
	Location    string      `json:"location"`
	Province    string      `json:"province"`
	Locations   []TravelLocation `json:"locations,omitempty"` // Multiple locations support
	Type        string      `json:"type"` // dalam_kota, luar_kota, luar_negeri
	Purpose     string      `json:"purpose"`
	Stakeholder string      `json:"stakeholder"`
	Agenda      string      `json:"agenda"`
	Status      string      `json:"status"` // Draft, Submitted, Approved, Rejected
	IsViewed    bool        `json:"isViewed"`
	ReportStatus  string    `json:"reportStatus"` // Pending, Completed
	PaymentStatus string    `json:"paymentStatus"` // Unpaid, Paid
	TotalCost     float64   `json:"totalCost"`

	// Documents
	SuratTugasPath   string `json:"suratTugasPath"`
	SuratTugasNumber string `json:"suratTugasNumber"`

	// Relationships
	Cost   *TravelCost   `json:"costs,omitempty"`
	Report *TravelReport `json:"reportData,omitempty"`
}

type TravelLocation struct {
	Base
	TravelRecordID uuid.UUID `json:"travelRecordId"`
	Location       string    `json:"location"`
	Province       string    `json:"province"`
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
}

type TravelCost struct {
	TravelRecordID     uuid.UUID `json:"-"`
	TicketGo           float64   `json:"ticketGo"`
	TicketBack         float64   `json:"ticketBack"`
	DailyAllowanceDays int       `json:"dailyAllowanceDays"`
	DailyAllowanceRate float64   `json:"dailyAllowanceRate"`
	HotelDays          int       `json:"hotelDays"`
	HotelRate          float64   `json:"hotelRate"`
	LocalTransport     float64   `json:"localTransport"`
	RegionalTransport  float64   `json:"regionalTransport"`
	TransportMode      string    `json:"transportMode"`
	TransportAmount    float64   `json:"transportAmount"`

	// Additional Costs
	OtherCost     float64 `json:"otherCost"`
	OtherCostDesc string  `json:"otherCostDesc"`

	// Receipts and Documents
	ReceiptFiles     json.RawMessage `json:"receiptFiles"`
	TicketGoFile     json.RawMessage `json:"ticketGoFile"`
	TicketBackFile   json.RawMessage `json:"ticketBackFile"`
	BoardingPassFile json.RawMessage `json:"boardingPassFile"`
	HotelFile        json.RawMessage `json:"hotelFile"`
	TransportFile    json.RawMessage `json:"transportFile"`
	AdditionalCosts  json.RawMessage `json:"additionalCosts"`
}

type TravelReport struct {
	TravelRecordID uuid.UUID      `json:"-"`
	Text           string         `json:"text"`
	SubmittedAt    time.Time      `json:"submittedAt"`
	Files          json.RawMessage `json:"files"` // Storing file metadata/links as JSON
	SppdFile       json.RawMessage `json:"sppdFile"`
	SuratTugasFile json.RawMessage `json:"suratTugasFile"`
}

// ==========================================
// Master Data Models (SBM & Reference)
// ==========================================

type Province struct {
	Base
	Name string `json:"name"`
	Code string `json:"code"` // e.g., "31" for DKI Jakarta
}

type SBMRate struct {
	Base
	ProvinceID uuid.UUID `json:"provinceId"`
	Province   Province  `json:"province"`

	Year int `json:"year"` // e.g., 2025

	// Uang Harian (Per diem)
	FullboardRate   float64 `json:"fullboardRate"`
	FullhalfRate    float64 `json:"fullhalfRate"`
	OutsideCityRate float64 `json:"outsideCityRate"` // Luar Kota Biasa
	InsideCityRate  float64 `json:"insideCityRate"`  // Dalam Kota > 8 Jam
	DiklatRate      float64 `json:"diklatRate"`

	// Batas Tertinggi Penginapan (Hotel)
	HotelEchelon1 float64 `json:"hotelEchelon1"` // Menteri/Eselon I
	HotelEchelon2 float64 `json:"hotelEchelon2"`
	HotelEchelon3 float64 `json:"hotelEchelon3"`
	HotelEchelon4 float64 `json:"hotelEchelon4"` // Gol III
	HotelStaff    float64 `json:"hotelStaff"`    // Gol II/I

	// Transport Taksi (Perjalanan Dinas Dalam Negeri)
	TaxiRate float64 `json:"taxiRate"`
}
