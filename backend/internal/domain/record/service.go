package record

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GenerateSpdNumber() (string, error)
	CreateRecord(record *models.TravelRecord) error
	CreateRecordDirect(record *models.TravelRecord) error // For import: skips overlap check & WA notification
	InvalidateAllCache()                                  // Clear all record caches
	GetRecords(filters map[string]interface{}) ([]models.TravelRecord, error)
	GetRecordByID(id uuid.UUID) (*models.TravelRecord, error)
	UpdateRecord(record *models.TravelRecord) error
	DeleteRecord(id uuid.UUID) error
	DeleteRecordsBySpd(ctx context.Context, spd string) error
}

type service struct {
	repo        Repository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewService(repo Repository, cfg *config.Config, rdb *redis.Client) Service {
	return &service{repo: repo, cfg: cfg, redisClient: rdb}
}

func (s *service) GenerateSpdNumber() (string, error) {
	ctx := context.Background()
	latestSpd, err := s.repo.GetLatestSpdNumber(ctx)
	if err != nil {
		return "", err
	}

	if latestSpd == "" {
		return "ID-SPJ-001", nil
	}

	// Extract the last sequence of numbers using regex
	re := regexp.MustCompile(`[0-9]+$`)
	match := re.FindString(latestSpd)
	
	if match == "" {
		return "ID-SPJ-001", nil
	}

	num, err := strconv.Atoi(match)
	if err != nil {
		return "ID-SPJ-001", nil
	}

	// Determine the prefix
	prefix := strings.TrimSuffix(latestSpd, match)
	
	// Determine the padding length based on the original match length, defaulting to 3
	padding := len(match)
	if padding < 3 {
		padding = 3
	}

	return fmt.Sprintf("%s%0*d", prefix, padding, num+1), nil
}

func (s *service) invalidateCache(ctx context.Context, pattern string) {
	if s.redisClient == nil {
		return
	}
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("Error finding cache keys for pattern %s: %v\n", pattern, err)
		return
	}
	if len(keys) > 0 {
		err = s.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			fmt.Printf("Error deleting cache keys for pattern %s: %v\n", pattern, err)
		}
	}
}

func (s *service) InvalidateAllCache() {
	s.invalidateCache(context.Background(), "records:*")
}

func (s *service) CreateRecord(record *models.TravelRecord) error {
	if len(record.Locations) == 0 {
		return errors.New("at least one location is required")
	}

	// Calculate overall start and end dates if not provided or to ensure accuracy
	var minStart, maxEnd time.Time
	for i, loc := range record.Locations {
		if loc.EndDate.Before(loc.StartDate) {
			return fmt.Errorf("location %s: end date cannot be before start date", loc.Location)
		}
		if i == 0 || loc.StartDate.Before(minStart) {
			minStart = loc.StartDate
		}
		if i == 0 || loc.EndDate.After(maxEnd) {
			maxEnd = loc.EndDate
		}
	}
	record.StartDate = minStart
	record.EndDate = maxEnd

	overlapping, err := s.repo.GetOverlappingRecords(record.EmployeeID, record.StartDate, record.EndDate)
	if err != nil {
		return err
	}
	if len(overlapping) > 0 {
		return errors.New("employee is already assigned to a trip during these dates")
	}

	record.Status = "Draft"
	record.ReportStatus = "Pending"

	err = s.repo.CreateTravelRecord(record)
	if err == nil {
		s.invalidateCache(context.Background(), "records:*")

		// Send WhatsApp Notification
		go func() {
			user, err := s.repo.GetUserByID(context.Background(), record.EmployeeID)
			if err == nil && user != nil && user.NomorHP != "" {
				msg := fmt.Sprintf("*PEMBERITAHUAN PERJALANAN DINAS*\n\nHalo %s,\nAnda telah ditugaskan untuk perjalanan dinas baru.\n\n*Detail Penugasan:*\nNo. SPD: %s\nTujuan: %s, %s\nTanggal: %s s/d %s\nKeperluan: %s\n\nSilakan cek aplikasi Perjadin untuk detail selengkapnya dan mengunduh Surat Tugas.",
					user.Name,
					record.SPDNumber,
					record.Location,
					record.Province,
					record.StartDate.Format("02 Jan 2006"),
					record.EndDate.Format("02 Jan 2006"),
					record.Purpose,
				)
				
				if err := utils.SendWhatsAppMessage(s.cfg, user.NomorHP, msg); err != nil {
					fmt.Printf("Failed to send WhatsApp message to %s: %v\n", user.NomorHP, err)
				} else {
					fmt.Printf("WhatsApp notification sent to %s for SPD %s\n", user.NomorHP, record.SPDNumber)
				}
			}
		}()
	}
	return err
}

// CreateRecordDirect creates a record without overlap checking or WhatsApp notification.
// Used for bulk import operations.
func (s *service) CreateRecordDirect(record *models.TravelRecord) error {
	if len(record.Locations) == 0 {
		return errors.New("at least one location is required")
	}

	// Calculate overall start and end dates
	var minStart, maxEnd time.Time
	for i, loc := range record.Locations {
		if i == 0 || loc.StartDate.Before(minStart) {
			minStart = loc.StartDate
		}
		if i == 0 || loc.EndDate.After(maxEnd) {
			maxEnd = loc.EndDate
		}
	}
	record.StartDate = minStart
	record.EndDate = maxEnd

	if record.Status == "" {
		record.Status = "Draft"
	}
	if record.ReportStatus == "" {
		record.ReportStatus = "Pending"
	}

	err := s.repo.CreateTravelRecord(record)
	return err
}

func (s *service) GetRecords(filters map[string]interface{}) ([]models.TravelRecord, error) {
	ctx := context.Background()
	var statusFilter string
	if status, ok := filters["status"]; ok && status != "" {
		statusFilter = status.(string)
	}

	cacheKey := "records:all"
	if statusFilter != "" {
		cacheKey = fmt.Sprintf("records:status:%s", statusFilter)
	}

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var records []models.TravelRecord
			if err := json.Unmarshal([]byte(cached), &records); err == nil {
				return records, nil
			}
		}
	}

	records, err := s.repo.GetTravelRecords(filters)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(records)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 15*time.Minute)
		}
	}

	return records, nil
}

func (s *service) GetRecordByID(id uuid.UUID) (*models.TravelRecord, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("records:id:%s", id.String())

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var record models.TravelRecord
			if err := json.Unmarshal([]byte(cached), &record); err == nil {
				return &record, nil
			}
		}
	}

	record, err := s.repo.GetTravelRecordByID(id)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(record)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 15*time.Minute)
		}
	}

	return record, nil
}

func (s *service) UpdateRecord(record *models.TravelRecord) error {
	if len(record.Locations) > 0 {
		// Calculate overall start and end dates
		var minStart, maxEnd time.Time
		for i, loc := range record.Locations {
			if loc.EndDate.Before(loc.StartDate) {
				return fmt.Errorf("location %s: end date cannot be before start date", loc.Location)
			}
			if i == 0 || loc.StartDate.Before(minStart) {
				minStart = loc.StartDate
			}
			if i == 0 || loc.EndDate.After(maxEnd) {
				maxEnd = loc.EndDate
			}
		}
		record.StartDate = minStart
		record.EndDate = maxEnd
	}

	err := s.repo.UpdateTravelRecord(record)
	if err == nil {
		s.invalidateCache(context.Background(), "records:*")

		// NEW: If a report was updated, sync it to all other records in the same SPD group
		if record.Report != nil && record.SPDNumber != "" {
			go func(spd string, rep models.TravelReport, stNumber string, stDate time.Time, repStatus string) {
				allRecs, err := s.repo.GetTravelRecords(map[string]interface{}{})
				if err != nil {
					return
				}
				for _, r := range allRecs {
					if r.SPDNumber == spd && r.ID != record.ID {
						r.Report = &rep
						r.Report.TravelRecordID = r.ID
						if stNumber != "" {
							r.SuratTugasNumber = stNumber
						}
						if !stDate.IsZero() {
							r.SuratTugasDate = stDate
						}
						if repStatus != "" {
							r.ReportStatus = repStatus
						}
						s.repo.UpdateTravelRecord(&r)
					}
				}
			}(record.SPDNumber, *record.Report, record.SuratTugasNumber, record.SuratTugasDate, record.ReportStatus)
		}
	}
	return err
}

func (s *service) DeleteRecord(id uuid.UUID) error {
	err := s.repo.DeleteTravelRecord(id)
	if err == nil {
		go s.invalidateCache(context.Background(), "records:*")
	}
	return err
}

func (s *service) DeleteRecordsBySpd(ctx context.Context, spd string) error {
	err := s.repo.DeleteTravelRecordsBySpd(ctx, spd)
	if err == nil {
		go s.invalidateCache(context.Background(), "records:*")
	}
	return err
}
