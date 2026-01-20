package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// ConnectorHandler handles connector-related endpoints
type ConnectorHandler struct {
	repo   *repository.ConnectorRepository
	config *config.Config
}

// NewConnectorHandler creates a new connector handler
func NewConnectorHandler(repo *repository.ConnectorRepository, cfg *config.Config) *ConnectorHandler {
	return &ConnectorHandler{
		repo:   repo,
		config: cfg,
	}
}

// CreateConnector creates a new connector
// POST /api/v1/connectors
func (h *ConnectorHandler) CreateConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.ConnectorCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ExchangeID == "" || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exchange_id and display_name are required",
		})
	}

	if req.RateLimit.Limit <= 0 || req.RateLimit.PeriodMs < 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Valid rate limit configuration is required",
		})
	}

	// Create connector model
	connector := &models.Connector{
		ExchangeID:  req.ExchangeID,
		DisplayName: req.DisplayName,
		Status:      "active",
		SandboxMode: req.SandboxMode,
		RateLimit: models.RateLimit{
			Limit:       req.RateLimit.Limit,
			PeriodMs:    req.RateLimit.PeriodMs,
			Usage:       0,
			PeriodStart: time.Now(),
		},
	}

	// Create in database
	if err := h.repo.Create(ctx, connector); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(connector)
}

// GetConnectors retrieves all connectors
// GET /api/v1/connectors
func (h *ConnectorHandler) GetConnectors(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optional filter by status
	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Optional filter by sandbox mode
	if sandboxMode := c.Query("sandbox_mode"); sandboxMode != "" {
		filter["sandbox_mode"] = sandboxMode == "true"
	}

	connectors, err := h.repo.FindAll(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  connectors,
		"count": len(connectors),
	})
}

// GetConnector retrieves a connector by ID
// GET /api/v1/connectors/:id
func (h *ConnectorHandler) GetConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(connector)
}

// UpdateConnector updates a connector
// PUT /api/v1/connectors/:id
func (h *ConnectorHandler) UpdateConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req models.ConnectorUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Build update map
	update := bson.M{}

	if req.DisplayName != nil {
		update["display_name"] = *req.DisplayName
	}

	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "disabled" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "status must be 'active' or 'disabled'",
			})
		}
		update["status"] = *req.Status
	}

	if req.SandboxMode != nil {
		update["sandbox_mode"] = *req.SandboxMode
	}

	if req.RateLimit != nil {
		if req.RateLimit.Limit != nil {
			update["rate_limit.limit"] = *req.RateLimit.Limit
		}
		if req.RateLimit.PeriodMs != nil {
			update["rate_limit.period_ms"] = *req.RateLimit.PeriodMs
		}
	}

	if len(update) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	// Update connector
	if err := h.repo.Update(ctx, id, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch updated connector
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(connector)
}

// DeleteConnector deletes a connector
// DELETE /api/v1/connectors/:id
func (h *ConnectorHandler) DeleteConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.repo.Delete(ctx, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ToggleSandboxMode toggles sandbox mode for a connector
// PATCH /api/v1/connectors/:id/sandbox
func (h *ConnectorHandler) ToggleSandboxMode(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req struct {
		SandboxMode bool `json:"sandbox_mode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.repo.UpdateSandboxMode(ctx, id, req.SandboxMode); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch updated connector
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Sandbox mode updated successfully",
		"sandbox_mode": connector.SandboxMode,
		"connector":    connector,
	})
}
