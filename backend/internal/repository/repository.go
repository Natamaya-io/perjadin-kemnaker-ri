package repository

import (
	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// --- User ---
func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) GetUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

// --- Travel Record ---
func (r *Repository) CreateTravelRecord(record *models.TravelRecord) error {
	return r.db.Create(record).Error
}

func (r *Repository) GetTravelRecords(filters map[string]interface{}) ([]models.TravelRecord, error) {
	var records []models.TravelRecord
	query := r.db.Preload("Employee").Preload("Creator").Preload("Cost").Preload("Report")

	// Apply basic filters if any (simplified)
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at desc").Limit(1000).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetTravelRecordByID(id uuid.UUID) (*models.TravelRecord, error) {
	var record models.TravelRecord
	if err := r.db.Preload("Employee").Preload("Creator").Preload("Cost").Preload("Report").First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) UpdateTravelRecord(record *models.TravelRecord) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(record).Error
}

func (r *Repository) DeleteTravelRecord(id uuid.UUID) error {
	return r.db.Delete(&models.TravelRecord{}, id).Error
}
