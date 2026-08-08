package main

import (
	"context"
	"database/sql"
	"net/http"
	_ "net/http/pprof" // Zero-Trust memory profiler
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/auth"
	"github.com/kemnaker/perjadin-backend/internal/domain/master"
	"github.com/kemnaker/perjadin-backend/internal/domain/record"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/domain/dalkot"
	"github.com/kemnaker/perjadin-backend/internal/domain/gup"
	"github.com/kemnaker/perjadin-backend/internal/domain/chatbot"

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
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		sugar.Fatalf("Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if strings.Contains(err.Error(), "Dirty database") {
			sugar.Warnf("Detected dirty database. Forcing version 12 and retrying...")
			// We force to 12 because migration 13 failed
			if forceErr := m.Force(12); forceErr != nil {
				sugar.Fatalf("Failed to force migration version: %v", forceErr)
			}
			// Retry Up after forcing
			if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
				sugar.Fatalf("Migration failed after force: %v", retryErr)
			}
		} else {
			sugar.Fatalf("Migration failed: %v", err)
		}
	}
	sugar.Info("Migrations applied successfully.")
}

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync() //nolint:errcheck
	}()
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
		if closeErr := db.Close(); closeErr != nil {
			sugar.Errorf("Error closing database connection: %v", closeErr)
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

	dalkotRepo := dalkot.NewRepository(db)
	dalkotSvc := dalkot.NewService(dalkotRepo, userRepo)
	dalkotHandler := dalkot.NewHandler(dalkotSvc)

	gupRepo := gup.NewRepository(db)
	gupSvc := gup.NewService(gupRepo)
	gupHandler := gup.NewHandler(gupSvc)

	chatbotRepo := chatbot.NewRepository(db)
	chatbotSvc := chatbot.NewService(chatbotRepo)
	chatbotHandler := chatbot.NewHandler(chatbotSvc)

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
		protected.GET("/auth/me", authHandler.GetMe)
		protected.PUT("/auth/change-password", authHandler.ChangePassword)
		
		protected.POST("/upload", recordHandler.UploadFile)

		protected.GET("/dashboard/summary", recordHandler.GetDashboardSummary)
		protected.GET("/records", recordHandler.GetRecords)
		protected.GET("/records/paginated", recordHandler.GetPaginatedRecords)
		protected.POST("/records", recordHandler.CreateRecord)
		protected.GET("/records/:id", recordHandler.GetRecordByID)
		protected.GET("/records/:id/spd-stream", recordHandler.ExportSpdPDF)
		protected.GET("/records/:id/spd-docx", recordHandler.ExportSpdDocx)
		protected.GET("/records/:id/laporan-stream", recordHandler.ExportLaporanPDF)
		protected.GET("/records/:id/laporan-docx", recordHandler.ExportLaporanDocx)
		protected.GET("/records/:id/rincian-stream", recordHandler.ExportRincianPDF)
		protected.GET("/records/:id/rincian-docx", recordHandler.ExportRincianDocx)
		protected.PUT("/records/:id", recordHandler.UpdateRecord)
		protected.DELETE("/records/:id", recordHandler.DeleteRecord)
		protected.DELETE("/records/spd/:spd", recordHandler.DeleteRecordsBySpd)

		// Dalkot Routes
		protected.POST("/dalkot", dalkotHandler.CreateRecord)
		protected.GET("/dalkot", dalkotHandler.GetRecords)
		protected.GET("/dalkot/:id", dalkotHandler.GetRecordByID)
		protected.GET("/dalkot/:id/laporan-stream", dalkotHandler.ExportLaporanPDF)
		protected.PUT("/dalkot/:id", dalkotHandler.UpdateRecord)
		protected.DELETE("/dalkot/:id", dalkotHandler.DeleteRecord)
		protected.GET("/dalkot/locations/all", dalkotHandler.GetLocations)
		protected.GET("/dalkot-rates", dalkotHandler.GetRates)
		protected.POST("/dalkot/assignments", dalkotHandler.AddAssignment)
		protected.DELETE("/dalkot/assignments/:assignmentId", dalkotHandler.RemoveAssignment)

		// GUP Routes
		gupGroup := protected.Group("/gup")
		gupGroup.GET("/master-data", gupHandler.GetMasterData)
		gupGroup.GET("/laporan", gupHandler.GetLaporan)
		gupGroup.GET("/dashboard", gupHandler.GetDashboardSummary)
		gupGroup.GET("/next-spm", gupHandler.GetNextBusinessID)
		gupGroup.POST("/pengajuan", gupHandler.CreateTransaction)
		
		protected.GET("/gup/pengajuan", gupHandler.GetTransactions)
		protected.GET("/gup/pengajuan/:id", gupHandler.GetTransactionByID)
		protected.PUT("/gup/pengajuan/:id", gupHandler.UpdateTransaction)
		protected.DELETE("/gup/pengajuan/:id", gupHandler.DeleteTransaction)
		
		protected.GET("/gup/laporan", gupHandler.GetLaporan)
		protected.POST("/gup/laporan/budget", gupHandler.SaveBudget)
		protected.GET("/gup/ls", gupHandler.GetMonthlyLS)
		protected.POST("/gup/ls", gupHandler.SaveMonthlyLS)
		protected.GET("/gup/data", gupHandler.GetBudgets)
		protected.GET("/gup/master-data", gupHandler.GetMasterData)
		protected.POST("/gup/master-data/account-codes", gupHandler.CreateAccountCode)
		protected.PUT("/gup/master-data/account-codes/:id", gupHandler.UpdateAccountCode)
		protected.DELETE("/gup/master-data/account-codes/:id", gupHandler.DeleteAccountCode)
		protected.POST("/gup/master-data/procurement-types", gupHandler.CreateProcurementType)
		protected.PUT("/gup/master-data/procurement-types/:id", gupHandler.UpdateProcurementType)
		protected.DELETE("/gup/master-data/procurement-types/:id", gupHandler.DeleteProcurementType)

		// Chatbot
		protected.POST("/chatbot/ask", chatbotHandler.Ask)
		protected.GET("/chatbot/report", chatbotHandler.GetSnapshot)

		// User Management
		protected.GET("/users", userHandler.GetUsers)
		protected.POST("/users", userHandler.CreateUser, middleware.RoleMiddleware("super_admin"))
		protected.PUT("/users/:id", userHandler.UpdateUser, middleware.RoleMiddleware("super_admin"))
		protected.DELETE("/users/:id", userHandler.DeleteUser, middleware.RoleMiddleware("super_admin"))

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
	
	// Start pprof on hidden port 6060 (Bind ke 0.0.0.0 agar bisa diakses host Docker)
	go func() {
		sugar.Info("Starting pprof on port 6060 for internal memory telemetry")
		if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
			sugar.Warnf("pprof server failed: %v", err)
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
