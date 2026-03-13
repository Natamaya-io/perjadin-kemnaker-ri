package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(email, password, name, role string) (*models.User, error)
	Login(email, password string) (string, *models.User, error)
	GetDemoUsers() ([]map[string]string, error)
}

type service struct {
	userRepo    user.Repository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewService(userRepo user.Repository, cfg *config.Config, rdb *redis.Client) Service {
	return &service{userRepo: userRepo, cfg: cfg, redisClient: rdb}
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

func (s *service) Register(email, password, name, role string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Email:        email,
		Password:     string(hashedPassword),
		Name:         name,
		Role:         role,
		DemoPassword: password,
	}

	if err := s.userRepo.CreateUser(u); err != nil {
		return nil, err
	}

	s.invalidateCache(context.Background(), "users:*")
	s.invalidateCache(context.Background(), "demo_users")

	return u, nil
}

func (s *service) Login(email, password string) (string, *models.User, error) {
	u, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	u.SessionID = uuid.New().String()
	if err := s.userRepo.UpdateUser(u); err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateJWT(u, s.cfg)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

func (s *service) GetDemoUsers() ([]map[string]string, error) {
	ctx := context.Background()
	cacheKey := "demo_users"

	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var demoUsers []map[string]string
			if err := json.Unmarshal([]byte(cachedData), &demoUsers); err == nil {
				return demoUsers, nil
			}
		}
	}

	users, err := s.userRepo.GetUsers()
	if err != nil {
		return nil, err
	}

	var demoUsers []map[string]string
	for _, u := range users {
		demoUsers = append(demoUsers, map[string]string{
			"name":     u.Name,
			"email":    u.Email,
			"role":     u.Role,
			"password": "123", // Exposed for demo
		})
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(demoUsers)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, time.Hour)
		}
	}

	return demoUsers, nil
}
