package models

import (
	"time"

	"github.com/google/uuid"
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
	Role     string `gorm:"not null;default:'user'" json:"role"` // super_admin, keuangan, ppk, user
}

type Employee struct {
	Base
	Name     string `gorm:"not null" json:"name"`
	NIP      string `gorm:"uniqueIndex;not null" json:"nip"`
	Rank     string `json:"rank"`     // e.g. "Pembina"
	Golongan string `json:"golongan"` // e.g. "IV/a"
}

type TravelRecord struct {
	Base
	SPDNumber    string    `gorm:"uniqueIndex" json:"spd"`
	EmployeeID   uuid.UUID `gorm:"type:uuid;not null" json:"employeeId"`
	Employee     Employee  `gorm:"foreignKey:EmployeeID" json:"employee"`
	CreatorID    uuid.UUID `gorm:"type:uuid;not null" json:"creatorId"` // User who created this
	Creator      User      `gorm:"foreignKey:CreatorID" json:"creator"`
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	Location     string    `json:"location"`
	Province     string    `json:"province"`
	Purpose      string    `json:"purpose"`
	Stakeholder  string    `json:"stakeholder"`
	Agenda       string    `json:"agenda"`
	Status       string    `gorm:"default:'Submitted'" json:"status"`           // Submitted, Approved, Rejected
	ReportStatus string    `gorm:"default:'Pending'" json:"reportStatus"`       // Pending, Completed
	TotalCost    float64   `json:"totalCost"`

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
}

type TravelReport struct {
	TravelRecordID uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	Text           string    `json:"text"`
	SubmittedAt    time.Time `json:"submittedAt"`
	Files          []byte    `gorm:"type:jsonb" json:"files"` // Storing file metadata/links as JSON
}

// Hooks
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

func (e *Employee) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}

func (t *TravelRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
