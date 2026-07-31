package dalkot

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

type Service interface {
	CreateRecord(record *models.DalkotRecord, assignments []models.DalkotAssignment) error
	GetRecordByID(id uuid.UUID) (*models.DalkotRecord, error)
	GetRecords() ([]models.DalkotRecord, error)
	UpdateRecord(record *models.DalkotRecord) error
	DeleteRecord(id uuid.UUID) error

	AddAssignment(assignment *models.DalkotAssignment) error
	RemoveAssignment(assignmentID uuid.UUID) error

	GetLocations() ([]models.DalkotLocation, error)
}

type service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) CreateRecord(record *models.DalkotRecord, assignments []models.DalkotAssignment) error {
	if record.Category == "" || record.Official == "" || record.DalkotType == "" || record.ActivityName == "" || record.Location == "" {
		return errors.New("missing required fields for dalkot record")
	}

	if record.Status == "" {
		record.Status = "Draft"
	}

	if err := s.repo.CreateRecord(record); err != nil {
		return err
	}

	for _, assignment := range assignments {
		assignment.DalkotRecordID = record.ID
		if assignment.Status == "" {
			assignment.Status = "Pending"
		}
		if err := s.repo.CreateAssignment(&assignment); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) GetRecordByID(id uuid.UUID) (*models.DalkotRecord, error) {
	record, err := s.repo.GetRecordByID(id)
	if err != nil {
		return nil, err
	}

	assignments, err := s.repo.GetAssignmentsByRecordID(id)
	if err != nil {
		return nil, err
	}

	for i := range assignments {
		user, err := s.userRepo.GetUserByID(assignments[i].UserID)
		if err == nil {
			assignments[i].User = user
		}
	}

	record.Assignments = assignments
	return record, nil
}

func (s *service) GetRecords() ([]models.DalkotRecord, error) {
	records, err := s.repo.GetRecords()
	if err != nil {
		return nil, err
	}

	for i := range records {
		assignments, err := s.repo.GetAssignmentsByRecordID(records[i].ID)
		if err == nil {
			for j := range assignments {
				user, err := s.userRepo.GetUserByID(assignments[j].UserID)
				if err == nil {
					assignments[j].User = user
				}
			}
			records[i].Assignments = assignments
		}
	}

	return records, nil
}

func (s *service) UpdateRecord(record *models.DalkotRecord) error {
	return s.repo.UpdateRecord(record)
}

func (s *service) DeleteRecord(id uuid.UUID) error {
	return s.repo.DeleteRecord(id)
}

func (s *service) AddAssignment(assignment *models.DalkotAssignment) error {
	if assignment.Status == "" {
		assignment.Status = "Pending"
	}
	return s.repo.CreateAssignment(assignment)
}

func (s *service) RemoveAssignment(assignmentID uuid.UUID) error {
	return s.repo.DeleteAssignment(assignmentID)
}

func (s *service) GetLocations() ([]models.DalkotLocation, error) {
	return s.repo.GetLocations()
}
