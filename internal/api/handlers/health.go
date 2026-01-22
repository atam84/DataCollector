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

// ValidateSymbol validates if a symbol exists on an exchange
// GET /api/v1/exchanges/:id/symbols/validate?symbol=BTC/USDT
func (h *HealthHandler) ValidateSymbol(c *fiber.Ctx) error {
	exchangeID := c.Params("id")
	symbol := c.Query("symbol", "")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbol query parameter is required",
		})
	}

	symbols, err := exchange.GetExchangeSymbols(exchangeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"valid":   false,
			"symbol":  symbol,
		})
	}

	// Check if symbol exists
	isValid := false
	for _, s := range symbols {
		if s == symbol {
			isValid = true
			break
		}
	}

	// Find similar symbols if not valid
	var suggestions []string
	if !isValid {
		// Extract base currency from symbol (e.g., "BTC" from "BTC/USDT")
		parts := splitSymbol(symbol)
		if len(parts) >= 1 {
			base := parts[0]
			for _, s := range symbols {
				if containsIgnoreCase(s, base) {
					suggestions = append(suggestions, s)
					if len(suggestions) >= 10 {
						break
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"valid":       isValid,
		"suggestions": suggestions,
	})
}

// ValidateSymbols validates multiple symbols at once
// POST /api/v1/exchanges/:id/symbols/validate
func (h *HealthHandler) ValidateSymbols(c *fiber.Ctx) error {
	exchangeID := c.Params("id")

	var request struct {
		Symbols []string `json:"symbols"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(request.Symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbols array is required",
		})
	}

	symbols, err := exchange.GetExchangeSymbols(exchangeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a set for faster lookup
	symbolSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolSet[s] = true
	}

	// Validate each symbol
	results := make(map[string]bool)
	validCount := 0
	invalidCount := 0

	for _, sym := range request.Symbols {
		isValid := symbolSet[sym]
		results[sym] = isValid
		if isValid {
			validCount++
		} else {
			invalidCount++
		}
	}

	return c.JSON(fiber.Map{
		"exchange_id":   exchangeID,
		"results":       results,
		"valid_count":   validCount,
		"invalid_count": invalidCount,
		"total":         len(request.Symbols),
	})
}

// GetPopularSymbols returns popular trading pairs that are available on the exchange
// GET /api/v1/exchanges/:id/symbols/popular
func (h *HealthHandler) GetPopularSymbols(c *fiber.Ctx) error {
	exchangeID := c.Params("id")

	// Popular pairs in priority order
	popularPairs := []string{
		"BTC/USDT", "ETH/USDT", "BNB/USDT", "SOL/USDT", "XRP/USDT",
		"ADA/USDT", "DOGE/USDT", "DOT/USDT", "MATIC/USDT", "AVAX/USDT",
		"LINK/USDT", "UNI/USDT", "ATOM/USDT", "LTC/USDT", "ETC/USDT",
		"BCH/USDT", "APT/USDT", "ARB/USDT", "OP/USDT", "INJ/USDT",
		// Additional popular pairs
		"BTC/USD", "ETH/USD", "BTC/EUR", "ETH/EUR",
		"BTC/BUSD", "ETH/BUSD", "BNB/BUSD",
		"SHIB/USDT", "PEPE/USDT", "WIF/USDT", "BONK/USDT",
	}

	symbols, err := exchange.GetExchangeSymbols(exchangeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a set for faster lookup
	symbolSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolSet[s] = true
	}

	// Filter to only available pairs
	available := make([]string, 0)
	for _, pair := range popularPairs {
		if symbolSet[pair] {
			available = append(available, pair)
		}
	}

	return c.JSON(fiber.Map{
		"exchange_id": exchangeID,
		"popular":     available,
		"count":       len(available),
	})
}

// splitSymbol splits a trading pair symbol into base and quote currencies
func splitSymbol(symbol string) []string {
	var parts []string
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			parts = append(parts, symbol[:i])
			if i+1 < len(symbol) {
				parts = append(parts, symbol[i+1:])
			}
			return parts
		}
	}
	return []string{symbol}
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
