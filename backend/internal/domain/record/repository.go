package record

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/sqlc-dev/pqtype"
)

type Repository interface {
	CreateTravelRecord(record *models.TravelRecord) error
	GetTravelRecords(filters map[string]interface{}) ([]models.TravelRecord, error)
	GetTravelRecordByID(id uuid.UUID) (*models.TravelRecord, error)
	GetOverlappingRecords(employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TravelRecord, error)
	UpdateTravelRecord(record *models.TravelRecord) error
	DeleteTravelRecord(id uuid.UUID) error
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

// Helpers
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

func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func fromNullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func toNullFloat(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

func fromNullFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func toNullInt32(i int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func fromNullInt32(ni sql.NullInt32) int {
	if ni.Valid {
		return int(ni.Int32)
	}
	return 0
}

func toJsonb(d json.RawMessage) pqtype.NullRawMessage {
	if len(d) == 0 {
		return pqtype.NullRawMessage{RawMessage: json.RawMessage("{}"), Valid: true}
	}
	return pqtype.NullRawMessage{RawMessage: json.RawMessage(d), Valid: true}
}

func fromJsonb(j pqtype.NullRawMessage) json.RawMessage {
	if !j.Valid || len(j.RawMessage) == 0 {
		return nil
	}
	return json.RawMessage(j.RawMessage)
}

func mapDBUser(dbu db.User) models.User {
	return models.User{
		Base: models.Base{
			ID:        dbu.ID,
			CreatedAt: fromNullTime(dbu.CreatedAt),
			UpdatedAt: fromNullTime(dbu.UpdatedAt),
		},
		Email:        dbu.Email,
		Password:     dbu.Password,
		Name:         dbu.Name,
		Role:         dbu.Role,
		NIP:          fromNullString(dbu.Nip),
		NomorHP:      fromNullString(dbu.NomorHp),
		Pangkat:      fromNullString(dbu.Pangkat),
		Golongan:     fromNullString(dbu.Golongan),
		Jabatan:      fromNullString(dbu.Jabatan),
		TingkatBiaya: fromNullString(dbu.TingkatBiaya),
		SessionID:    fromNullString(dbu.SessionID),
		DemoPassword: fromNullString(dbu.DemoPassword),
	}
}

func mapDBRecord(dbr db.TravelRecord) models.TravelRecord {
	return models.TravelRecord{
		Base: models.Base{
			ID:        dbr.ID,
			CreatedAt: fromNullTime(dbr.CreatedAt),
			UpdatedAt: fromNullTime(dbr.UpdatedAt),
		},
		SPDNumber:        fromNullString(dbr.SpdNumber),
		EmployeeID:       dbr.EmployeeID,
		CreatorID:        dbr.CreatorID,
		StartDate:        fromNullTime(dbr.StartDate),
		EndDate:          fromNullTime(dbr.EndDate),
		Location:         fromNullString(dbr.Location),
		Province:         fromNullString(dbr.Province),
		Type:             fromNullString(dbr.Type),
		Purpose:          fromNullString(dbr.Purpose),
		Stakeholder:      fromNullString(dbr.Stakeholder),
		Agenda:           fromNullString(dbr.Agenda),
		Status:           fromNullString(dbr.Status),
		IsViewed:         dbr.IsViewed.Bool,
		ReportStatus:     fromNullString(dbr.ReportStatus),
		PaymentStatus:    fromNullString(dbr.PaymentStatus),
		TotalCost:        fromNullFloat(dbr.TotalCost),
		SuratTugasPath:   fromNullString(dbr.SuratTugasPath),
		SuratTugasNumber: fromNullString(dbr.SuratTugasNumber),
	}
}

func mapDBLocation(dbl db.TravelLocation) models.TravelLocation {
	return models.TravelLocation{
		Base: models.Base{
			ID:        dbl.ID,
			CreatedAt: dbl.CreatedAt.Time,
			UpdatedAt: dbl.UpdatedAt.Time,
		},
		TravelRecordID: dbl.TravelRecordID,
		Location:       dbl.Location,
		Province:       dbl.Province,
		StartDate:      dbl.StartDate,
		EndDate:        dbl.EndDate,
	}
}

func mapDBCost(dbc db.TravelCost) models.TravelCost {
	return models.TravelCost{
		TravelRecordID:     dbc.TravelRecordID,
		TicketGo:           fromNullFloat(dbc.TicketGo),
		TicketBack:         fromNullFloat(dbc.TicketBack),
		DailyAllowanceDays: fromNullInt32(dbc.DailyAllowanceDays),
		DailyAllowanceRate: fromNullFloat(dbc.DailyAllowanceRate),
		HotelDays:          fromNullInt32(dbc.HotelDays),
		HotelRate:          fromNullFloat(dbc.HotelRate),
		LocalTransport:     fromNullFloat(dbc.LocalTransport),
		RegionalTransport:  fromNullFloat(dbc.RegionalTransport),
		TransportMode:      fromNullString(dbc.TransportMode),
		TransportAmount:    fromNullFloat(dbc.TransportAmount),
		OtherCost:          fromNullFloat(dbc.OtherCost),
		OtherCostDesc:      fromNullString(dbc.OtherCostDesc),
		ReceiptFiles:       fromJsonb(dbc.ReceiptFiles),
		TicketGoFile:       fromJsonb(dbc.TicketGoFile),
		TicketBackFile:     fromJsonb(dbc.TicketBackFile),
		BoardingPassFile:   fromJsonb(dbc.BoardingPassFile),
		HotelFile:          fromJsonb(dbc.HotelFile),
		TransportFile:      fromJsonb(dbc.TransportFile),
		AdditionalCosts:    fromJsonb(dbc.AdditionalCosts),
	}
}

func mapDBReport(dbrep db.TravelReport) models.TravelReport {
	return models.TravelReport{
		TravelRecordID: dbrep.TravelRecordID,
		Text:           fromNullString(dbrep.Text),
		SubmittedAt:    fromNullTime(dbrep.SubmittedAt),
		Files:          fromJsonb(dbrep.Files),
		SppdFile:       fromJsonb(dbrep.SppdFile),
		SuratTugasFile: fromJsonb(dbrep.SuratTugasFile),
	}
}

func (r *repository) getUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	dbu, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u := mapDBUser(dbu)
	return &u, nil
}

func (r *repository) CreateTravelRecord(record *models.TravelRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	ctx := context.Background()
	dbr, err := r.q.CreateTravelRecord(ctx, db.CreateTravelRecordParams{
		ID:               record.ID,
		SpdNumber:        toNullString(record.SPDNumber),
		EmployeeID:       record.EmployeeID,
		CreatorID:        record.CreatorID,
		StartDate:        toNullTime(record.StartDate),
		EndDate:          toNullTime(record.EndDate),
		Location:         toNullString(record.Location),
		Province:         toNullString(record.Province),
		Type:             toNullString(record.Type),
		Purpose:          toNullString(record.Purpose),
		Stakeholder:      toNullString(record.Stakeholder),
		Agenda:           toNullString(record.Agenda),
		Status:           toNullString(record.Status),
		IsViewed:         sql.NullBool{Bool: record.IsViewed, Valid: true},
		ReportStatus:     toNullString(record.ReportStatus),
		PaymentStatus:    toNullString(record.PaymentStatus),
		TotalCost:        toNullFloat(record.TotalCost),
		SuratTugasPath:   toNullString(record.SuratTugasPath),
		SuratTugasNumber: toNullString(record.SuratTugasNumber),
	})
	if err != nil {
		return err
	}
	
	// Preserve locations before mapping
	locations := record.Locations
	*record = mapDBRecord(dbr)
	record.Locations = locations

	// Create locations
	for _, loc := range record.Locations {
		// Always generate a new ID for the join table record to avoid conflicts
		locID := uuid.New()
		_, err := r.q.CreateTravelLocation(ctx, db.CreateTravelLocationParams{
			ID:             locID,
			TravelRecordID: record.ID,
			Location:       loc.Location,
			Province:       loc.Province,
			StartDate:      loc.StartDate,
			EndDate:        loc.EndDate,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) GetTravelRecords(filters map[string]interface{}) ([]models.TravelRecord, error) {
	ctx := context.Background()
	var statusFilter string
	if status, ok := filters["status"]; ok && status != "" {
		statusFilter = status.(string)
	}

	dbrs, err := r.q.GetTravelRecords(ctx, statusFilter)
	if err != nil {
		return nil, err
	}

	records := make([]models.TravelRecord, len(dbrs))
	var wg sync.WaitGroup

	for i, dbr := range dbrs {
		wg.Add(1)
		go func(index int, d db.TravelRecord) {
			defer wg.Done()
			rec := mapDBRecord(d)

			var relWg sync.WaitGroup

			relWg.Add(1)
			go func() {
				defer relWg.Done()
				emp, _ := r.getUserByID(ctx, rec.EmployeeID)
				if emp != nil {
					rec.Employee = *emp
				}
			}()

			relWg.Add(1)
			go func() {
				defer relWg.Done()
				creator, _ := r.getUserByID(ctx, rec.CreatorID)
				if creator != nil {
					rec.Creator = *creator
				}
			}()

			relWg.Add(1)
			go func() {
				defer relWg.Done()
				dbc, err := r.q.GetTravelCostByRecordID(ctx, rec.ID)
				if err == nil {
					cost := mapDBCost(dbc)
					rec.Cost = &cost
				}
			}()

			relWg.Add(1)
			go func() {
				defer relWg.Done()
				dbrep, err := r.q.GetTravelReportByRecordID(ctx, rec.ID)
				if err == nil {
					rep := mapDBReport(dbrep)
					rec.Report = &rep
				}
			}()

			relWg.Add(1)
			go func() {
				defer relWg.Done()
				dbls, err := r.q.GetTravelLocationsByRecordID(ctx, rec.ID)
				if err == nil {
					locs := make([]models.TravelLocation, len(dbls))
					for j, dbl := range dbls {
						locs[j] = mapDBLocation(dbl)
					}
					rec.Locations = locs
				}
			}()

			relWg.Wait()
			records[index] = rec
		}(i, dbr)
	}
	wg.Wait()

	return records, nil
}

func (r *repository) GetTravelRecordByID(id uuid.UUID) (*models.TravelRecord, error) {
	ctx := context.Background()
	dbr, err := r.q.GetTravelRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rec := mapDBRecord(dbr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		emp, _ := r.getUserByID(ctx, rec.EmployeeID)
		if emp != nil {
			rec.Employee = *emp
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		creator, _ := r.getUserByID(ctx, rec.CreatorID)
		if creator != nil {
			rec.Creator = *creator
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbc, err := r.q.GetTravelCostByRecordID(ctx, rec.ID)
		if err == nil {
			cost := mapDBCost(dbc)
			rec.Cost = &cost
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbrep, err := r.q.GetTravelReportByRecordID(ctx, rec.ID)
		if err == nil {
			rep := mapDBReport(dbrep)
			rec.Report = &rep
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbls, err := r.q.GetTravelLocationsByRecordID(ctx, rec.ID)
		if err == nil {
			locs := make([]models.TravelLocation, len(dbls))
			for j, dbl := range dbls {
				locs[j] = mapDBLocation(dbl)
			}
			rec.Locations = locs
		}
	}()

	wg.Wait()

	return &rec, nil
}

func (r *repository) GetOverlappingRecords(employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TravelRecord, error) {
	dbrs, err := r.q.GetOverlappingRecords(context.Background(), db.GetOverlappingRecordsParams{
		EmployeeID: employeeID,
		StartDate:  toNullTime(endDate),
		EndDate:    toNullTime(startDate),
	})
	if err != nil {
		return nil, err
	}

	var records []models.TravelRecord
	for _, dbr := range dbrs {
		records = append(records, mapDBRecord(dbr))
	}
	return records, nil
}

func (r *repository) UpdateTravelRecord(record *models.TravelRecord) error {
	ctx := context.Background()
	_, err := r.q.UpdateTravelRecord(ctx, db.UpdateTravelRecordParams{
		ID:               record.ID,
		SpdNumber:        toNullString(record.SPDNumber),
		EmployeeID:       record.EmployeeID,
		CreatorID:        record.CreatorID,
		StartDate:        toNullTime(record.StartDate),
		EndDate:          toNullTime(record.EndDate),
		Location:         toNullString(record.Location),
		Province:         toNullString(record.Province),
		Type:             toNullString(record.Type),
		Purpose:          toNullString(record.Purpose),
		Stakeholder:      toNullString(record.Stakeholder),
		Agenda:           toNullString(record.Agenda),
		Status:           toNullString(record.Status),
		IsViewed:         sql.NullBool{Bool: record.IsViewed, Valid: true},
		ReportStatus:     toNullString(record.ReportStatus),
		PaymentStatus:    toNullString(record.PaymentStatus),
		TotalCost:        toNullFloat(record.TotalCost),
		SuratTugasPath:   toNullString(record.SuratTugasPath),
		SuratTugasNumber: toNullString(record.SuratTugasNumber),
	})
	if err != nil {
		return err
	}

	// Update locations (simpler to delete and recreate)
	r.q.DeleteTravelLocationsByRecordID(ctx, record.ID)
	for _, loc := range record.Locations {
		if loc.ID == uuid.Nil {
			loc.ID = uuid.New()
		}
		r.q.CreateTravelLocation(ctx, db.CreateTravelLocationParams{
			ID:             loc.ID,
			TravelRecordID: record.ID,
			Location:       loc.Location,
			Province:       loc.Province,
			StartDate:      loc.StartDate,
			EndDate:        loc.EndDate,
		})
	}


	if record.Cost != nil {
		_, err := r.q.UpdateTravelCost(ctx, db.UpdateTravelCostParams{
			TravelRecordID:     record.ID,
			TicketGo:           toNullFloat(record.Cost.TicketGo),
			TicketBack:         toNullFloat(record.Cost.TicketBack),
			DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
			DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
			HotelDays:          toNullInt32(record.Cost.HotelDays),
			HotelRate:          toNullFloat(record.Cost.HotelRate),
			LocalTransport:     toNullFloat(record.Cost.LocalTransport),
			RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
			TransportMode:      toNullString(record.Cost.TransportMode),
			TransportAmount:    toNullFloat(record.Cost.TransportAmount),
			OtherCost:          toNullFloat(record.Cost.OtherCost),
			OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
			ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
			TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
			TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
			BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
			HotelFile:          toJsonb(record.Cost.HotelFile),
			TransportFile:      toJsonb(record.Cost.TransportFile),
			AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
		})
		if err != nil {
			r.q.CreateTravelCost(ctx, db.CreateTravelCostParams{
				TravelRecordID:     record.ID,
				TicketGo:           toNullFloat(record.Cost.TicketGo),
				TicketBack:         toNullFloat(record.Cost.TicketBack),
				DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
				DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
				HotelDays:          toNullInt32(record.Cost.HotelDays),
				HotelRate:          toNullFloat(record.Cost.HotelRate),
				LocalTransport:     toNullFloat(record.Cost.LocalTransport),
				RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
				TransportMode:      toNullString(record.Cost.TransportMode),
				TransportAmount:    toNullFloat(record.Cost.TransportAmount),
				OtherCost:          toNullFloat(record.Cost.OtherCost),
				OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
				ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
				TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
				TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
				BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
				HotelFile:          toJsonb(record.Cost.HotelFile),
				TransportFile:      toJsonb(record.Cost.TransportFile),
				AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
			})
		}
	}

	if record.Report != nil {
		_, err := r.q.UpdateTravelReport(ctx, db.UpdateTravelReportParams{
			TravelRecordID: record.ID,
			Text:           toNullString(record.Report.Text),
			SubmittedAt:    toNullTime(record.Report.SubmittedAt),
			Files:          toJsonb(record.Report.Files),
			SppdFile:       toJsonb(record.Report.SppdFile),
			SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
		})
		if err != nil {
			r.q.CreateTravelReport(ctx, db.CreateTravelReportParams{
				TravelRecordID: record.ID,
				Text:           toNullString(record.Report.Text),
				SubmittedAt:    toNullTime(record.Report.SubmittedAt),
				Files:          toJsonb(record.Report.Files),
				SppdFile:       toJsonb(record.Report.SppdFile),
				SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
			})
		}
	}

	return nil
}

func (r *repository) DeleteTravelRecord(id uuid.UUID) error {
	return r.q.DeleteTravelRecord(context.Background(), id)
}
