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
	Year        int16  `json:"year"`
	MonthNumber int16  `json:"monthNumber"`
	MonthName   string `json:"monthName"`
	GupLabel    string `json:"gupLabel"`
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
	FundingSourceID uuid.UUID `json:"fundingSourceId"`
	Amount          float64   `json:"amount"`
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
