package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/yourusername/datacollector/internal/api/errors"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/service"
)

// MLExportHandler handles ML export endpoints
type MLExportHandler struct {
	exportService *service.MLExportService
}

// NewMLExportHandler creates a new ML export handler
func NewMLExportHandler(exportService *service.MLExportService) *MLExportHandler {
	return &MLExportHandler{
		exportService: exportService,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// StartExportRequest is the request body for starting an export
type StartExportRequest struct {
	JobIDs []string             `json:"job_ids"`
	Config models.MLExportConfig `json:"config"`
}

// ExportResponse is the response for export operations
type ExportResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	Progress        float64   `json:"progress"`
	CurrentPhase    string    `json:"current_phase,omitempty"`
	TotalRecords    int64     `json:"total_records,omitempty"`
	ProcessedRecords int64    `json:"processed_records,omitempty"`
	OutputPath      string    `json:"output_path,omitempty"`
	FileSizeBytes   int64     `json:"file_size_bytes,omitempty"`
	FeatureCount    int       `json:"feature_count,omitempty"`
	RowCount        int64     `json:"row_count,omitempty"`
	ColumnNames     []string  `json:"column_names,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	LastError       string    `json:"last_error,omitempty"`
	DownloadURL     string    `json:"download_url,omitempty"`
}

// ConfigResponse wraps export config for API responses
type ConfigResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	IsPreset    bool                 `json:"is_preset"`
	Format      string               `json:"format"`
	Config      models.MLExportConfig `json:"config"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// ============================================================================
// Export Job Endpoints
// ============================================================================

// StartExport starts a background ML export job
// @Summary Start ML export
// @Description Starts a background job to export data with feature engineering
// @Tags ML Export
// @Accept json
// @Produce json
// @Param request body StartExportRequest true "Export configuration"
// @Success 202 {object} ExportResponse "Export job started"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /ml/export/start [post]
func (h *MLExportHandler) StartExport(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req StartExportRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate job IDs
	if len(req.JobIDs) == 0 {
		return errors.SendError(c, errors.ValidationError("No job IDs provided", map[string]string{
			"job_ids": "at least one job ID is required",
		}))
	}

	// Convert string IDs to ObjectIDs
	jobIDs := make([]primitive.ObjectID, len(req.JobIDs))
	for i, idStr := range req.JobIDs {
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return errors.SendError(c, errors.ValidationError("Invalid job ID", map[string]string{
				"job_id": idStr,
			}))
		}
		jobIDs[i] = objID
	}

	// Set defaults if not provided
	if req.Config.Format == "" {
		req.Config.Format = models.MLExportFormatCSV
	}

	// Start export
	exportJob, err := h.exportService.StartExport(ctx, req.Config, jobIDs)
	if err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	response := ExportResponse{
		ID:           exportJob.ID.Hex(),
		Status:       string(exportJob.Status),
		Progress:     exportJob.Progress,
		CurrentPhase: exportJob.CurrentPhase,
		DownloadURL:  fmt.Sprintf("/api/v1/ml/export/jobs/%s/download", exportJob.ID.Hex()),
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Export job started",
	})
}

// GetExportJob gets the status of an export job
// @Summary Get export job status
// @Description Returns the current status and progress of an export job
// @Tags ML Export
// @Produce json
// @Param id path string true "Export job ID"
// @Success 200 {object} ExportResponse "Export job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /ml/export/jobs/{id} [get]
func (h *MLExportHandler) GetExportJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing job ID"))
	}

	exportJob, err := h.exportService.GetExportJob(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Export job"))
	}

	response := ExportResponse{
		ID:               exportJob.ID.Hex(),
		Status:           string(exportJob.Status),
		Progress:         exportJob.Progress,
		CurrentPhase:     exportJob.CurrentPhase,
		TotalRecords:     exportJob.TotalRecords,
		ProcessedRecords: exportJob.ProcessedRecords,
		OutputPath:       exportJob.OutputPath,
		FileSizeBytes:    exportJob.FileSizeBytes,
		FeatureCount:     exportJob.FeatureCount,
		RowCount:         exportJob.RowCount,
		ColumnNames:      exportJob.ColumnNames,
		StartedAt:        exportJob.StartedAt,
		CompletedAt:      exportJob.CompletedAt,
		LastError:        exportJob.LastError,
	}

	if exportJob.Status == models.MLExportStatusCompleted {
		response.DownloadURL = fmt.Sprintf("/api/v1/ml/export/jobs/%s/download", exportJob.ID.Hex())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// ListExportJobs lists all export jobs
// @Summary List export jobs
// @Description Returns a paginated list of export jobs
// @Tags ML Export
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of export jobs"
// @Router /ml/export/jobs [get]
func (h *MLExportHandler) ListExportJobs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	limit := int64(c.QueryInt("limit", 20))
	offset := int64(c.QueryInt("offset", 0))

	jobs, total, err := h.exportService.ListExportJobs(ctx, limit, offset)
	if err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	responses := make([]ExportResponse, len(jobs))
	for i, job := range jobs {
		responses[i] = ExportResponse{
			ID:               job.ID.Hex(),
			Status:           string(job.Status),
			Progress:         job.Progress,
			CurrentPhase:     job.CurrentPhase,
			TotalRecords:     job.TotalRecords,
			ProcessedRecords: job.ProcessedRecords,
			FileSizeBytes:    job.FileSizeBytes,
			FeatureCount:     job.FeatureCount,
			RowCount:         job.RowCount,
			StartedAt:        job.StartedAt,
			CompletedAt:      job.CompletedAt,
			LastError:        job.LastError,
		}
		if job.Status == models.MLExportStatusCompleted {
			responses[i].DownloadURL = fmt.Sprintf("/api/v1/ml/export/jobs/%s/download", job.ID.Hex())
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// DownloadExport downloads the exported file
// @Summary Download export file
// @Description Downloads the completed export file
// @Tags ML Export
// @Produce application/octet-stream
// @Param id path string true "Export job ID"
// @Success 200 {file} binary "Export file"
// @Failure 404 {object} map[string]interface{} "Job not found or not completed"
// @Router /ml/export/jobs/{id}/download [get]
func (h *MLExportHandler) DownloadExport(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing job ID"))
	}

	exportJob, err := h.exportService.GetExportJob(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Export job"))
	}

	if exportJob.Status != models.MLExportStatusCompleted {
		return errors.SendError(c, errors.BadRequest("Export job is not completed"))
	}

	if exportJob.OutputPath == "" {
		return errors.SendError(c, errors.NotFound("Export file"))
	}

	// Check if file exists
	if _, err := os.Stat(exportJob.OutputPath); os.IsNotExist(err) {
		return errors.SendError(c, errors.NotFound("Export file has expired or been deleted"))
	}

	// Set headers for download
	filename := filepath.Base(exportJob.OutputPath)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Determine content type
	ext := filepath.Ext(exportJob.OutputPath)
	contentType := "application/octet-stream"
	switch ext {
	case ".csv":
		contentType = "text/csv"
	case ".gz":
		contentType = "application/gzip"
	case ".parquet":
		contentType = "application/vnd.apache.parquet"
	case ".npz", ".npy":
		contentType = "application/x-numpy"
	case ".jsonl":
		contentType = "application/x-jsonlines"
	}
	c.Set("Content-Type", contentType)

	return c.SendFile(exportJob.OutputPath)
}

// CancelExport cancels a running export job
// @Summary Cancel export job
// @Description Cancels a running export job
// @Tags ML Export
// @Param id path string true "Export job ID"
// @Success 200 {object} map[string]interface{} "Job cancelled"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /ml/export/jobs/{id}/cancel [post]
func (h *MLExportHandler) CancelExport(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing job ID"))
	}

	if err := h.exportService.CancelExportJob(ctx, id); err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Export job cancelled",
	})
}

// DeleteExport deletes an export job and its files
// @Summary Delete export job
// @Description Deletes an export job and its output files
// @Tags ML Export
// @Param id path string true "Export job ID"
// @Success 200 {object} map[string]interface{} "Job deleted"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /ml/export/jobs/{id} [delete]
func (h *MLExportHandler) DeleteExport(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing job ID"))
	}

	if err := h.exportService.DeleteExportJob(ctx, id); err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Export job deleted",
	})
}

// ============================================================================
// Profile/Config Endpoints
// ============================================================================

// GetPresets returns built-in export presets
// @Summary Get export presets
// @Description Returns a list of built-in export configuration presets
// @Tags ML Export
// @Produce json
// @Success 200 {object} map[string]interface{} "List of presets"
// @Router /ml/profiles/presets [get]
func (h *MLExportHandler) GetPresets(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	presets, err := h.exportService.GetPresets(ctx)
	if err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	responses := make([]fiber.Map, len(presets))
	for i, preset := range presets {
		responses[i] = fiber.Map{
			"name":        preset.Name,
			"description": preset.Description,
			"format":      preset.Format,
			"config":      preset,
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
	})
}

// ListProfiles lists saved export profiles
// @Summary List export profiles
// @Description Returns a list of saved export configuration profiles
// @Tags ML Export
// @Produce json
// @Success 200 {object} map[string]interface{} "List of profiles"
// @Router /ml/profiles [get]
func (h *MLExportHandler) ListProfiles(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs, err := h.exportService.ListConfigs(ctx)
	if err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	responses := make([]ConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = ConfigResponse{
			ID:          config.ID.Hex(),
			Name:        config.Name,
			Description: config.Description,
			IsPreset:    config.IsPreset,
			Format:      string(config.Format),
			Config:      *config,
			CreatedAt:   config.CreatedAt,
			UpdatedAt:   config.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
	})
}

// GetProfile gets a specific export profile
// @Summary Get export profile
// @Description Returns a specific export configuration profile
// @Tags ML Export
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} ConfigResponse "Profile details"
// @Failure 404 {object} map[string]interface{} "Profile not found"
// @Router /ml/profiles/{id} [get]
func (h *MLExportHandler) GetProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing profile ID"))
	}

	config, err := h.exportService.GetConfig(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Profile"))
	}

	response := ConfigResponse{
		ID:          config.ID.Hex(),
		Name:        config.Name,
		Description: config.Description,
		IsPreset:    config.IsPreset,
		Format:      string(config.Format),
		Config:      *config,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// CreateProfile creates a new export profile
// @Summary Create export profile
// @Description Creates a new export configuration profile
// @Tags ML Export
// @Accept json
// @Produce json
// @Param request body models.MLExportConfig true "Profile configuration"
// @Success 201 {object} ConfigResponse "Profile created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /ml/profiles [post]
func (h *MLExportHandler) CreateProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config models.MLExportConfig
	if err := c.BodyParser(&config); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate required fields
	if config.Name == "" {
		return errors.SendError(c, errors.ValidationError("Name is required", map[string]string{
			"name": "required",
		}))
	}

	config.IsPreset = false
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := h.exportService.CreateConfig(ctx, &config); err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	response := ConfigResponse{
		ID:          config.ID.Hex(),
		Name:        config.Name,
		Description: config.Description,
		IsPreset:    config.IsPreset,
		Format:      string(config.Format),
		Config:      config,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Profile created",
	})
}

// UpdateProfile updates an export profile
// @Summary Update export profile
// @Description Updates an existing export configuration profile
// @Tags ML Export
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body models.MLExportConfig true "Profile configuration"
// @Success 200 {object} ConfigResponse "Profile updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Profile not found"
// @Router /ml/profiles/{id} [put]
func (h *MLExportHandler) UpdateProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing profile ID"))
	}

	// Get existing config
	existing, err := h.exportService.GetConfig(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Profile"))
	}

	// Don't allow updating presets
	if existing.IsPreset {
		return errors.SendError(c, errors.BadRequest("Cannot modify preset profiles"))
	}

	var config models.MLExportConfig
	if err := c.BodyParser(&config); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Preserve ID and timestamps
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()
	config.IsPreset = false

	if err := h.exportService.UpdateConfig(ctx, &config); err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	response := ConfigResponse{
		ID:          config.ID.Hex(),
		Name:        config.Name,
		Description: config.Description,
		IsPreset:    config.IsPreset,
		Format:      string(config.Format),
		Config:      config,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Profile updated",
	})
}

// DeleteProfile deletes an export profile
// @Summary Delete export profile
// @Description Deletes an export configuration profile
// @Tags ML Export
// @Param id path string true "Profile ID"
// @Success 200 {object} map[string]interface{} "Profile deleted"
// @Failure 400 {object} map[string]interface{} "Cannot delete preset"
// @Failure 404 {object} map[string]interface{} "Profile not found"
// @Router /ml/profiles/{id} [delete]
func (h *MLExportHandler) DeleteProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing profile ID"))
	}

	// Get existing config to check if it's a preset
	existing, err := h.exportService.GetConfig(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Profile"))
	}

	if existing.IsPreset {
		return errors.SendError(c, errors.BadRequest("Cannot delete preset profiles"))
	}

	if err := h.exportService.DeleteConfig(ctx, id); err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile deleted",
	})
}

// ============================================================================
// Utility Endpoints
// ============================================================================

// GetSupportedFormats returns supported export formats
// @Summary Get supported formats
// @Description Returns a list of supported export formats with details
// @Tags ML Export
// @Produce json
// @Success 200 {object} map[string]interface{} "List of formats"
// @Router /ml/formats [get]
func (h *MLExportHandler) GetSupportedFormats(c *fiber.Ctx) error {
	formats := []fiber.Map{
		{
			"id":          "csv",
			"name":        "CSV",
			"description": "Comma-separated values, compatible with pandas, Excel",
			"extension":   ".csv",
			"compression": "optional gzip",
		},
		{
			"id":          "parquet",
			"name":        "Parquet",
			"description": "Apache Parquet columnar format, efficient for large datasets",
			"extension":   ".parquet",
			"compression": "built-in",
		},
		{
			"id":          "numpy",
			"name":        "NumPy",
			"description": "NumPy NPZ format, direct loading with np.load()",
			"extension":   ".npz",
			"compression": "optional",
		},
		{
			"id":          "jsonl",
			"name":        "JSON Lines",
			"description": "Line-delimited JSON, good for streaming processing",
			"extension":   ".jsonl",
			"compression": "optional gzip",
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    formats,
	})
}

// GetAvailableFeatures returns available features for export
// @Summary Get available features
// @Description Returns a list of available features that can be exported
// @Tags ML Export
// @Produce json
// @Success 200 {object} map[string]interface{} "List of features"
// @Router /ml/features [get]
func (h *MLExportHandler) GetAvailableFeatures(c *fiber.Ctx) error {
	features := fiber.Map{
		"ohlcv": []string{"open", "high", "low", "close", "volume", "timestamp"},
		"price_features": []string{
			"returns", "log_returns", "price_change", "volatility",
			"gaps", "body_ratio", "range_pct", "upper_wick", "lower_wick",
		},
		"temporal_features": []string{
			"hour", "hour_sin", "hour_cos",
			"day_of_week", "dow_sin", "dow_cos",
			"day_of_month", "month", "month_sin", "month_cos",
			"is_weekend", "quarter",
		},
		"cross_features": []string{
			"bb_position", "price_vs_sma20", "price_vs_sma50", "price_vs_sma200",
			"ma_crossover", "rsi_oversold", "rsi_overbought",
		},
		"indicator_categories": fiber.Map{
			"trend": []string{
				"sma20", "sma50", "sma200", "ema12", "ema26", "ema50",
				"dema", "tema", "wma", "hma", "vwma",
				"ichimoku_tenkan", "ichimoku_kijun", "ichimoku_senkou_a", "ichimoku_senkou_b",
				"adx", "plus_di", "minus_di", "supertrend",
			},
			"momentum": []string{
				"rsi6", "rsi14", "rsi24", "stoch_k", "stoch_d",
				"macd", "macd_signal", "macd_hist",
				"roc", "cci", "williams_r", "momentum",
			},
			"volatility": []string{
				"bb_upper", "bb_middle", "bb_lower", "bb_bandwidth", "bb_percent_b",
				"atr", "keltner_upper", "keltner_middle", "keltner_lower",
				"donchian_upper", "donchian_middle", "donchian_lower", "stddev",
			},
			"volume": []string{
				"obv", "vwap", "mfi", "cmf", "volume_sma",
			},
		},
		"target_types": []fiber.Map{
			{
				"id":          "future_returns",
				"name":        "Future Returns",
				"description": "Percentage return over N periods",
				"output":      "float64",
			},
			{
				"id":          "future_direction",
				"name":        "Future Direction",
				"description": "Binary up/down classification",
				"output":      "0 or 1",
			},
			{
				"id":          "future_class",
				"name":        "Future Class",
				"description": "Multi-class classification with custom bins",
				"output":      "0, 1, 2, ...",
			},
			{
				"id":          "future_volatility",
				"name":        "Future Volatility",
				"description": "Average true range over N periods",
				"output":      "float64",
			},
		},
		"normalization_types": []fiber.Map{
			{"id": "none", "name": "None", "description": "No normalization"},
			{"id": "minmax", "name": "Min-Max", "description": "Scale to [0, 1] range"},
			{"id": "zscore", "name": "Z-Score", "description": "Standardize to zero mean, unit variance"},
			{"id": "robust", "name": "Robust", "description": "Scale using median and IQR"},
		},
		"nan_handling": []fiber.Map{
			{"id": "drop", "name": "Drop", "description": "Remove rows with NaN"},
			{"id": "forward_fill", "name": "Forward Fill", "description": "Fill with last valid value"},
			{"id": "backward_fill", "name": "Backward Fill", "description": "Fill with next valid value"},
			{"id": "interpolate", "name": "Interpolate", "description": "Linear interpolation"},
			{"id": "zero", "name": "Zero", "description": "Replace with zero"},
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    features,
	})
}

// GetDefaultConfig returns the default export configuration
// @Summary Get default configuration
// @Description Returns the default export configuration
// @Tags ML Export
// @Produce json
// @Success 200 {object} models.MLExportConfig "Default configuration"
// @Router /ml/config/default [get]
func (h *MLExportHandler) GetDefaultConfig(c *fiber.Ctx) error {
	config := models.DefaultMLExportConfig()

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// ============================================================================
// Dataset Endpoints (Multi-Job)
// ============================================================================

// CreateDatasetRequest is the request for creating a combined dataset
type CreateDatasetRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	JobIDs      []string             `json:"job_ids"`
	Config      models.MLExportConfig `json:"config"`
}

// CreateDataset creates a combined dataset from multiple jobs
// @Summary Create combined dataset
// @Description Creates a combined ML dataset from multiple data collection jobs
// @Tags ML Export
// @Accept json
// @Produce json
// @Param request body CreateDatasetRequest true "Dataset configuration"
// @Success 202 {object} ExportResponse "Dataset creation started"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /ml/datasets [post]
func (h *MLExportHandler) CreateDataset(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req CreateDatasetRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate
	if req.Name == "" {
		return errors.SendError(c, errors.ValidationError("Name is required", map[string]string{
			"name": "required",
		}))
	}

	if len(req.JobIDs) < 2 {
		return errors.SendError(c, errors.ValidationError("At least 2 job IDs required for a dataset", map[string]string{
			"job_ids": "at least 2 required",
		}))
	}

	// Convert to ObjectIDs
	jobIDs := make([]primitive.ObjectID, len(req.JobIDs))
	for i, idStr := range req.JobIDs {
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return errors.SendError(c, errors.ValidationError("Invalid job ID", map[string]string{
				"job_id": idStr,
			}))
		}
		jobIDs[i] = objID
	}

	// Set config name
	req.Config.Name = req.Name
	req.Config.Description = req.Description

	// Start export
	exportJob, err := h.exportService.StartExport(ctx, req.Config, jobIDs)
	if err != nil {
		return errors.SendError(c, errors.InternalError(err.Error()))
	}

	response := ExportResponse{
		ID:           exportJob.ID.Hex(),
		Status:       string(exportJob.Status),
		Progress:     exportJob.Progress,
		CurrentPhase: exportJob.CurrentPhase,
		DownloadURL:  fmt.Sprintf("/api/v1/ml/export/jobs/%s/download", exportJob.ID.Hex()),
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Dataset creation started",
	})
}

// GetExportMetadata returns detailed metadata for a completed export
// @Summary Get export metadata
// @Description Returns detailed metadata including feature schema and normalization params
// @Tags ML Export
// @Produce json
// @Param id path string true "Export job ID"
// @Success 200 {object} models.MLExportMetadata "Export metadata"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /ml/export/jobs/{id}/metadata [get]
func (h *MLExportHandler) GetExportMetadata(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return errors.SendError(c, errors.BadRequest("Missing job ID"))
	}

	exportJob, err := h.exportService.GetExportJob(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Export job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    exportJob.Metadata,
	})
}
