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
	"github.com/gofiber/swagger"

	"github.com/yourusername/datacollector/internal/api/handlers"
	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"

	_ "github.com/yourusername/datacollector/docs"
)

// @title DataCollector API
// @version 1.1.0
// @description Cryptocurrency market data collection service with OHLCV data, technical indicators, and job management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/yourusername/datacollector
// @contact.email support@datacollector.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1

// @schemes http https

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
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"database":  dbStatus,
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API v1 routes
	api := app.Group("/api/v1")

	// Initialize repositories
	connectorRepo := repository.NewConnectorRepository(db)
	jobRepo := repository.NewJobRepository(db)
	ohlcvRepo := repository.NewOHLCVRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	retentionRepo := repository.NewRetentionRepository(db)
	indicatorConfigRepo := repository.NewIndicatorConfigRepository(db)
	qualityRepo := repository.NewQualityRepository(db)
	mlExportRepo := repository.NewMLExportRepository(db)

	// Initialize services
	rateLimiter := service.NewRateLimiter(connectorRepo)
	ccxtService := service.NewCCXTServiceWithRateLimiter(rateLimiter)
	jobExecutor := service.NewJobExecutor(jobRepo, connectorRepo, ohlcvRepo, cfg)
	jobScheduler := service.NewJobScheduler(jobRepo, jobExecutor)
	recalcService := service.NewRecalculatorService(jobRepo, connectorRepo, ohlcvRepo)
	alertService := service.NewAlertService(alertRepo, jobRepo, connectorRepo)
	retentionService := service.NewRetentionService(retentionRepo)
	qualityService := service.NewQualityService(qualityRepo, ohlcvRepo, jobRepo, ccxtService, connectorRepo, rateLimiter)
	mlExportService := service.NewMLExportService(ohlcvRepo, jobRepo, mlExportRepo)

	// Start automatic job scheduler
	jobScheduler.Start()
	defer jobScheduler.Stop()

	// Start quality scheduler (runs every 1 hour)
	qualityScheduler := service.NewQualityScheduler(qualityService, 1*time.Hour)
	qualityScheduler.Start()
	defer qualityScheduler.Stop()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	connectorHandler := handlers.NewConnectorHandlerWithOHLCV(connectorRepo, jobRepo, ohlcvRepo, cfg)
	jobHandler := handlers.NewJobHandler(jobRepo, connectorRepo, ohlcvRepo, jobExecutor)
	indicatorHandler := handlers.NewIndicatorHandler(ohlcvRepo, recalcService)
	indicatorConfigHandler := handlers.NewIndicatorConfigHandler(indicatorConfigRepo)
	alertHandler := handlers.NewAlertHandler(alertRepo, alertService)
	retentionHandler := handlers.NewRetentionHandler(retentionRepo, retentionService)
	qualityHandler := handlers.NewQualityHandler(qualityService, jobRepo)
	mlExportHandler := handlers.NewMLExportHandler(mlExportService)

	// Health routes
	api.Get("/health", healthHandler.GetHealth)
	api.Get("/exchanges", healthHandler.GetSupportedExchanges)
	api.Get("/exchanges/test", healthHandler.TestExchangeAvailability)
	api.Get("/exchanges/metadata", healthHandler.GetExchangesMetadata)
	api.Post("/exchanges/refresh", healthHandler.RefreshExchangeCache)
	api.Get("/exchanges/:id/metadata", healthHandler.GetExchangeMetadata)
	api.Get("/exchanges/:id/debug", healthHandler.DebugExchange)
	api.Get("/exchanges/:id/symbols", healthHandler.GetExchangeSymbols)
	api.Get("/exchanges/:id/symbols/validate", healthHandler.ValidateSymbol)
	api.Post("/exchanges/:id/symbols/validate", healthHandler.ValidateSymbols)
	api.Get("/exchanges/:id/symbols/popular", healthHandler.GetPopularSymbols)

	// Connector routes (health route MUST come before :id route)
	api.Post("/connectors", connectorHandler.CreateConnector)
	api.Get("/connectors", connectorHandler.GetConnectors)
	api.Get("/connectors/health", connectorHandler.GetAllConnectorsHealth)
	api.Get("/connectors/:id", connectorHandler.GetConnector)
	api.Put("/connectors/:id", connectorHandler.UpdateConnector)
	api.Delete("/connectors/:id", connectorHandler.DeleteConnector)
	api.Post("/connectors/:id/suspend", connectorHandler.SuspendConnector)
	api.Post("/connectors/:id/resume", connectorHandler.ResumeConnector)
	api.Get("/connectors/:id/rate-limit", connectorHandler.GetRateLimitStatus)
	api.Post("/connectors/:id/rate-limit/reset", connectorHandler.ResetRateLimitUsage)
	api.Get("/connectors/:id/stats", connectorHandler.GetConnectorStats)
	api.Get("/connectors/:id/health", connectorHandler.GetConnectorHealth)

	// Global stats
	api.Get("/stats", connectorHandler.GetAllStats)

	// Job routes (queue and batch routes MUST come before :id routes)
	api.Post("/jobs", jobHandler.CreateJob)
	api.Post("/jobs/batch", jobHandler.CreateJobsBatch)
	api.Get("/jobs", jobHandler.GetJobs)
	api.Get("/jobs/queue", jobHandler.GetQueue)
	api.Get("/jobs/:id", jobHandler.GetJob)
	api.Put("/jobs/:id", jobHandler.UpdateJob)
	api.Delete("/jobs/:id", jobHandler.DeleteJob)
	api.Post("/jobs/:id/pause", jobHandler.PauseJob)
	api.Post("/jobs/:id/resume", jobHandler.ResumeJob)
	api.Post("/jobs/:id/execute", jobHandler.ExecuteJob)

	// Job data export routes
	api.Get("/jobs/:id/ohlcv", jobHandler.GetJobOHLCVData)
	api.Get("/jobs/:id/export", jobHandler.ExportJobData)
	api.Get("/jobs/:id/export/ml", jobHandler.ExportJobDataForML)

	// Job dependency routes
	api.Get("/jobs/:id/dependencies", jobHandler.GetJobDependencies)
	api.Put("/jobs/:id/dependencies", jobHandler.SetJobDependencies)
	api.Get("/jobs/:id/dependents", jobHandler.GetJobDependents)

	// Data quality routes (cached results)
	api.Get("/jobs/:id/quality", qualityHandler.GetJobQuality)
	api.Post("/jobs/:id/quality/refresh", qualityHandler.RefreshJobQuality)
	api.Post("/jobs/:id/quality/fill-gaps", qualityHandler.FillJobGaps)
	api.Get("/jobs/:id/quality/fill-gaps/status", qualityHandler.GetGapFillStatus)
	api.Get("/jobs/:id/quality/fill-gaps/history", qualityHandler.GetGapFillHistory)
	api.Post("/jobs/:id/quality/backfill", qualityHandler.StartBackfill)
	api.Get("/jobs/:id/quality/backfill/status", qualityHandler.GetBackfillStatus)
	api.Get("/jobs/:id/quality/backfill/history", qualityHandler.GetBackfillHistory)
	api.Get("/quality", qualityHandler.GetCachedResults)
	api.Get("/quality/summary", qualityHandler.GetCachedSummary)
	api.Post("/quality/check", qualityHandler.StartQualityCheck)
	api.Get("/quality/checks", qualityHandler.GetRecentCheckJobs)
	api.Get("/quality/checks/active", qualityHandler.GetActiveCheckJobs)
	api.Get("/quality/checks/:id", qualityHandler.GetCheckJobStatus)

	// Connector-specific job routes
	api.Get("/connectors/:exchangeId/jobs", jobHandler.GetJobsByConnector)

	// Indicator data retrieval routes (using query parameter for symbol to handle slashes)
	api.Get("/indicators/:exchange/:timeframe/latest", indicatorHandler.GetLatestIndicators)
	api.Get("/indicators/:exchange/:timeframe/range", indicatorHandler.GetIndicatorRange)
	api.Get("/indicators/:exchange/:timeframe/:indicator", indicatorHandler.GetSpecificIndicator)

	// Indicator recalculation routes
	api.Post("/jobs/:id/indicators/recalculate", indicatorHandler.RecalculateJob)
	api.Post("/connectors/:id/indicators/recalculate", indicatorHandler.RecalculateConnector)

	// Indicator configuration routes (builtin-defaults, default, validation-rules, and validate MUST come before :id routes)
	api.Get("/indicators/configs", indicatorConfigHandler.GetConfigs)
	api.Get("/indicators/configs/builtin-defaults", indicatorConfigHandler.GetBuiltinDefaults)
	api.Get("/indicators/configs/default", indicatorConfigHandler.GetDefaultConfig)
	api.Get("/indicators/configs/validation-rules", indicatorConfigHandler.GetValidationRules)
	api.Post("/indicators/configs/validate", indicatorConfigHandler.ValidateConfig)
	api.Post("/indicators/configs", indicatorConfigHandler.CreateConfig)
	api.Get("/indicators/configs/:id", indicatorConfigHandler.GetConfig)
	api.Put("/indicators/configs/:id", indicatorConfigHandler.UpdateConfig)
	api.Delete("/indicators/configs/:id", indicatorConfigHandler.DeleteConfig)
	api.Post("/indicators/configs/:id/default", indicatorConfigHandler.SetDefaultConfig)

	// Alert routes (summary and config routes MUST come before :id routes)
	api.Get("/alerts", alertHandler.GetAlerts)
	api.Get("/alerts/active", alertHandler.GetActiveAlerts)
	api.Get("/alerts/summary", alertHandler.GetAlertSummary)
	api.Get("/alerts/config", alertHandler.GetAlertConfig)
	api.Put("/alerts/config", alertHandler.UpdateAlertConfig)
	api.Post("/alerts/acknowledge-all", alertHandler.AcknowledgeAllAlerts)
	api.Post("/alerts/check", alertHandler.TriggerAlertCheck)
	api.Post("/alerts/cleanup", alertHandler.CleanupAlerts)
	api.Get("/alerts/:id", alertHandler.GetAlert)
	api.Post("/alerts/:id/acknowledge", alertHandler.AcknowledgeAlert)
	api.Post("/alerts/:id/resolve", alertHandler.ResolveAlert)
	api.Delete("/alerts/:id", alertHandler.DeleteAlert)

	// Job and connector alerts
	api.Get("/jobs/:id/alerts", alertHandler.GetAlertsByJob)
	api.Get("/connectors/:exchangeId/alerts", alertHandler.GetAlertsByConnector)

	// Retention policy routes
	api.Get("/retention/policies", retentionHandler.GetPolicies)
	api.Post("/retention/policies", retentionHandler.CreatePolicy)
	api.Get("/retention/policies/:id", retentionHandler.GetPolicy)
	api.Put("/retention/policies/:id", retentionHandler.UpdatePolicy)
	api.Delete("/retention/policies/:id", retentionHandler.DeletePolicy)

	// Retention config and cleanup routes
	api.Get("/retention/config", retentionHandler.GetConfig)
	api.Put("/retention/config", retentionHandler.UpdateConfig)
	api.Get("/retention/usage", retentionHandler.GetDataUsage)
	api.Post("/retention/cleanup", retentionHandler.RunCleanup)
	api.Post("/retention/cleanup/default", retentionHandler.RunDefaultCleanup)
	api.Post("/retention/cleanup/empty", retentionHandler.DeleteEmptyChunks)
	api.Post("/retention/cleanup/exchange/:exchangeId", retentionHandler.CleanupExchange)

	// ML Export routes
	ml := api.Group("/ml")

	// ML Export job routes
	ml.Post("/export/start", mlExportHandler.StartExport)
	ml.Get("/export/jobs", mlExportHandler.ListExportJobs)
	ml.Get("/export/jobs/:id", mlExportHandler.GetExportJob)
	ml.Get("/export/jobs/:id/download", mlExportHandler.DownloadExport)
	ml.Get("/export/jobs/:id/metadata", mlExportHandler.GetExportMetadata)
	ml.Post("/export/jobs/:id/cancel", mlExportHandler.CancelExport)
	ml.Delete("/export/jobs/:id", mlExportHandler.DeleteExport)

	// ML Export profile routes (presets MUST come before :id routes)
	ml.Get("/profiles", mlExportHandler.ListProfiles)
	ml.Get("/profiles/presets", mlExportHandler.GetPresets)
	ml.Post("/profiles", mlExportHandler.CreateProfile)
	ml.Get("/profiles/:id", mlExportHandler.GetProfile)
	ml.Put("/profiles/:id", mlExportHandler.UpdateProfile)
	ml.Delete("/profiles/:id", mlExportHandler.DeleteProfile)

	// ML Export utility routes
	ml.Get("/formats", mlExportHandler.GetSupportedFormats)
	ml.Get("/features", mlExportHandler.GetAvailableFeatures)
	ml.Get("/config/default", mlExportHandler.GetDefaultConfig)

	// ML Dataset routes
	ml.Post("/datasets", mlExportHandler.CreateDataset)

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", address)

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
