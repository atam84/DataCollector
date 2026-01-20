package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/yourusername/datacollector/internal/api/handlers"
	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to MongoDB
	db, err := repository.Connect(cfg.Database.URI, cfg.Database.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	log.Println("Connected to MongoDB successfully")

	// Fix existing jobs that don't have next_run_time set
	jobRepoForFix := repository.NewJobRepository(db)
	fixCtx, fixCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer fixCancel()
	fixedCount, err := jobRepoForFix.FixMissingNextRunTime(fixCtx)
	if err != nil {
		log.Printf("Warning: Failed to fix jobs: %v", err)
	} else if fixedCount > 0 {
		log.Printf("Fixed %d jobs with missing next_run_time", fixedCount)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "DataCollector API",
		ServerHeader: "DataCollector",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check database health
		dbStatus := "healthy"
		if err := db.HealthCheck(ctx); err != nil {
			dbStatus = "unhealthy"
		}

		return c.JSON(fiber.Map{
			"status": "ok",
			"timestamp": time.Now().Unix(),
			"database": dbStatus,
			"sandbox_mode": cfg.Exchange.SandboxMode,
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Initialize repositories
	connectorRepo := repository.NewConnectorRepository(db)
	jobRepo := repository.NewJobRepository(db)
	ohlcvRepo := repository.NewOHLCVRepository(db)

	// Initialize services
	jobExecutor := service.NewJobExecutor(jobRepo, connectorRepo, ohlcvRepo, cfg)
	jobScheduler := service.NewJobScheduler(jobRepo, jobExecutor)
	recalcService := service.NewRecalculatorService(jobRepo, connectorRepo, ohlcvRepo)

	// Start automatic job scheduler
	jobScheduler.Start()
	defer jobScheduler.Stop()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	connectorHandler := handlers.NewConnectorHandler(connectorRepo, jobRepo, cfg)
	jobHandler := handlers.NewJobHandler(jobRepo, connectorRepo, jobExecutor)
	indicatorHandler := handlers.NewIndicatorHandler(ohlcvRepo, recalcService)

	// Health routes
	api.Get("/health", healthHandler.GetHealth)

	// Connector routes
	api.Post("/connectors", connectorHandler.CreateConnector)
	api.Get("/connectors", connectorHandler.GetConnectors)
	api.Get("/connectors/:id", connectorHandler.GetConnector)
	api.Put("/connectors/:id", connectorHandler.UpdateConnector)
	api.Delete("/connectors/:id", connectorHandler.DeleteConnector)
	api.Patch("/connectors/:id/sandbox", connectorHandler.ToggleSandboxMode)
	api.Post("/connectors/:id/suspend", connectorHandler.SuspendConnector)
	api.Post("/connectors/:id/resume", connectorHandler.ResumeConnector)

	// Job routes (queue route MUST come before :id routes)
	api.Post("/jobs", jobHandler.CreateJob)
	api.Get("/jobs", jobHandler.GetJobs)
	api.Get("/jobs/queue", jobHandler.GetQueue)
	api.Get("/jobs/:id", jobHandler.GetJob)
	api.Put("/jobs/:id", jobHandler.UpdateJob)
	api.Delete("/jobs/:id", jobHandler.DeleteJob)
	api.Post("/jobs/:id/pause", jobHandler.PauseJob)
	api.Post("/jobs/:id/resume", jobHandler.ResumeJob)
	api.Post("/jobs/:id/execute", jobHandler.ExecuteJob)

	// Connector-specific job routes
	api.Get("/connectors/:exchangeId/jobs", jobHandler.GetJobsByConnector)

	// Indicator data retrieval routes (timeframe before symbol to avoid slash conflicts)
	api.Get("/indicators/:exchange/:timeframe/:symbol+/latest", indicatorHandler.GetLatestIndicators)
	api.Get("/indicators/:exchange/:timeframe/:symbol+/range", indicatorHandler.GetIndicatorRange)
	api.Get("/indicators/:exchange/:timeframe/:symbol+/:indicator", indicatorHandler.GetSpecificIndicator)

	// Indicator recalculation routes
	api.Post("/jobs/:id/indicators/recalculate", indicatorHandler.RecalculateJob)
	api.Post("/connectors/:id/indicators/recalculate", indicatorHandler.RecalculateConnector)

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s (Sandbox Mode: %v)", address, cfg.Exchange.SandboxMode)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(address); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
