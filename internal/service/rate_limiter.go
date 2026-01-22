package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// RateLimiter manages API call rate limiting per connector/exchange
type RateLimiter struct {
	connectorRepo *repository.ConnectorRepository
	mu            sync.Mutex
	// In-memory cache of last API call times for faster access
	lastCallTimes map[string]time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(connectorRepo *repository.ConnectorRepository) *RateLimiter {
	return &RateLimiter{
		connectorRepo: connectorRepo,
		lastCallTimes: make(map[string]time.Time),
	}
}

// WaitForSlot waits until it's safe to make an API call, then acquires a slot
// This should be called BEFORE every API call to the exchange
func (r *RateLimiter) WaitForSlot(ctx context.Context, exchangeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get connector configuration
	connector, err := r.connectorRepo.FindByExchangeID(ctx, exchangeID)
	if err != nil {
		return fmt.Errorf("failed to find connector for rate limiting: %w", err)
	}

	// Calculate minimum delay
	minDelayMs := r.getMinDelay(connector)

	// Check time since last API call
	lastCall, exists := r.lastCallTimes[exchangeID]
	if exists {
		elapsed := time.Since(lastCall)
		minDelay := time.Duration(minDelayMs) * time.Millisecond

		if elapsed < minDelay {
			waitTime := minDelay - elapsed
			log.Printf("[RATE_LIMIT] %s: Waiting %v before next API call (min delay: %dms)",
				exchangeID, waitTime.Round(time.Millisecond), minDelayMs)

			// Wait with context cancellation support
			select {
			case <-time.After(waitTime):
				// Continue after waiting
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Check period-based rate limit
	now := time.Now()
	periodElapsed := now.Sub(connector.RateLimit.PeriodStart).Milliseconds()

	if periodElapsed >= int64(connector.RateLimit.PeriodMs) {
		// Period has elapsed, reset
		log.Printf("[RATE_LIMIT] %s: Period elapsed, resetting usage counter", exchangeID)
		if err := r.resetPeriod(ctx, exchangeID); err != nil {
			log.Printf("[RATE_LIMIT] Warning: Failed to reset period: %v", err)
		}
	} else if connector.RateLimit.Usage >= connector.RateLimit.Limit {
		// Limit reached, wait for period to reset
		remainingMs := int64(connector.RateLimit.PeriodMs) - periodElapsed
		waitTime := time.Duration(remainingMs) * time.Millisecond

		log.Printf("[RATE_LIMIT] %s: Rate limit reached (%d/%d), waiting %v for period reset",
			exchangeID, connector.RateLimit.Usage, connector.RateLimit.Limit, waitTime.Round(time.Millisecond))

		select {
		case <-time.After(waitTime):
			// Reset after waiting
			if err := r.resetPeriod(ctx, exchangeID); err != nil {
				log.Printf("[RATE_LIMIT] Warning: Failed to reset period after wait: %v", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Record this API call
	r.lastCallTimes[exchangeID] = time.Now()

	// Increment usage counter in database
	if err := r.incrementUsage(ctx, exchangeID); err != nil {
		log.Printf("[RATE_LIMIT] Warning: Failed to increment usage: %v", err)
	}

	log.Printf("[RATE_LIMIT] %s: API call slot acquired", exchangeID)
	return nil
}

// getMinDelay returns the minimum delay between API calls in milliseconds
func (r *RateLimiter) getMinDelay(connector *models.Connector) int {
	// If MinDelayMs is explicitly set, use it
	if connector.RateLimit.MinDelayMs > 0 {
		return connector.RateLimit.MinDelayMs
	}

	// Otherwise calculate from limit/period
	// Example: 20 calls per 60000ms = 60000/20 = 3000ms between calls
	if connector.RateLimit.Limit > 0 && connector.RateLimit.PeriodMs > 0 {
		calculated := connector.RateLimit.PeriodMs / connector.RateLimit.Limit
		// Ensure minimum of 1000ms (1 second) for safety
		if calculated < 1000 {
			return 1000
		}
		return calculated
	}

	// Default to 5 seconds if nothing is configured
	return 5000
}

// resetPeriod resets the rate limit period
func (r *RateLimiter) resetPeriod(ctx context.Context, exchangeID string) error {
	return r.connectorRepo.ResetRateLimitPeriod(ctx, exchangeID)
}

// incrementUsage increments the API call counter
func (r *RateLimiter) incrementUsage(ctx context.Context, exchangeID string) error {
	return r.connectorRepo.IncrementAPIUsage(ctx, exchangeID)
}

// GetRateLimitStatus returns current rate limit status for an exchange
func (r *RateLimiter) GetRateLimitStatus(ctx context.Context, exchangeID string) (*RateLimitStatus, error) {
	connector, err := r.connectorRepo.FindByExchangeID(ctx, exchangeID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	periodElapsed := now.Sub(connector.RateLimit.PeriodStart).Milliseconds()
	periodRemaining := int64(connector.RateLimit.PeriodMs) - periodElapsed
	if periodRemaining < 0 {
		periodRemaining = 0
	}

	minDelayMs := r.getMinDelay(connector)

	var timeSinceLastCall int64
	if lastCall, exists := r.lastCallTimes[exchangeID]; exists {
		timeSinceLastCall = time.Since(lastCall).Milliseconds()
	}

	return &RateLimitStatus{
		ExchangeID:        exchangeID,
		Limit:             connector.RateLimit.Limit,
		Usage:             connector.RateLimit.Usage,
		PeriodMs:          connector.RateLimit.PeriodMs,
		PeriodRemainingMs: periodRemaining,
		MinDelayMs:        minDelayMs,
		LastCallMs:        timeSinceLastCall,
		CanCallNow:        timeSinceLastCall >= int64(minDelayMs) && connector.RateLimit.Usage < connector.RateLimit.Limit,
	}, nil
}

// RateLimitStatus represents the current rate limiting state
type RateLimitStatus struct {
	ExchangeID        string `json:"exchange_id"`
	Limit             int    `json:"limit"`               // Max calls per period
	Usage             int    `json:"usage"`               // Current usage in period
	PeriodMs          int    `json:"period_ms"`           // Period duration
	PeriodRemainingMs int64  `json:"period_remaining_ms"` // Time until period resets
	MinDelayMs        int    `json:"min_delay_ms"`        // Min delay between calls
	LastCallMs        int64  `json:"last_call_ms"`        // Time since last call
	CanCallNow        bool   `json:"can_call_now"`        // Whether an API call is allowed now
}
