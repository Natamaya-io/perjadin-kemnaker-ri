package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DalkotRecord struct {
	Base
	ExecutionDate     time.Time          `json:"executionDate"`
	DocumentationFile json.RawMessage    `json:"documentationFile,omitempty"`
	Assignments       []DalkotAssignment `json:"assignments,omitempty"`
	SPDNumber         string             `json:"spdNumber"`
	Category          string             `json:"category"`
	Official          string             `json:"official"`
	DalkotType        string             `json:"dalkotType"`
	ActivityName      string             `json:"activityName"`
	Location          string             `json:"location"`
	Status            string             `json:"status"`
	ReportContent     string             `json:"reportContent"`
}

type DalkotAssignment struct {
	Base
	User           *User     `json:"user,omitempty"`
	AssignmentType string    `json:"assignmentType"`
	Status         string    `json:"status"`
	SPJCost        float64   `json:"spjCost"`
	ActualCost     float64   `json:"actualCost"`
	SequenceNumber int       `json:"sequenceNumber"`
	DalkotRecordID uuid.UUID `json:"dalkotRecordId"`
	UserID         uuid.UUID `json:"userId"`
}

type DalkotLocation struct {
	ID        int32     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
}

type DalkotRate struct {
	ID           uuid.UUID `json:"id"`
	CategoryName string    `json:"categoryName"`
	RateAmount   float64   `json:"rateAmount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
