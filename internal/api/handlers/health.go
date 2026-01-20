package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/datacollector/internal/exchange"
	"github.com/yourusername/datacollector/internal/repository"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *repository.Database
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *repository.Database) *HealthHandler {
	return &HealthHandler{db: db}
}

// GetHealth returns the health status of the application
func (h *HealthHandler) GetHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check database health
	dbStatus := "healthy"
	dbError := ""
	if err := h.db.HealthCheck(ctx); err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
	}

	response := fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"services": fiber.Map{
			"database": fiber.Map{
				"status": dbStatus,
				"error":  dbError,
			},
		},
	}

	if dbStatus == "unhealthy" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}

// GetSupportedExchanges returns the list of exchanges supported by CCXT
func (h *HealthHandler) GetSupportedExchanges(c *fiber.Ctx) error {
	exchanges := exchange.GetSupportedExchanges()

	return c.JSON(fiber.Map{
		"exchanges": exchanges,
		"count":     len(exchanges),
	})
}

// TestExchangeAvailability tests which exchanges can be instantiated
// This is useful for discovering which exchanges are available in the current CCXT build
func (h *HealthHandler) TestExchangeAvailability(c *fiber.Ctx) error {
	results := exchange.TestExchangeAvailability()

	available := []string{}
	unavailable := []string{}

	for exchangeID, isAvailable := range results {
		if isAvailable {
			available = append(available, exchangeID)
		} else {
			unavailable = append(unavailable, exchangeID)
		}
	}

	return c.JSON(fiber.Map{
		"available":        available,
		"unavailable":      unavailable,
		"available_count":  len(available),
		"unavailable_count": len(unavailable),
		"results":          results,
	})
}
