package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/auth"
	"github.com/kemnaker/perjadin-backend/internal/domain/master"
	"github.com/kemnaker/perjadin-backend/internal/domain/record"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/middleware"
	"github.com/kemnaker/perjadin-backend/internal/seeder"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
	"github.com/kemnaker/perjadin-backend/pkg/database"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func runMigrations(db *sql.DB, sugar *zap.SugaredLogger) {
	sugar.Info("Running migrations...")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		sugar.Fatalf("Could not create postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/db/migrations",
		"postgres", driver)
	if err != nil {
		sugar.Fatalf("Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		sugar.Fatalf("Migration failed: %v", err)
	}
	sugar.Info("Migrations applied successfully.")
}

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
	defer func() {
		sugar.Info("Closing database connection...")
		if err := db.Close(); err != nil {
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

	// 4. Run Migrations
	runMigrations(db, sugar)

	// 5. Seed Data (if enabled)
	if os.Getenv("SEED_DB") == "true" {
		sugar.Info("Seeding database...")
		seeder.Seed(db, rdb)
	}

	// 6. Initialize Layers
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, cfg, rdb)
	userHandler := user.NewHandler(userSvc)

	masterRepo := master.NewRepository(db)
	masterSvc := master.NewService(masterRepo, rdb)
	masterHandler := master.NewHandler(masterSvc)

	recordRepo := record.NewRepository(db)
	recordSvc := record.NewService(recordRepo, cfg, rdb)
	recordHandler := record.NewHandler(recordSvc, userRepo, cfg, masterSvc)

	authSvc := auth.NewService(userRepo, cfg, rdb)
	authHandler := auth.NewHandler(authSvc)

	// Ensure uploads directory exists
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		sugar.Warnf("Failed to create uploads directory: %v", err)
	}

	// 7. Setup Echo
	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.BodyLimit("50M"))

	// Static files
	e.Static("/uploads", "uploads")

	// 8. Routes
	api := e.Group("/api/v1")

	// Auth Routes
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/register", authHandler.Register)
	api.GET("/auth/demo-users", authHandler.GetDemoUsers)

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware(cfg, userRepo))
	{
		protected.POST("/upload", recordHandler.UploadFile)

		protected.GET("/records", recordHandler.GetRecords)
		protected.POST("/records", recordHandler.CreateRecord)
		protected.GET("/records/:id", recordHandler.GetRecordByID)
		protected.GET("/records/:id/spd-pdf", recordHandler.ExportSpdPDF)
		protected.GET("/records/:id/laporan-pdf", recordHandler.ExportLaporanPDF)
		protected.GET("/records/:id/rincian-pdf", recordHandler.ExportRincianPDF)
		protected.PUT("/records/:id", recordHandler.UpdateRecord)
		protected.DELETE("/records/:id", recordHandler.DeleteRecord)
		protected.DELETE("/records/spd/:spd", recordHandler.DeleteRecordsBySpd)

		// User Management
		protected.GET("/users", userHandler.GetUsers)
		protected.POST("/users", userHandler.CreateUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		protected.DELETE("/users/:id", userHandler.DeleteUser)

		// Master Data
		protected.GET("/master/provinces", masterHandler.GetProvinces)
		protected.GET("/master/sbm-rates", masterHandler.GetSBMRates)
		protected.GET("/master/settings", masterHandler.GetSettings)
		protected.PUT("/master/settings", masterHandler.UpdateSettings, middleware.RoleMiddleware("super_admin", "kasubag"))
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
