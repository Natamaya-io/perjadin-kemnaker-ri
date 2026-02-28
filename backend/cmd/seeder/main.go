package main

import (
	"fmt"
	"log"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to DB using config
	// Note: When running via 'go run', ensure DB_HOST points to localhost if outside docker,
	// or use container name if inside docker. We'll rely on the environment variables passed to it.
	dsn := "host=" + cfg.Database.Host + " user=" + cfg.Database.User + " password=" + cfg.Database.Password + " dbname=" + cfg.Database.Name + " port=" + cfg.Database.Port + " sslmode=" + cfg.Database.SSLMode

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database. Seeding users...")

	// Default Password "123"
	password := "123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPwdStr := string(hashedPassword)

	users := []models.User{
		{
			Name:         "Super Admin",
			Email:        "superadmin@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "super_admin",
			DemoPassword: password,
			NomorHP:      "081200000000",
		},
		{
			Name:         "Keuangan",
			Email:        "keuangan@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "keuangan",
			DemoPassword: password,
			NomorHP:      "081200000001",
		},
		{
			Name:         "Kasubag",
			Email:        "kasubag@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "kasubag",
			DemoPassword: password,
			NomorHP:      "081200000002",
		},
	}

	for i := 1; i <= 49; i++ {
		name := fmt.Sprintf("Protokol %d", i)
		if i == 1 {
			name = "Vito"
		} else if i == 2 {
			name = "Keneth"
		}

		email := fmt.Sprintf("protokol%d@kemnaker.go.id", i)
		if i == 1 {
			email = "vito@kemnaker.go.id"
		} else if i == 2 {
			email = "keneth@kemnaker.go.id"
		}

		users = append(users, models.User{
			Name:         name,
			Email:        email,
			Password:     hashedPwdStr,
			Role:         "protokol", // Protokol role
			DemoPassword: password,
			Jabatan:      "Protokol",
			Pangkat:      "Penata Muda",
			Golongan:     "III/a",
			TingkatBiaya: "C",
			NIP:          fmt.Sprintf("198001012005011%03d", i),
			NomorHP:      fmt.Sprintf("081234567%03d", i),
		})
	}

	for _, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&u).Error; err != nil {
					log.Printf("Failed to create user %s: %v", u.Email, err)
				} else {
					log.Printf("Created user: %s", u.Email)
				}
			} else {
				log.Printf("Error checking user %s: %v", u.Email, err)
			}
		} else {
			log.Printf("User %s already exists.", u.Email)
		}
	}

	log.Println("Seeding complete.")
}
