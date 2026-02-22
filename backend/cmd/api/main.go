package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/handlers"
	"github.com/kemnaker/perjadin-backend/internal/middleware"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/pkg/database"
	"github.com/kemnaker/perjadin-backend/internal/repository"
	"github.com/kemnaker/perjadin-backend/internal/services"
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

	// 4. Auto Migrate
	sugar.Info("Migrating database schemas...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Employee{},
		&models.TravelRecord{},
		&models.TravelCost{},
		&models.TravelReport{},
	)
	if err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}

	// 5. Initialize Layers
	repo := repository.NewRepository(db)
	svc := services.NewService(repo, cfg)
	authHandler := handlers.NewAuthHandler(svc)
	recordHandler := handlers.NewRecordHandler(svc)

	// 6. Setup Echo
	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// 7. Routes
	api := e.Group("/api/v1")

	// Auth Routes
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/register", authHandler.Register)

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware(cfg))
	{
		protected.GET("/records", recordHandler.GetRecords)
		protected.POST("/records", recordHandler.CreateRecord)
		protected.GET("/records/:id", recordHandler.GetRecordByID)
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
