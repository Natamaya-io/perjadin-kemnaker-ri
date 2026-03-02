package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/handlers"
	"github.com/kemnaker/perjadin-backend/internal/middleware"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/repository"
	"github.com/kemnaker/perjadin-backend/internal/seeder"
	"github.com/kemnaker/perjadin-backend/internal/services"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
	"github.com/kemnaker/perjadin-backend/pkg/database"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// 3. Connect to Database
	db, err := database.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		sugar.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer func() {
		sugar.Info("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			sugar.Errorf("Error closing database connection: %v", err)
		}
	}()

	// 3.5 Connect to Redis
	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		sugar.Warnf("Failed to connect to Redis: %v. Running without cache.", err)
		rdb = nil
	} else {
		defer func() {
			if err := rdb.Close(); err != nil {
				sugar.Errorf("Error closing Redis connection: %v", err)
			}
		}()
	}

	// 4. Auto Migrate
	sugar.Info("Migrating database schemas...")
	
	// Drop the unique index to allow multiple records per SPD
	if err := db.Exec("DROP INDEX IF EXISTS idx_travel_records_spd_number;").Error; err != nil {
		sugar.Warnf("Failed to drop index idx_travel_records_spd_number: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.TravelRecord{},
		&models.TravelCost{},
		&models.TravelReport{},
		// Master Data
		&models.Province{},
		&models.SBMRate{},
	)
	if err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}

	// Fix missing unique constraint for ON CONFLICT during report update
	db.Exec("ALTER TABLE travel_reports ADD CONSTRAINT travel_reports_record_id_key UNIQUE (travel_record_id);")
	db.Exec("ALTER TABLE travel_costs ADD CONSTRAINT travel_costs_record_id_key UNIQUE (travel_record_id);")

	// 5. Seed Data (if enabled)
	if os.Getenv("SEED_DB") == "true" {
		sugar.Info("Seeding database...")
		seeder.Seed(db)
	}

	// 6. Initialize Layers
	repo := repository.NewRepository(db)
	svc := services.NewService(repo, cfg, rdb)
	authHandler := handlers.NewAuthHandler(svc)
	recordHandler := handlers.NewRecordHandler(svc)
	// employeeHandler := handlers.NewEmployeeHandler(svc) // Removed
	userHandler := handlers.NewUserHandler(svc)

	// 7. Setup Echo
	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// 8. Routes
	api := e.Group("/api/v1")

	// Auth Routes
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/register", authHandler.Register)
	api.GET("/auth/demo-users", authHandler.GetDemoUsers)

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware(cfg, repo))
	{
		protected.GET("/records", recordHandler.GetRecords)
		protected.POST("/records", recordHandler.CreateRecord)
		protected.GET("/records/:id", recordHandler.GetRecordByID)
		protected.PUT("/records/:id", recordHandler.UpdateRecord)
		protected.DELETE("/records/:id", recordHandler.DeleteRecord)

		// Employee Management Removed
		
		// User Management
		protected.GET("/users", userHandler.GetUsers)
		protected.POST("/users", userHandler.CreateUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		protected.DELETE("/users/:id", userHandler.DeleteUser)
	}

	// 8. Start Server with Graceful Shutdown
	go func() {
		if err := e.Start(":" + cfg.App.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	sugar.Infof("Server started on port %s", cfg.App.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	sugar.Info("Server exited properly")
}
