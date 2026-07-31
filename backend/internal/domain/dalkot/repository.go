package dalkot

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/sqlc-dev/pqtype"
)

type Repository interface {
	CreateRecord(record *models.DalkotRecord) error
	UpdateRecord(record *models.DalkotRecord) error
	DeleteRecord(id uuid.UUID) error
	GetRecordByID(id uuid.UUID) (*models.DalkotRecord, error)
	GetRecords() ([]models.DalkotRecord, error)

	CreateAssignment(assignment *models.DalkotAssignment) error
	UpdateAssignment(assignment *models.DalkotAssignment) error
	DeleteAssignment(id uuid.UUID) error
	GetAssignmentsByRecordID(recordID uuid.UUID) ([]models.DalkotAssignment, error)

	GetLocations() ([]models.DalkotLocation, error)
}

type repository struct {
	q *db.Queries
	d *sql.DB
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		q: db.New(sqlDB),
		d: sqlDB,
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromNullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func toNullRawMessage(raw json.RawMessage) pqtype.NullRawMessage {
	if len(raw) == 0 {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{RawMessage: raw, Valid: true}
}

func mapDBRecord(dbr db.DalkotRecord) models.DalkotRecord {
	var docFile json.RawMessage
	if dbr.DocumentationFile.Valid {
		docFile = json.RawMessage(dbr.DocumentationFile.RawMessage)
	}

	return models.DalkotRecord{
		Base: models.Base{
			ID:        dbr.ID,
			CreatedAt: fromNullTime(dbr.CreatedAt),
			UpdatedAt: fromNullTime(dbr.UpdatedAt),
		},
		SPDNumber:         fromNullString(dbr.SpdNumber),
		ExecutionDate:     dbr.ExecutionDate,
		Category:          dbr.Category,
		Official:          dbr.Official,
		DalkotType:        dbr.DalkotType,
		ActivityName:      dbr.ActivityName,
		Location:          dbr.Location,
		Status:            fromNullString(dbr.Status),
		DocumentationFile: docFile,
	}
}

func mapDBAssignment(dba db.DalkotAssignment) models.DalkotAssignment {
	return models.DalkotAssignment{
		Base: models.Base{
			ID:        dba.ID,
			CreatedAt: fromNullTime(dba.CreatedAt),
			UpdatedAt: fromNullTime(dba.UpdatedAt),
		},
		DalkotRecordID: dba.DalkotRecordID,
		UserID:         dba.UserID,
		AssignmentType: dba.AssignmentType,
		SPJCost:        dba.SpjCost.Float64,
		ActualCost:     dba.ActualCost.Float64,
		Status:         fromNullString(dba.Status),
	}
}

func mapDBLocation(dbl db.DalkotLocation) models.DalkotLocation {
	return models.DalkotLocation{
		ID:        dbl.ID,
		CreatedAt: fromNullTime(dbl.CreatedAt),
		Name:      dbl.Name,
		IsActive:  dbl.IsActive.Bool,
	}
}

func (r *repository) CreateRecord(record *models.DalkotRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	dbr, err := r.q.CreateDalkotRecord(context.Background(), db.CreateDalkotRecordParams{
		ID:                record.ID,
		SpdNumber:         toNullString(record.SPDNumber),
		ExecutionDate:     record.ExecutionDate,
		Category:          record.Category,
		Official:          record.Official,
		DalkotType:        record.DalkotType,
		ActivityName:      record.ActivityName,
		Location:          record.Location,
		SuratTugasNumber:  sql.NullString{},
		SuratTugasDate:    sql.NullTime{},
		ReportContent:     sql.NullString{},
		Status:            toNullString(record.Status),
		DocumentationFile: toNullRawMessage(record.DocumentationFile),
	})
	if err != nil {
		return err
	}
	*record = mapDBRecord(dbr)
	return nil
}

func (r *repository) UpdateRecord(record *models.DalkotRecord) error {
	dbr, err := r.q.UpdateDalkotRecord(context.Background(), db.UpdateDalkotRecordParams{
		ID:                record.ID,
		SpdNumber:         toNullString(record.SPDNumber),
		ExecutionDate:     record.ExecutionDate,
		Category:          record.Category,
		Official:          record.Official,
		DalkotType:        record.DalkotType,
		ActivityName:      record.ActivityName,
		Location:          record.Location,
		SuratTugasNumber:  sql.NullString{},
		SuratTugasDate:    sql.NullTime{},
		ReportContent:     sql.NullString{},
		Status:            toNullString(record.Status),
		DocumentationFile: toNullRawMessage(record.DocumentationFile),
	})
	if err != nil {
		return err
	}
	*record = mapDBRecord(dbr)
	return nil
}

func (r *repository) DeleteRecord(id uuid.UUID) error {
	return r.q.DeleteDalkotRecord(context.Background(), id)
}

func (r *repository) GetRecordByID(id uuid.UUID) (*models.DalkotRecord, error) {
	dbr, err := r.q.GetDalkotRecordByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	rec := mapDBRecord(dbr)
	return &rec, nil
}

func (r *repository) GetRecords() ([]models.DalkotRecord, error) {
	dbrs, err := r.q.GetDalkotRecords(context.Background())
	if err != nil {
		return nil, err
	}
	records := make([]models.DalkotRecord, len(dbrs))
	for i, dbr := range dbrs {
		records[i] = mapDBRecord(dbr)
	}
	return records, nil
}

func (r *repository) CreateAssignment(assignment *models.DalkotAssignment) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	dba, err := r.q.CreateDalkotAssignment(context.Background(), db.CreateDalkotAssignmentParams{
		ID:             assignment.ID,
		DalkotRecordID: assignment.DalkotRecordID,
		UserID:         assignment.UserID,
		AssignmentType: assignment.AssignmentType,
		SpjCost:        sql.NullFloat64{Float64: assignment.SPJCost, Valid: true},
		ActualCost:     sql.NullFloat64{Float64: assignment.ActualCost, Valid: true},
		Status:         toNullString(assignment.Status),
	})
	if err != nil {
		return err
	}
	*assignment = mapDBAssignment(dba)
	return nil
}

func (r *repository) UpdateAssignment(assignment *models.DalkotAssignment) error {
	dba, err := r.q.UpdateDalkotAssignment(context.Background(), db.UpdateDalkotAssignmentParams{
		ID:             assignment.ID,
		AssignmentType: assignment.AssignmentType,
		SpjCost:        sql.NullFloat64{Float64: assignment.SPJCost, Valid: true},
		ActualCost:     sql.NullFloat64{Float64: assignment.ActualCost, Valid: true},
		Status:         toNullString(assignment.Status),
	})
	if err != nil {
		return err
	}
	*assignment = mapDBAssignment(dba)
	return nil
}

func (r *repository) DeleteAssignment(id uuid.UUID) error {
	return r.q.DeleteDalkotAssignment(context.Background(), id)
}

func (r *repository) GetAssignmentsByRecordID(recordID uuid.UUID) ([]models.DalkotAssignment, error) {
	dbas, err := r.q.GetDalkotAssignmentsByRecordID(context.Background(), recordID)
	if err != nil {
		return nil, err
	}
	assignments := make([]models.DalkotAssignment, len(dbas))
	for i, dba := range dbas {
		assignments[i] = mapDBAssignment(dba)
	}
	return assignments, nil
}

func (r *repository) GetLocations() ([]models.DalkotLocation, error) {
	dbls, err := r.q.GetDalkotLocations(context.Background())
	if err != nil {
		return nil, err
	}
	locations := make([]models.DalkotLocation, len(dbls))
	for i, dbl := range dbls {
		locations[i] = mapDBLocation(dbl)
	}
	return locations, nil
}
