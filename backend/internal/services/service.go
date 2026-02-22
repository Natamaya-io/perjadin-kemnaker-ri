package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/repository"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *repository.Repository
	Config *config.Config
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{Repo: repo, Config: cfg}
}

// --- Auth Service ---

func (s *Service) Register(email, password, name, role string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Role:     role,
	}

	if err := s.Repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (string, *models.User, error) {
	user, err := s.Repo.GetUserByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateJWT(user, s.Config)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// --- Record Service ---

func (s *Service) CreateRecord(record *models.TravelRecord) error {
	// Business Logic: Validate dates, check budget (mocked for now), etc.
	if record.EndDate.Before(record.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	
	// Set initial status
	record.Status = "Submitted"
	record.ReportStatus = "Pending"
	
	return s.Repo.CreateTravelRecord(record)
}

func (s *Service) GetRecords(filters map[string]interface{}) ([]models.TravelRecord, error) {
	return s.Repo.GetTravelRecords(filters)
}

func (s *Service) GetRecordByID(id uuid.UUID) (*models.TravelRecord, error) {
	return s.Repo.GetTravelRecordByID(id)
}
