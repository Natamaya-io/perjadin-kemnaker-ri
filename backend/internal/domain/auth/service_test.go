package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)

	// Mock Configuration
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expiry: 24,
		},
	}

	// Create service with mock repo and nil redis (no caching needed for this test)
	service := NewService(mockRepo, cfg, nil)

	// Setup fake user with hashed password
	rawPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	
	fakeUser := &models.User{
		Base: models.Base{ID: uuid.New()},
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	// Setup expectations
	mockRepo.On("GetUserByEmail", "test@example.com").Return(fakeUser, nil)
	mockRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	token, user, err := service.Login("test@example.com", "password123")

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.SessionID) // Session ID should be generated

	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg, nil)

	// Expectation: repo returns error for email not found
	mockRepo.On("GetUserByEmail", "wrong@example.com").Return(nil, errors.New("not found"))

	// Execute
	token, user, err := service.Login("wrong@example.com", "password123")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg, nil)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	fakeUser := &models.User{
		Base: models.Base{ID: uuid.New()},
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmail", "test@example.com").Return(fakeUser, nil)

	// Execute with wrong password
	token, user, err := service.Login("test@example.com", "wrongpassword")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Nil(t, user)

	// UpdateUser should NEVER be called if password fails
	mockRepo.AssertNotCalled(t, "UpdateUser")
}
