package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(host, port, user, password, name, sslMode string) (*gorm.DB, error) {
	// First connect to default 'postgres' db to check and create the target db
	dsnDefault := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s client_encoding=UTF8",
		host, user, password, port, sslMode)

	dbDefault, err := gorm.Open(postgres.Open(dsnDefault), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Printf("Warning: failed to connect to default postgres database to check for existence: %v", err)
	} else {
		// Check if database exists
		var checkDb int
		err = dbDefault.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", name).Scan(&checkDb).Error
		if err == nil && checkDb == 0 {
			log.Printf("Database %s does not exist. Creating...", name)
			// We cannot use parameterized queries for CREATE DATABASE
			err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", name)).Error
			if err != nil {
				log.Printf("Failed to create database: %v", err)
			} else {
				log.Printf("Database %s created successfully.", name)
			}
		}

		sqlDBDefault, _ := dbDefault.DB()
		if sqlDBDefault != nil {
			sqlDBDefault.Close()
		}
	}

	// Now connect to the target database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s client_encoding=UTF8",
		host, user, password, name, port, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection Pooling Configuration
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL database with connection pooling enabled")
	return db, nil
}
