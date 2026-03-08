package main

import (
	"log"

	"github.com/kemnaker/perjadin-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=perjadin_db port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	user := models.User{
		Email:        "vito@kemnaker.go.id",
		Password:     string(hashedPassword),
		Name:         "Vito",
		Role:         "protokol",
		NomorHP:      "082382012932",
		NIP:          "-",
		Pangkat:      "-",
		Golongan:     "-",
		Jabatan:      "Staf Protokol",
		TingkatBiaya: "C",
		DemoPassword: "password",
	}

	result := db.Where("email = ?", user.Email).FirstOrCreate(&user)
	if result.Error != nil {
		log.Fatal("Failed to create user:", result.Error)
	}

	log.Println("Successfully created user vito@kemnaker.go.id!")
}
