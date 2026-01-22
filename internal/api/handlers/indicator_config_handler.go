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
)

// IndicatorConfigHandler handles indicator configuration endpoints
type IndicatorConfigHandler struct {
	configRepo *repository.IndicatorConfigRepository
}

// NewIndicatorConfigHandler creates a new indicator config handler
func NewIndicatorConfigHandler(configRepo *repository.IndicatorConfigRepository) *IndicatorConfigHandler {
	return &IndicatorConfigHandler{
		configRepo: configRepo,
	}
}

// GetConfigs retrieves all indicator configurations
// GET /api/v1/indicators/configs
func (h *IndicatorConfigHandler) GetConfigs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs, err := h.configRepo.FindAll(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve indicator configs"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    configs,
		"count":   len(configs),
	})
}

// GetConfig retrieves a single indicator configuration by ID
// GET /api/v1/indicators/configs/:id
func (h *IndicatorConfigHandler) GetConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	config, err := h.configRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid config ID format"))
		}
		return errors.SendError(c, errors.NotFound("Indicator config"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// GetDefaultConfig retrieves the default indicator configuration
// GET /api/v1/indicators/configs/default
func (h *IndicatorConfigHandler) GetDefaultConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := h.configRepo.FindDefault(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve default config"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// CreateConfig creates a new indicator configuration
// POST /api/v1/indicators/configs
func (h *IndicatorConfigHandler) CreateConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.IndicatorConfigCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	if req.Name == "" {
		return errors.SendError(c, errors.ValidationError("Name is required", nil))
	}

	// Create config with defaults
	config := &models.IndicatorConfig{
		Name:             req.Name,
		IsDefault:        req.IsDefault,
		EnableTrend:      req.EnableTrend,
		EnableMomentum:   req.EnableMomentum,
		EnableVolatility: req.EnableVolatility,
		EnableVolume:     req.EnableVolume,
		Trend:            models.DefaultTrendConfig(),
		Momentum:         models.DefaultMomentumConfig(),
		Volatility:       models.DefaultVolatilityConfig(),
		Volume:           models.DefaultVolumeConfig(),
	}

	// Override with provided values
	if req.Trend != nil {
		config.Trend = *req.Trend
	}
	if req.Momentum != nil {
		config.Momentum = *req.Momentum
	}
	if req.Volatility != nil {
		config.Volatility = *req.Volatility
	}
	if req.Volume != nil {
		config.Volume = *req.Volume
	}

	// Validate the configuration
	validationResult := config.Validate()
	if !validationResult.Valid {
		return errors.SendError(c, errors.ValidationError("Invalid indicator configuration", validationResult.Errors))
	}

	if err := h.configRepo.Create(ctx, config); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return errors.SendError(c, errors.Conflict("Config with this name already exists"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to create indicator config"))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// UpdateConfig updates an indicator configuration
// PUT /api/v1/indicators/configs/:id
func (h *IndicatorConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// First, get the existing config
	existingConfig, err := h.configRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid config ID format"))
		}
		return errors.SendError(c, errors.NotFound("Indicator config"))
	}

	var req models.IndicatorConfigUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	update := bson.M{}

	// Apply updates to a copy for validation
	configCopy := *existingConfig

	if req.Name != nil {
		update["name"] = *req.Name
		configCopy.Name = *req.Name
	}
	if req.IsDefault != nil {
		update["is_default"] = *req.IsDefault
		configCopy.IsDefault = *req.IsDefault
	}
	if req.EnableTrend != nil {
		update["enable_trend"] = *req.EnableTrend
		configCopy.EnableTrend = *req.EnableTrend
	}
	if req.EnableMomentum != nil {
		update["enable_momentum"] = *req.EnableMomentum
		configCopy.EnableMomentum = *req.EnableMomentum
	}
	if req.EnableVolatility != nil {
		update["enable_volatility"] = *req.EnableVolatility
		configCopy.EnableVolatility = *req.EnableVolatility
	}
	if req.EnableVolume != nil {
		update["enable_volume"] = *req.EnableVolume
		configCopy.EnableVolume = *req.EnableVolume
	}
	if req.Trend != nil {
		update["trend"] = *req.Trend
		configCopy.Trend = *req.Trend
	}
	if req.Momentum != nil {
		update["momentum"] = *req.Momentum
		configCopy.Momentum = *req.Momentum
	}
	if req.Volatility != nil {
		update["volatility"] = *req.Volatility
		configCopy.Volatility = *req.Volatility
	}
	if req.Volume != nil {
		update["volume"] = *req.Volume
		configCopy.Volume = *req.Volume
	}

	if len(update) == 0 {
		return errors.SendError(c, errors.BadRequest("No fields to update"))
	}

	// Validate the resulting configuration
	validationResult := configCopy.Validate()
	if !validationResult.Valid {
		return errors.SendError(c, errors.ValidationError("Invalid indicator configuration", validationResult.Errors))
	}

	if err := h.configRepo.Update(ctx, id, update); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid config ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Indicator config"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to update indicator config"))
	}

	config, _ := h.configRepo.FindByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// DeleteConfig deletes an indicator configuration
// DELETE /api/v1/indicators/configs/:id
func (h *IndicatorConfigHandler) DeleteConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.configRepo.Delete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid config ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Indicator config"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to delete indicator config"))
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// SetDefaultConfig sets a configuration as the default
// POST /api/v1/indicators/configs/:id/default
func (h *IndicatorConfigHandler) SetDefaultConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.configRepo.SetDefault(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid config ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Indicator config"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to set default config"))
	}

	config, _ := h.configRepo.FindByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Config set as default",
		"data":    config,
	})
}

// GetBuiltinDefaults returns the built-in default configuration values
// GET /api/v1/indicators/configs/builtin-defaults
func (h *IndicatorConfigHandler) GetBuiltinDefaults(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    models.DefaultIndicatorConfig(),
	})
}

// GetValidationRules returns the validation constraints for indicator configurations
// GET /api/v1/indicators/configs/validation-rules
func (h *IndicatorConfigHandler) GetValidationRules(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"period": fiber.Map{
				"min": models.MinPeriod,
				"max": models.MaxPeriod,
			},
			"multiplier": fiber.Map{
				"min": models.MinMultiplier,
				"max": models.MaxMultiplier,
			},
			"stddev": fiber.Map{
				"min": models.MinStdDev,
				"max": models.MaxStdDev,
			},
			"periods_list": fiber.Map{
				"max_items": models.MaxPeriodsInList,
			},
			"constraints": fiber.Map{
				"macd_fast_less_than_slow":       true,
				"ichimoku_tenkan_less_than_kijun": true,
			},
		},
	})
}

// ValidateConfig validates a configuration without saving it
// POST /api/v1/indicators/configs/validate
func (h *IndicatorConfigHandler) ValidateConfig(c *fiber.Ctx) error {
	var req models.IndicatorConfigCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Create config with defaults
	config := &models.IndicatorConfig{
		Name:             req.Name,
		IsDefault:        req.IsDefault,
		EnableTrend:      req.EnableTrend,
		EnableMomentum:   req.EnableMomentum,
		EnableVolatility: req.EnableVolatility,
		EnableVolume:     req.EnableVolume,
		Trend:            models.DefaultTrendConfig(),
		Momentum:         models.DefaultMomentumConfig(),
		Volatility:       models.DefaultVolatilityConfig(),
		Volume:           models.DefaultVolumeConfig(),
	}

	// Override with provided values
	if req.Trend != nil {
		config.Trend = *req.Trend
	}
	if req.Momentum != nil {
		config.Momentum = *req.Momentum
	}
	if req.Volatility != nil {
		config.Volatility = *req.Volatility
	}
	if req.Volume != nil {
		config.Volume = *req.Volume
	}

	validationResult := config.Validate()

	return c.JSON(fiber.Map{
		"success": true,
		"data":    validationResult,
	})
}
