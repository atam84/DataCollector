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
		"available":         available,
		"unavailable":       unavailable,
		"available_count":   len(available),
		"unavailable_count": len(unavailable),
		"results":           results,
	})
}

// GetExchangesMetadata returns metadata for all supported exchanges
// GET /api/v1/exchanges/metadata
func (h *HealthHandler) GetExchangesMetadata(c *fiber.Ctx) error {
	metadata := exchange.GetAllExchangesMetadata()

	return c.JSON(fiber.Map{
		"exchanges": metadata,
		"count":     len(metadata),
	})
}

// GetExchangeMetadata returns metadata for a specific exchange
// GET /api/v1/exchanges/:id/metadata
func (h *HealthHandler) GetExchangeMetadata(c *fiber.Ctx) error {
	exchangeID := c.Params("id")

	metadata, err := exchange.GetExchangeMetadata(exchangeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(metadata)
}

// RefreshExchangeCache clears and refreshes the exchange metadata cache
// POST /api/v1/exchanges/refresh
func (h *HealthHandler) RefreshExchangeCache(c *fiber.Ctx) error {
	// Clear the cache
	exchange.ClearCache()

	// Trigger rediscovery
	exchanges := exchange.GetSupportedExchanges()

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Exchange cache refreshed",
		"exchanges": len(exchanges),
	})
}

// DebugExchange returns detailed debug info for a specific exchange
// GET /api/v1/exchanges/:id/debug
func (h *HealthHandler) DebugExchange(c *fiber.Ctx) error {
	exchangeID := c.Params("id")

	debugInfo := exchange.DebugExchange(exchangeID)

	return c.JSON(fiber.Map{
		"exchange_id": exchangeID,
		"debug":       debugInfo,
	})
}

// GetExchangeSymbols returns all available symbols for an exchange
// GET /api/v1/exchanges/:id/symbols
func (h *HealthHandler) GetExchangeSymbols(c *fiber.Ctx) error {
	exchangeID := c.Params("id")

	// Check for search query
	search := c.Query("search", "")

	symbols, err := exchange.GetExchangeSymbols(exchangeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Filter symbols if search is provided
	if search != "" {
		filtered := make([]string, 0)
		searchLower := c.Query("search", "")
		for _, s := range symbols {
			if containsIgnoreCase(s, searchLower) {
				filtered = append(filtered, s)
			}
		}
		symbols = filtered
	}

	return c.JSON(fiber.Map{
		"exchange_id": exchangeID,
		"symbols":     symbols,
		"count":       len(symbols),
	})
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			sLower[i] = s[i] + 32
		} else {
			sLower[i] = s[i]
		}
	}
	for i := 0; i < len(substr); i++ {
		if substr[i] >= 'A' && substr[i] <= 'Z' {
			substrLower[i] = substr[i] + 32
		} else {
			substrLower[i] = substr[i]
		}
	}
	return bytesContains(sLower, substrLower)
}

// bytesContains checks if b contains sub
func bytesContains(b, sub []byte) bool {
	if len(sub) == 0 {
		return true
	}
	if len(b) < len(sub) {
		return false
	}
	for i := 0; i <= len(b)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if b[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
