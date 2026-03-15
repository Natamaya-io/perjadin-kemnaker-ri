package record

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	CreateRecord(record *models.TravelRecord) error
	GetRecords(filters map[string]interface{}) ([]models.TravelRecord, error)
	GetRecordByID(id uuid.UUID) (*models.TravelRecord, error)
	UpdateRecord(record *models.TravelRecord) error
	DeleteRecord(id uuid.UUID) error
}

type service struct {
	repo        Repository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewService(repo Repository, cfg *config.Config, rdb *redis.Client) Service {
	return &service{repo: repo, cfg: cfg, redisClient: rdb}
}

func (s *service) invalidateCache(ctx context.Context, pattern string) {
	if s.redisClient == nil {
		return
	}
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		fmt.Printf("Error invalidating cache for pattern %s: %v\n", pattern, err)
	}
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
	}
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
	}
	return err
}

func (s *service) DeleteRecord(id uuid.UUID) error {
	err := s.repo.DeleteTravelRecord(id)
	if err == nil {
		s.invalidateCache(context.Background(), "records:*")
	}
	return err
}
