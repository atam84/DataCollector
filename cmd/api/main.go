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

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	connectorHandler := handlers.NewConnectorHandler(connectorRepo, cfg)
	jobHandler := handlers.NewJobHandler(jobRepo, connectorRepo)

	// Health routes
	api.Get("/health", healthHandler.GetHealth)

	// Connector routes
	api.Post("/connectors", connectorHandler.CreateConnector)
	api.Get("/connectors", connectorHandler.GetConnectors)
	api.Get("/connectors/:id", connectorHandler.GetConnector)
	api.Put("/connectors/:id", connectorHandler.UpdateConnector)
	api.Delete("/connectors/:id", connectorHandler.DeleteConnector)
	api.Patch("/connectors/:id/sandbox", connectorHandler.ToggleSandboxMode)

	// Job routes
	api.Post("/jobs", jobHandler.CreateJob)
	api.Get("/jobs", jobHandler.GetJobs)
	api.Get("/jobs/:id", jobHandler.GetJob)
	api.Put("/jobs/:id", jobHandler.UpdateJob)
	api.Delete("/jobs/:id", jobHandler.DeleteJob)
	api.Post("/jobs/:id/pause", jobHandler.PauseJob)
	api.Post("/jobs/:id/resume", jobHandler.ResumeJob)

	// Connector-specific job routes
	api.Get("/connectors/:exchangeId/jobs", jobHandler.GetJobsByConnector)

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
