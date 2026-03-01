package main

import (
	"log"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/seeder"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	dsn := "host=" + cfg.Database.Host + " user=" + cfg.Database.User + " password=" + cfg.Database.Password + " dbname=" + cfg.Database.Name + " port=" + cfg.Database.Port + " sslmode=" + cfg.Database.SSLMode

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database.")

	// Ensure Migrations are up to date before seeding
	log.Println("Running AutoMigrate...")
	err = db.AutoMigrate(
		&models.User{},
		&models.TravelRecord{},
		&models.TravelCost{},
		&models.TravelReport{},
		&models.Province{},
		&models.SBMRate{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Call the shared seeder logic
	seeder.Seed(db)
}
