package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/api/errors"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

// RetentionHandler handles retention policy endpoints
type RetentionHandler struct {
	retentionRepo    *repository.RetentionRepository
	retentionService *service.RetentionService
}

// NewRetentionHandler creates a new retention handler
func NewRetentionHandler(retentionRepo *repository.RetentionRepository, retentionService *service.RetentionService) *RetentionHandler {
	return &RetentionHandler{
		retentionRepo:    retentionRepo,
		retentionService: retentionService,
	}
}

// GetPolicies retrieves all retention policies
// GET /api/v1/retention/policies
func (h *RetentionHandler) GetPolicies(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	policies, err := h.retentionRepo.FindAllPolicies(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve retention policies"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    policies,
		"count":   len(policies),
	})
}

// GetPolicy retrieves a single retention policy
// GET /api/v1/retention/policies/:id
func (h *RetentionHandler) GetPolicy(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	policy, err := h.retentionRepo.FindPolicyByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid policy ID format"))
		}
		return errors.SendError(c, errors.NotFound("Retention policy"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    policy,
	})
}

// CreatePolicy creates a new retention policy
// POST /api/v1/retention/policies
func (h *RetentionHandler) CreatePolicy(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.RetentionPolicyCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate required fields
	if req.Name == "" || req.RetentionDays < 1 {
		return errors.SendError(c, errors.ValidationError("Invalid policy configuration", map[string]string{
			"name":           "required",
			"retention_days": "must be at least 1",
		}))
	}

	// Validate type
	if req.Type != models.RetentionPolicyTypeGlobal &&
		req.Type != models.RetentionPolicyTypeExchange &&
		req.Type != models.RetentionPolicyTypeTimeframe {
		return errors.SendError(c, errors.ValidationError("Invalid policy type", map[string]interface{}{
			"type":           req.Type,
			"allowed_values": []string{"global", "exchange", "timeframe"},
		}))
	}

	policy := &models.RetentionPolicy{
		Name:           req.Name,
		Type:           req.Type,
		Enabled:        req.Enabled,
		ExchangeID:     req.ExchangeID,
		Timeframe:      req.Timeframe,
		RetentionDays:  req.RetentionDays,
		MaxCandles:     req.MaxCandles,
		KeepLatestOnly: req.KeepLatestOnly,
		RunSchedule:    req.RunSchedule,
	}

	if err := h.retentionRepo.CreatePolicy(ctx, policy); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to create retention policy"))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    policy,
	})
}

// UpdatePolicy updates a retention policy
// PUT /api/v1/retention/policies/:id
func (h *RetentionHandler) UpdatePolicy(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req models.RetentionPolicyUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	update := bson.M{}

	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Enabled != nil {
		update["enabled"] = *req.Enabled
	}
	if req.RetentionDays != nil {
		if *req.RetentionDays < 1 {
			return errors.SendError(c, errors.ValidationError("Invalid retention days", map[string]string{
				"retention_days": "must be at least 1",
			}))
		}
		update["retention_days"] = *req.RetentionDays
	}
	if req.MaxCandles != nil {
		update["max_candles"] = *req.MaxCandles
	}
	if req.KeepLatestOnly != nil {
		update["keep_latest_only"] = *req.KeepLatestOnly
	}
	if req.RunSchedule != nil {
		update["run_schedule"] = *req.RunSchedule
	}

	if len(update) == 0 {
		return errors.SendError(c, errors.BadRequest("No fields to update"))
	}

	if err := h.retentionRepo.UpdatePolicy(ctx, id, update); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid policy ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Retention policy"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to update retention policy"))
	}

	policy, _ := h.retentionRepo.FindPolicyByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    policy,
	})
}

// DeletePolicy deletes a retention policy
// DELETE /api/v1/retention/policies/:id
func (h *RetentionHandler) DeletePolicy(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.retentionRepo.DeletePolicy(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid policy ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Retention policy"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to delete retention policy"))
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetConfig retrieves the retention configuration
// GET /api/v1/retention/config
func (h *RetentionHandler) GetConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := h.retentionRepo.GetConfig(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve retention configuration"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// UpdateConfig updates the retention configuration
// PUT /api/v1/retention/config
func (h *RetentionHandler) UpdateConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config models.RetentionConfig
	if err := c.BodyParser(&config); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate
	if config.DefaultRetentionDays < 1 {
		return errors.SendError(c, errors.ValidationError("Invalid configuration", map[string]string{
			"default_retention_days": "must be at least 1",
		}))
	}

	if err := h.retentionRepo.SaveConfig(ctx, &config); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to save retention configuration"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Retention configuration updated",
		"data":    config,
	})
}

// RunCleanup manually triggers a cleanup operation
// POST /api/v1/retention/cleanup
func (h *RetentionHandler) RunCleanup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	summary, err := h.retentionService.RunCleanup(ctx)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Cleanup failed: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Cleanup completed",
		"data":    summary,
	})
}

// RunDefaultCleanup runs cleanup with default retention days
// POST /api/v1/retention/cleanup/default
func (h *RetentionHandler) RunDefaultCleanup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	days := c.QueryInt("days", 365)
	if days < 1 {
		return errors.SendError(c, errors.ValidationError("Invalid retention days", map[string]string{
			"days": "must be at least 1",
		}))
	}

	result, err := h.retentionService.RunDefaultCleanup(ctx, days)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Cleanup failed: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Default cleanup completed",
		"data":    result,
	})
}

// CleanupExchange runs cleanup for a specific exchange
// POST /api/v1/retention/cleanup/exchange/:exchangeId
func (h *RetentionHandler) CleanupExchange(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	exchangeID := c.Params("exchangeId")
	days := c.QueryInt("days", 365)

	if days < 1 {
		return errors.SendError(c, errors.ValidationError("Invalid retention days", map[string]string{
			"days": "must be at least 1",
		}))
	}

	result, err := h.retentionService.CleanupByExchange(ctx, exchangeID, days)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Cleanup failed: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Exchange cleanup completed",
		"data":    result,
	})
}

// GetDataUsage returns data usage statistics
// GET /api/v1/retention/usage
func (h *RetentionHandler) GetDataUsage(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exchangeID := c.Query("exchange_id", "")

	stats, err := h.retentionService.GetDataUsage(ctx, exchangeID)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve data usage"))
	}

	// Get totals
	chunks, candles, sizeMB, _ := h.retentionService.GetTotalUsage(ctx)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
		"summary": fiber.Map{
			"total_chunks":       chunks,
			"total_candles":      candles,
			"estimated_size_mb":  sizeMB,
		},
	})
}

// DeleteEmptyChunks removes chunks with no candles
// POST /api/v1/retention/cleanup/empty
func (h *RetentionHandler) DeleteEmptyChunks(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deleted, err := h.retentionService.DeleteEmptyChunks(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to delete empty chunks"))
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"message":        "Empty chunks deleted",
		"chunks_deleted": deleted,
	})
}
