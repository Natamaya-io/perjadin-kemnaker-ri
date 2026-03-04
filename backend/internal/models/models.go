package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Base model for UUID support
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	Base
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"` // Never return password
	Name     string `gorm:"not null" json:"name"`
	Role     string `gorm:"not null;default:'protokol'" json:"role"` // super_admin, keuangan, kasubag, protokol
	NIP      string `json:"nip"`
	NomorHP  string `json:"nomorHp"`
	Pangkat  string `json:"pangkat"`
	Golongan string `json:"golongan"`
	Jabatan  string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
	SessionID string `json:"-"` // Tracks the current active session ID
	DemoPassword string `json:"-"` // Stores plain text password for Demo Banner (INSECURE - DEMO ONLY)
}

// Employee struct removed as it is replaced by User (Protokol role)

type TravelRecord struct {
	Base
	SPDNumber    string    `gorm:"index" json:"spd"`
	EmployeeID   uuid.UUID `gorm:"type:uuid;not null" json:"employeeId"`
	EmployeeIDs  []uuid.UUID `gorm:"-" json:"employeeIds,omitempty"` // For bulk creation
	Employee     User      `gorm:"foreignKey:EmployeeID" json:"employee"` // Linked to User now
	CreatorID    uuid.UUID `gorm:"type:uuid;not null" json:"creatorId"` // User who created this
	Creator      User      `gorm:"foreignKey:CreatorID" json:"creator"`
	StartDate    time.Time `json:"startDate"`
	// ... rest identical
	EndDate      time.Time `json:"endDate"`
	Location     string    `json:"location"`
	Province     string    `json:"province"`
	Type         string    `json:"type"` // dalam_kota, luar_kota, luar_negeri
	Purpose      string    `json:"purpose"`
	Stakeholder  string    `json:"stakeholder"`
	Agenda       string    `json:"agenda"`
	Status       string    `gorm:"default:'Draft'" json:"status"`           // Draft, Submitted, Approved, Rejected
	ReportStatus string    `gorm:"default:'Pending'" json:"reportStatus"`       // Pending, Completed
	PaymentStatus string   `gorm:"default:'Unpaid'" json:"paymentStatus"`   // Unpaid, Paid
	TotalCost    float64   `json:"totalCost"`

	// Documents
	SuratTugasPath string `json:"suratTugasPath"`

	// Relationships
	Cost   *TravelCost   `gorm:"foreignKey:TravelRecordID;constraint:OnDelete:CASCADE" json:"costs,omitempty"`
	Report *TravelReport `gorm:"foreignKey:TravelRecordID;constraint:OnDelete:CASCADE" json:"reportData,omitempty"`
}

type TravelCost struct {
	TravelRecordID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	TicketGo           float64   `json:"ticketGo"`
	TicketBack         float64   `json:"ticketBack"`
	DailyAllowanceDays int       `json:"dailyAllowanceDays"`
	DailyAllowanceRate float64   `json:"dailyAllowanceRate"`
	HotelDays          int       `json:"hotelDays"`
	HotelRate          float64   `json:"hotelRate"`
	LocalTransport     float64   `json:"localTransport"`
	RegionalTransport  float64   `json:"regionalTransport"`
	TransportMode      string    `json:"transportMode"`
	
	// Additional Costs
	OtherCost     float64 `json:"otherCost"`
	OtherCostDesc string  `json:"otherCostDesc"`

	// Receipts and Documents
	ReceiptFiles  datatypes.JSON `gorm:"type:jsonb" json:"receiptFiles"`
}

type TravelReport struct {
	TravelRecordID uuid.UUID      `gorm:"type:uuid;primaryKey" json:"-"`
	Text           string         `json:"text"`
	SubmittedAt    time.Time      `json:"submittedAt"`
	Files          datatypes.JSON `gorm:"type:jsonb" json:"files"` // Storing file metadata/links as JSON
	SppdFile       datatypes.JSON `gorm:"type:jsonb" json:"sppdFile"`
	SuratTugasFile datatypes.JSON `gorm:"type:jsonb" json:"suratTugasFile"`
}

// ==========================================
// Master Data Models (SBM & Reference)
// ==========================================

type Province struct {
	Base
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	Code string `gorm:"uniqueIndex" json:"code"` // e.g., "31" for DKI Jakarta
}

type SBMRate struct {
	Base
	ProvinceID uuid.UUID `gorm:"type:uuid;not null;index" json:"provinceId"`
	Province   Province  `gorm:"foreignKey:ProvinceID" json:"province"`
	
	Year int `gorm:"index;not null" json:"year"` // e.g., 2025

	// Uang Harian (Per diem)
	FullboardRate    float64 `json:"fullboardRate"`
	FullhalfRate     float64 `json:"fullhalfRate"`
	OutsideCityRate  float64 `json:"outsideCityRate"` // Luar Kota Biasa
	InsideCityRate   float64 `json:"insideCityRate"`  // Dalam Kota > 8 Jam
	DiklatRate       float64 `json:"diklatRate"`
	
	// Batas Tertinggi Penginapan (Hotel)
	HotelEchelon1 float64 `json:"hotelEchelon1"` // Menteri/Eselon I
	HotelEchelon2 float64 `json:"hotelEchelon2"`
	HotelEchelon3 float64 `json:"hotelEchelon3"`
	HotelEchelon4 float64 `json:"hotelEchelon4"` // Gol III
	HotelStaff    float64 `json:"hotelStaff"`    // Gol II/I
	
	// Transport Taksi (Perjalanan Dinas Dalam Negeri)
	TaxiRate float64 `json:"taxiRate"`
}

// Hooks
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

func (t *TravelRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (p *Province) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

func (s *SBMRate) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
