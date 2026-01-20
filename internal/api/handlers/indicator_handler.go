package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

// IndicatorHandler handles indicator-related HTTP requests
type IndicatorHandler struct {
	ohlcvRepo     *repository.OHLCVRepository
	recalcService *service.RecalculatorService
}

// NewIndicatorHandler creates a new indicator handler
func NewIndicatorHandler(
	ohlcvRepo *repository.OHLCVRepository,
	recalcService *service.RecalculatorService,
) *IndicatorHandler {
	return &IndicatorHandler{
		ohlcvRepo:     ohlcvRepo,
		recalcService: recalcService,
	}
}

// GetLatestIndicators retrieves the latest candle with indicators
// GET /api/v1/indicators/:exchange/:timeframe/latest?symbol=ETH/USDT
func (h *IndicatorHandler) GetLatestIndicators(c *fiber.Ctx) error {
	exchangeID := c.Params("exchange")
	timeframe := c.Params("timeframe")
	symbol := c.Query("symbol")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "symbol query parameter is required"})
	}

	// Fetch OHLCV document
	doc, err := h.ohlcvRepo.FindByJob(c.Context(), exchangeID, symbol, timeframe)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if doc == nil || len(doc.Candles) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No data found"})
	}

	// Return the latest candle (index 0)
	latestCandle := doc.Candles[0]

	return c.JSON(fiber.Map{
		"exchange":  exchangeID,
		"symbol":    symbol,
		"timeframe": timeframe,
		"timestamp": latestCandle.Timestamp,
		"ohlcv": fiber.Map{
			"open":   latestCandle.Open,
			"high":   latestCandle.High,
			"low":    latestCandle.Low,
			"close":  latestCandle.Close,
			"volume": latestCandle.Volume,
		},
		"indicators": latestCandle.Indicators,
	})
}

// GetIndicatorRange retrieves indicators for a range of candles
// GET /api/v1/indicators/:exchange/:timeframe/range?symbol=ETH/USDT&limit=100&offset=0
func (h *IndicatorHandler) GetIndicatorRange(c *fiber.Ctx) error {
	exchangeID := c.Params("exchange")
	timeframe := c.Params("timeframe")
	symbol := c.Query("symbol")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "symbol query parameter is required"})
	}

	// Parse query parameters
	limit := 100 // default
	offset := 0  // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Fetch OHLCV document
	doc, err := h.ohlcvRepo.FindByJob(c.Context(), exchangeID, symbol, timeframe)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if doc == nil || len(doc.Candles) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No data found"})
	}

	// Apply offset and limit
	candles := doc.Candles
	totalCandles := len(candles)

	if offset >= totalCandles {
		return c.JSON(fiber.Map{
			"exchange":      exchangeID,
			"symbol":        symbol,
			"timeframe":     timeframe,
			"total_candles": totalCandles,
			"offset":        offset,
			"limit":         limit,
			"candles":       []models.Candle{},
		})
	}

	end := offset + limit
	if end > totalCandles {
		end = totalCandles
	}

	selectedCandles := candles[offset:end]

	return c.JSON(fiber.Map{
		"exchange":      exchangeID,
		"symbol":        symbol,
		"timeframe":     timeframe,
		"total_candles": totalCandles,
		"offset":        offset,
		"limit":         limit,
		"returned":      len(selectedCandles),
		"candles":       selectedCandles,
	})
}

// GetSpecificIndicator retrieves history of a specific indicator
// GET /api/v1/indicators/:exchange/:timeframe/:indicator?symbol=ETH/USDT&limit=100
func (h *IndicatorHandler) GetSpecificIndicator(c *fiber.Ctx) error {
	exchangeID := c.Params("exchange")
	timeframe := c.Params("timeframe")
	indicatorName := c.Params("indicator")
	symbol := c.Query("symbol")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "symbol query parameter is required"})
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Fetch OHLCV document
	doc, err := h.ohlcvRepo.FindByJob(c.Context(), exchangeID, symbol, timeframe)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if doc == nil || len(doc.Candles) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No data found"})
	}

	// Extract specific indicator values
	result := make([]fiber.Map, 0)
	count := 0

	for _, candle := range doc.Candles {
		if count >= limit {
			break
		}

		value := extractIndicatorValue(candle.Indicators, indicatorName)
		if value != nil {
			result = append(result, fiber.Map{
				"timestamp": candle.Timestamp,
				"value":     value,
			})
			count++
		}
	}

	return c.JSON(fiber.Map{
		"exchange":    exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
		"indicator":   indicatorName,
		"data_points": len(result),
		"data":        result,
	})
}

// RecalculateJob triggers recalculation for a specific job
// POST /api/v1/jobs/:id/indicators/recalculate
func (h *IndicatorHandler) RecalculateJob(c *fiber.Ctx) error {
	jobID := c.Params("id")

	err := h.recalcService.RecalculateJob(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Recalculation failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Indicators recalculated successfully",
		"job_id":  jobID,
	})
}

// RecalculateConnector triggers recalculation for all jobs on a connector
// POST /api/v1/connectors/:id/indicators/recalculate
func (h *IndicatorHandler) RecalculateConnector(c *fiber.Ctx) error {
	connectorID := c.Params("id")

	err := h.recalcService.RecalculateConnector(c.Context(), connectorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Recalculation failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Indicators recalculated successfully for all jobs",
		"connector_id": connectorID,
	})
}

// Helper functions

func extractIndicatorValue(indicators models.Indicators, name string) interface{} {
	switch name {
	// Trend indicators
	case "sma20":
		return indicators.SMA20
	case "sma50":
		return indicators.SMA50
	case "sma200":
		return indicators.SMA200
	case "ema12":
		return indicators.EMA12
	case "ema26":
		return indicators.EMA26
	case "ema50":
		return indicators.EMA50
	case "dema":
		return indicators.DEMA
	case "tema":
		return indicators.TEMA
	case "wma":
		return indicators.WMA
	case "hma":
		return indicators.HMA
	case "vwma":
		return indicators.VWMA
	case "ichimoku_tenkan":
		return indicators.IchimokuTenkan
	case "ichimoku_kijun":
		return indicators.IchimokuKijun
	case "ichimoku_senkou_a":
		return indicators.IchimokuSenkouA
	case "ichimoku_senkou_b":
		return indicators.IchimokuSenkouB
	case "ichimoku_chikou":
		return indicators.IchimokuChikou
	case "adx":
		return indicators.ADX
	case "plus_di":
		return indicators.PlusDI
	case "minus_di":
		return indicators.MinusDI
	case "supertrend":
		return indicators.SuperTrend
	case "supertrend_signal":
		return indicators.SuperTrendSignal

	// Momentum indicators
	case "rsi6":
		return indicators.RSI6
	case "rsi14":
		return indicators.RSI14
	case "rsi24":
		return indicators.RSI24
	case "stoch_k":
		return indicators.StochK
	case "stoch_d":
		return indicators.StochD
	case "macd":
		return indicators.MACD
	case "macd_signal":
		return indicators.MACDSignal
	case "macd_hist":
		return indicators.MACDHist
	case "roc":
		return indicators.ROC
	case "cci":
		return indicators.CCI
	case "williams_r":
		return indicators.WilliamsR
	case "momentum":
		return indicators.Momentum

	// Volatility indicators
	case "bb_upper":
		return indicators.BollingerUpper
	case "bb_middle":
		return indicators.BollingerMiddle
	case "bb_lower":
		return indicators.BollingerLower
	case "bb_bandwidth":
		return indicators.BollingerBandwidth
	case "bb_percent_b":
		return indicators.BollingerPercentB
	case "atr":
		return indicators.ATR
	case "keltner_upper":
		return indicators.KeltnerUpper
	case "keltner_middle":
		return indicators.KeltnerMiddle
	case "keltner_lower":
		return indicators.KeltnerLower
	case "donchian_upper":
		return indicators.DonchianUpper
	case "donchian_middle":
		return indicators.DonchianMiddle
	case "donchian_lower":
		return indicators.DonchianLower
	case "stddev":
		return indicators.StdDev

	// Volume indicators
	case "obv":
		return indicators.OBV
	case "vwap":
		return indicators.VWAP
	case "mfi":
		return indicators.MFI
	case "cmf":
		return indicators.CMF
	case "volume_sma":
		return indicators.VolumeSMA

	default:
		return nil
	}
}

func parseIntParam(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
