// cmd/server/main.go

// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api/v1

package main

import (
	"context"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ZnNr/subscription-service/docs" // <-- новый импорт
	"github.com/ZnNr/subscription-service/internal/config"
	"github.com/ZnNr/subscription-service/internal/handler"
	"github.com/ZnNr/subscription-service/internal/repository"
	"github.com/ZnNr/subscription-service/internal/service"
	"github.com/ZnNr/subscription-service/pkg/database"
	"github.com/ZnNr/subscription-service/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)

	// Connect to database
	db, err := database.ConnectPostgres(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize repository, service, and handler
	repo := repository.NewPostgresRepository(db)
	svc := service.NewSubscriptionService(repo)
	h := handler.NewHandler(svc)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinLogger())

	// Настраиваем Swagger
	// Настройка Swagger
	docs.SwaggerInfo.Title = "Subscription Service API"
	docs.SwaggerInfo.Description = "REST API for managing user subscriptions"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h.SetupRoutes(router)

	// Start server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
