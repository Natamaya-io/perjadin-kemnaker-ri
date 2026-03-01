package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/repository"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo        *repository.Repository
	Config      *config.Config
	RedisClient *redis.Client
}

func NewService(repo *repository.Repository, cfg *config.Config, rdb *redis.Client) *Service {
	return &Service{Repo: repo, Config: cfg, RedisClient: rdb}
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
		DemoPassword: password,
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

	// Generate new session ID
	user.SessionID = uuid.New().String()
	if err := s.Repo.UpdateUser(user); err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateJWT(user, s.Config)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *Service) GetUsers() ([]models.User, error) {
	return s.Repo.GetUsers()
}

func (s *Service) CreateUser(user *models.User) error {
	user.DemoPassword = user.Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.Repo.CreateUser(user)
}

func (s *Service) UpdateUser(user *models.User, newPassword string) error {
	if newPassword != "" {
		user.DemoPassword = newPassword
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	return s.Repo.UpdateUser(user)
}

func (s *Service) DeleteUser(id uuid.UUID) error {
	return s.Repo.DeleteUser(id)
}

// --- Record Service ---

func (s *Service) CreateRecord(record *models.TravelRecord) error {
	// Business Logic: Validate dates, check budget (mocked for now), etc.
	if record.EndDate.Before(record.StartDate) {
		return errors.New("end date cannot be before start date")
	}
	
	// Set initial status
	record.Status = "Draft"
	record.ReportStatus = "Pending"
	
	return s.Repo.CreateTravelRecord(record)
}

func (s *Service) GetRecords(filters map[string]interface{}) ([]models.TravelRecord, error) {
	return s.Repo.GetTravelRecords(filters)
}

func (s *Service) GetRecordByID(id uuid.UUID) (*models.TravelRecord, error) {
	return s.Repo.GetTravelRecordByID(id)
}

func (s *Service) UpdateRecord(record *models.TravelRecord) error {
	// Add business logic validation if needed
	return s.Repo.UpdateTravelRecord(record)
}

func (s *Service) DeleteRecord(id uuid.UUID) error {
	return s.Repo.DeleteTravelRecord(id)
}
