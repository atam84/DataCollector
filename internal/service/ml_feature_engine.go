package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/yourusername/datacollector/internal/models"
)

// MLFeatureEngine handles feature engineering for ML export
type MLFeatureEngine struct{}

// NewMLFeatureEngine creates a new feature engine
func NewMLFeatureEngine() *MLFeatureEngine {
	return &MLFeatureEngine{}
}

// GenerateFeatures generates all requested features from candle data
func (e *MLFeatureEngine) GenerateFeatures(candles []models.Candle, config models.FeatureConfig) (*models.FeatureMatrix, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles provided")
	}

	// Sort candles by timestamp ascending (oldest first)
	sortedCandles := make([]models.Candle, len(candles))
	copy(sortedCandles, candles)
	sort.Slice(sortedCandles, func(i, j int) bool {
		return sortedCandles[i].Timestamp < sortedCandles[j].Timestamp
	})

	// Initialize feature matrix
	matrix := &models.FeatureMatrix{
		Columns:     []string{},
		Data:        make([][]float64, len(sortedCandles)),
		Timestamps:  make([]int64, len(sortedCandles)),
		RowCount:    len(sortedCandles),
		ColumnCount: 0,
		Schema:      []models.FeatureSchema{},
	}

	// Extract timestamps
	for i, c := range sortedCandles {
		matrix.Timestamps[i] = c.Timestamp
		matrix.Data[i] = []float64{}
	}

	// Add timestamp as feature if requested
	if config.IncludeTimestamp {
		e.addColumn(matrix, "timestamp", "float64", "ohlcv", float64SliceFromInt64(matrix.Timestamps))
	}

	// Add OHLCV features
	if config.IncludeOHLCV {
		opens := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Open })
		highs := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.High })
		lows := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Low })
		closes := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Close })

		e.addColumn(matrix, "open", "float64", "ohlcv", opens)
		e.addColumn(matrix, "high", "float64", "ohlcv", highs)
		e.addColumn(matrix, "low", "float64", "ohlcv", lows)
		e.addColumn(matrix, "close", "float64", "ohlcv", closes)
	}

	if config.IncludeVolume {
		volumes := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Volume })
		e.addColumn(matrix, "volume", "float64", "ohlcv", volumes)
	}

	// Add indicators
	e.addIndicatorFeatures(matrix, sortedCandles, config)

	// Add price-based features
	e.addPriceFeatures(matrix, sortedCandles, config.PriceFeatures)

	// Add temporal features
	e.addTemporalFeatures(matrix, sortedCandles, config.TemporalFeatures)

	// Add cross-indicator features
	e.addCrossFeatures(matrix, sortedCandles, config.CrossFeatures)

	// Add lagged features (after all base features are added)
	if config.LaggedFeatures.Enabled {
		e.addLaggedFeatures(matrix, config.LaggedFeatures)
	}

	// Add rolling features
	if config.RollingFeatures.Enabled {
		e.addRollingFeatures(matrix, config.RollingFeatures)
	}

	return matrix, nil
}

// GenerateTargets generates target variables
func (e *MLFeatureEngine) GenerateTargets(matrix *models.FeatureMatrix, candles []models.Candle, config models.TargetConfig) error {
	if !config.Enabled {
		return nil
	}

	// Sort candles
	sortedCandles := make([]models.Candle, len(candles))
	copy(sortedCandles, candles)
	sort.Slice(sortedCandles, func(i, j int) bool {
		return sortedCandles[i].Timestamp < sortedCandles[j].Timestamp
	})

	closes := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Close })

	for _, period := range config.LookaheadPeriods {
		switch config.Type {
		case models.TargetTypeFutureReturns:
			targets := e.calculateFutureReturns(closes, period)
			e.addColumn(matrix, fmt.Sprintf("target_future_returns_%d", period), "float64", "target", targets)

		case models.TargetTypeFutureDirection:
			targets := e.calculateFutureDirection(closes, period)
			e.addColumn(matrix, fmt.Sprintf("target_future_direction_%d", period), "int64", "target", targets)

		case models.TargetTypeFutureClass:
			targets := e.calculateFutureClass(closes, period, config.ClassificationBins)
			e.addColumn(matrix, fmt.Sprintf("target_future_class_%d", period), "int64", "target", targets)

		case models.TargetTypeFutureVolatility:
			highs := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.High })
			lows := extractFloats(sortedCandles, func(c models.Candle) float64 { return c.Low })
			targets := e.calculateFutureVolatility(highs, lows, closes, period)
			e.addColumn(matrix, fmt.Sprintf("target_future_volatility_%d", period), "float64", "target", targets)
		}
	}

	return nil
}

// addColumn adds a column to the feature matrix
func (e *MLFeatureEngine) addColumn(matrix *models.FeatureMatrix, name, dtype, source string, values []float64) {
	matrix.Columns = append(matrix.Columns, name)
	matrix.ColumnCount++

	// Add to each row
	for i := range matrix.Data {
		if i < len(values) {
			matrix.Data[i] = append(matrix.Data[i], values[i])
		} else {
			matrix.Data[i] = append(matrix.Data[i], math.NaN())
		}
	}

	// Add schema
	nanCount := countNaN(values)
	stats := calculateStats(values)
	matrix.Schema = append(matrix.Schema, models.FeatureSchema{
		Name:       name,
		Type:       dtype,
		Source:     source,
		NaNCount:   int64(nanCount),
		NaNPercent: float64(nanCount) / float64(len(values)) * 100,
		Min:        stats.min,
		Max:        stats.max,
		Mean:       stats.mean,
		Std:        stats.std,
	})
}

// addIndicatorFeatures adds technical indicator features
func (e *MLFeatureEngine) addIndicatorFeatures(matrix *models.FeatureMatrix, candles []models.Candle, config models.FeatureConfig) {
	// Define indicator mappings
	trendIndicators := []struct {
		name   string
		getter func(ind models.Indicators) *float64
	}{
		{"sma20", func(ind models.Indicators) *float64 { return ind.SMA20 }},
		{"sma50", func(ind models.Indicators) *float64 { return ind.SMA50 }},
		{"sma200", func(ind models.Indicators) *float64 { return ind.SMA200 }},
		{"ema12", func(ind models.Indicators) *float64 { return ind.EMA12 }},
		{"ema26", func(ind models.Indicators) *float64 { return ind.EMA26 }},
		{"ema50", func(ind models.Indicators) *float64 { return ind.EMA50 }},
		{"dema", func(ind models.Indicators) *float64 { return ind.DEMA }},
		{"tema", func(ind models.Indicators) *float64 { return ind.TEMA }},
		{"wma", func(ind models.Indicators) *float64 { return ind.WMA }},
		{"hma", func(ind models.Indicators) *float64 { return ind.HMA }},
		{"vwma", func(ind models.Indicators) *float64 { return ind.VWMA }},
		{"ichimoku_tenkan", func(ind models.Indicators) *float64 { return ind.IchimokuTenkan }},
		{"ichimoku_kijun", func(ind models.Indicators) *float64 { return ind.IchimokuKijun }},
		{"ichimoku_senkou_a", func(ind models.Indicators) *float64 { return ind.IchimokuSenkouA }},
		{"ichimoku_senkou_b", func(ind models.Indicators) *float64 { return ind.IchimokuSenkouB }},
		{"adx", func(ind models.Indicators) *float64 { return ind.ADX }},
		{"plus_di", func(ind models.Indicators) *float64 { return ind.PlusDI }},
		{"minus_di", func(ind models.Indicators) *float64 { return ind.MinusDI }},
		{"supertrend", func(ind models.Indicators) *float64 { return ind.SuperTrend }},
	}

	momentumIndicators := []struct {
		name   string
		getter func(ind models.Indicators) *float64
	}{
		{"rsi6", func(ind models.Indicators) *float64 { return ind.RSI6 }},
		{"rsi14", func(ind models.Indicators) *float64 { return ind.RSI14 }},
		{"rsi24", func(ind models.Indicators) *float64 { return ind.RSI24 }},
		{"stoch_k", func(ind models.Indicators) *float64 { return ind.StochK }},
		{"stoch_d", func(ind models.Indicators) *float64 { return ind.StochD }},
		{"macd", func(ind models.Indicators) *float64 { return ind.MACD }},
		{"macd_signal", func(ind models.Indicators) *float64 { return ind.MACDSignal }},
		{"macd_hist", func(ind models.Indicators) *float64 { return ind.MACDHist }},
		{"roc", func(ind models.Indicators) *float64 { return ind.ROC }},
		{"cci", func(ind models.Indicators) *float64 { return ind.CCI }},
		{"williams_r", func(ind models.Indicators) *float64 { return ind.WilliamsR }},
		{"momentum", func(ind models.Indicators) *float64 { return ind.Momentum }},
	}

	volatilityIndicators := []struct {
		name   string
		getter func(ind models.Indicators) *float64
	}{
		{"bb_upper", func(ind models.Indicators) *float64 { return ind.BollingerUpper }},
		{"bb_middle", func(ind models.Indicators) *float64 { return ind.BollingerMiddle }},
		{"bb_lower", func(ind models.Indicators) *float64 { return ind.BollingerLower }},
		{"bb_bandwidth", func(ind models.Indicators) *float64 { return ind.BollingerBandwidth }},
		{"bb_percent_b", func(ind models.Indicators) *float64 { return ind.BollingerPercentB }},
		{"atr", func(ind models.Indicators) *float64 { return ind.ATR }},
		{"keltner_upper", func(ind models.Indicators) *float64 { return ind.KeltnerUpper }},
		{"keltner_middle", func(ind models.Indicators) *float64 { return ind.KeltnerMiddle }},
		{"keltner_lower", func(ind models.Indicators) *float64 { return ind.KeltnerLower }},
		{"donchian_upper", func(ind models.Indicators) *float64 { return ind.DonchianUpper }},
		{"donchian_middle", func(ind models.Indicators) *float64 { return ind.DonchianMiddle }},
		{"donchian_lower", func(ind models.Indicators) *float64 { return ind.DonchianLower }},
		{"stddev", func(ind models.Indicators) *float64 { return ind.StdDev }},
	}

	volumeIndicators := []struct {
		name   string
		getter func(ind models.Indicators) *float64
	}{
		{"obv", func(ind models.Indicators) *float64 { return ind.OBV }},
		{"vwap", func(ind models.Indicators) *float64 { return ind.VWAP }},
		{"mfi", func(ind models.Indicators) *float64 { return ind.MFI }},
		{"cmf", func(ind models.Indicators) *float64 { return ind.CMF }},
		{"volume_sma", func(ind models.Indicators) *float64 { return ind.VolumeSMA }},
	}

	// Helper to check if indicator should be included
	shouldInclude := func(name, category string) bool {
		if config.IncludeAllIndicators {
			// Check exclusions
			for _, exc := range config.ExcludeIndicators {
				if exc == name {
					return false
				}
			}
			return true
		}

		// Check specific indicators
		for _, inc := range config.SpecificIndicators {
			if inc == name {
				return true
			}
		}

		// Check categories
		for _, cat := range config.IndicatorCategories {
			if cat == category {
				return true
			}
		}

		return false
	}

	// Add trend indicators
	for _, ind := range trendIndicators {
		if shouldInclude(ind.name, "trend") {
			values := extractIndicator(candles, ind.getter)
			e.addColumn(matrix, ind.name, "float64", "indicator", values)
		}
	}

	// Add momentum indicators
	for _, ind := range momentumIndicators {
		if shouldInclude(ind.name, "momentum") {
			values := extractIndicator(candles, ind.getter)
			e.addColumn(matrix, ind.name, "float64", "indicator", values)
		}
	}

	// Add volatility indicators
	for _, ind := range volatilityIndicators {
		if shouldInclude(ind.name, "volatility") {
			values := extractIndicator(candles, ind.getter)
			e.addColumn(matrix, ind.name, "float64", "indicator", values)
		}
	}

	// Add volume indicators
	for _, ind := range volumeIndicators {
		if shouldInclude(ind.name, "volume") {
			values := extractIndicator(candles, ind.getter)
			e.addColumn(matrix, ind.name, "float64", "indicator", values)
		}
	}
}

// addPriceFeatures adds price-based features
func (e *MLFeatureEngine) addPriceFeatures(matrix *models.FeatureMatrix, candles []models.Candle, features []string) {
	if len(features) == 0 {
		return
	}

	opens := extractFloats(candles, func(c models.Candle) float64 { return c.Open })
	highs := extractFloats(candles, func(c models.Candle) float64 { return c.High })
	lows := extractFloats(candles, func(c models.Candle) float64 { return c.Low })
	closes := extractFloats(candles, func(c models.Candle) float64 { return c.Close })

	for _, feature := range features {
		switch feature {
		case "returns":
			values := e.calculateReturns(closes)
			e.addColumn(matrix, "returns", "float64", "price_feature", values)

		case "log_returns":
			values := e.calculateLogReturns(closes)
			e.addColumn(matrix, "log_returns", "float64", "price_feature", values)

		case "price_change":
			values := e.calculatePriceChange(closes)
			e.addColumn(matrix, "price_change", "float64", "price_feature", values)

		case "volatility":
			values := e.calculateVolatility(highs, lows)
			e.addColumn(matrix, "volatility", "float64", "price_feature", values)

		case "gaps":
			values := e.calculateGaps(opens, closes)
			e.addColumn(matrix, "gaps", "float64", "price_feature", values)

		case "body_ratio":
			values := e.calculateBodyRatio(opens, highs, lows, closes)
			e.addColumn(matrix, "body_ratio", "float64", "price_feature", values)

		case "range_pct":
			values := e.calculateRangePct(highs, lows, closes)
			e.addColumn(matrix, "range_pct", "float64", "price_feature", values)

		case "upper_wick":
			values := e.calculateUpperWick(opens, highs, closes)
			e.addColumn(matrix, "upper_wick", "float64", "price_feature", values)

		case "lower_wick":
			values := e.calculateLowerWick(opens, lows, closes)
			e.addColumn(matrix, "lower_wick", "float64", "price_feature", values)
		}
	}
}

// addTemporalFeatures adds time-based features
func (e *MLFeatureEngine) addTemporalFeatures(matrix *models.FeatureMatrix, candles []models.Candle, features []string) {
	if len(features) == 0 {
		return
	}

	for _, feature := range features {
		values := make([]float64, len(candles))

		for i, c := range candles {
			t := time.UnixMilli(c.Timestamp).UTC()

			switch feature {
			case "hour":
				values[i] = float64(t.Hour())
			case "hour_sin":
				values[i] = math.Sin(2 * math.Pi * float64(t.Hour()) / 24)
			case "hour_cos":
				values[i] = math.Cos(2 * math.Pi * float64(t.Hour()) / 24)
			case "day_of_week":
				values[i] = float64(t.Weekday())
			case "dow_sin":
				values[i] = math.Sin(2 * math.Pi * float64(t.Weekday()) / 7)
			case "dow_cos":
				values[i] = math.Cos(2 * math.Pi * float64(t.Weekday()) / 7)
			case "day_of_month":
				values[i] = float64(t.Day())
			case "month":
				values[i] = float64(t.Month())
			case "month_sin":
				values[i] = math.Sin(2 * math.Pi * float64(t.Month()) / 12)
			case "month_cos":
				values[i] = math.Cos(2 * math.Pi * float64(t.Month()) / 12)
			case "is_weekend":
				if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
					values[i] = 1.0
				} else {
					values[i] = 0.0
				}
			case "quarter":
				values[i] = float64((t.Month()-1)/3 + 1)
			}
		}

		e.addColumn(matrix, feature, "float64", "temporal", values)
	}
}

// addCrossFeatures adds cross-indicator features
func (e *MLFeatureEngine) addCrossFeatures(matrix *models.FeatureMatrix, candles []models.Candle, features []string) {
	if len(features) == 0 {
		return
	}

	closes := extractFloats(candles, func(c models.Candle) float64 { return c.Close })

	for _, feature := range features {
		switch feature {
		case "bb_position":
			bbUpper := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.BollingerUpper })
			bbLower := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.BollingerLower })
			values := e.calculateBBPosition(closes, bbUpper, bbLower)
			e.addColumn(matrix, "bb_position", "float64", "cross", values)

		case "price_vs_sma20":
			sma20 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.SMA20 })
			values := e.calculatePriceVsMA(closes, sma20)
			e.addColumn(matrix, "price_vs_sma20", "float64", "cross", values)

		case "price_vs_sma50":
			sma50 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.SMA50 })
			values := e.calculatePriceVsMA(closes, sma50)
			e.addColumn(matrix, "price_vs_sma50", "float64", "cross", values)

		case "price_vs_sma200":
			sma200 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.SMA200 })
			values := e.calculatePriceVsMA(closes, sma200)
			e.addColumn(matrix, "price_vs_sma200", "float64", "cross", values)

		case "ma_crossover":
			ema12 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.EMA12 })
			ema26 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.EMA26 })
			values := e.calculateMACrossover(ema12, ema26)
			e.addColumn(matrix, "ma_crossover", "float64", "cross", values)

		case "rsi_oversold":
			rsi14 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.RSI14 })
			values := e.calculateRSIOversold(rsi14)
			e.addColumn(matrix, "rsi_oversold", "float64", "cross", values)

		case "rsi_overbought":
			rsi14 := extractIndicator(candles, func(ind models.Indicators) *float64 { return ind.RSI14 })
			values := e.calculateRSIOverbought(rsi14)
			e.addColumn(matrix, "rsi_overbought", "float64", "cross", values)
		}
	}
}

// addLaggedFeatures adds lagged versions of features
func (e *MLFeatureEngine) addLaggedFeatures(matrix *models.FeatureMatrix, config models.LagConfig) {
	if len(config.LagPeriods) == 0 {
		return
	}

	// Determine which columns to lag
	columnsToLag := config.LagFeatures
	if len(columnsToLag) == 0 {
		// Default: lag close and returns if present
		for _, col := range matrix.Columns {
			if col == "close" || col == "returns" || col == "volume" {
				columnsToLag = append(columnsToLag, col)
			}
		}
	}

	// Get column indices
	colIndices := make(map[string]int)
	for i, col := range matrix.Columns {
		colIndices[col] = i
	}

	// Add lagged features
	for _, col := range columnsToLag {
		idx, ok := colIndices[col]
		if !ok {
			continue
		}

		// Extract column values
		values := make([]float64, len(matrix.Data))
		for i, row := range matrix.Data {
			values[i] = row[idx]
		}

		// Create lagged versions
		for _, lag := range config.LagPeriods {
			laggedValues := e.lagSeries(values, lag)
			e.addColumn(matrix, fmt.Sprintf("%s_lag_%d", col, lag), "float64", "lagged", laggedValues)
		}
	}
}

// addRollingFeatures adds rolling statistics
func (e *MLFeatureEngine) addRollingFeatures(matrix *models.FeatureMatrix, config models.RollingConfig) {
	if len(config.Windows) == 0 || len(config.Stats) == 0 {
		return
	}

	// Determine which columns to use
	columnsToRoll := config.RollingFeatures
	if len(columnsToRoll) == 0 {
		// Default: close only
		columnsToRoll = []string{"close"}
	}

	// Get column indices
	colIndices := make(map[string]int)
	for i, col := range matrix.Columns {
		colIndices[col] = i
	}

	// Add rolling features
	for _, col := range columnsToRoll {
		idx, ok := colIndices[col]
		if !ok {
			continue
		}

		// Extract column values
		values := make([]float64, len(matrix.Data))
		for i, row := range matrix.Data {
			values[i] = row[idx]
		}

		// Create rolling stats for each window
		for _, window := range config.Windows {
			for _, stat := range config.Stats {
				var rolledValues []float64

				switch stat {
				case "mean":
					rolledValues = e.rollingMean(values, window)
				case "std":
					rolledValues = e.rollingStd(values, window)
				case "min":
					rolledValues = e.rollingMin(values, window)
				case "max":
					rolledValues = e.rollingMax(values, window)
				case "median":
					rolledValues = e.rollingMedian(values, window)
				default:
					continue
				}

				e.addColumn(matrix, fmt.Sprintf("%s_roll_%d_%s", col, window, stat), "float64", "rolling", rolledValues)
			}
		}
	}
}

// ==================== Price Feature Calculations ====================

func (e *MLFeatureEngine) calculateReturns(closes []float64) []float64 {
	result := make([]float64, len(closes))
	result[0] = 0

	for i := 1; i < len(closes); i++ {
		if closes[i-1] != 0 {
			result[i] = (closes[i] - closes[i-1]) / closes[i-1]
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateLogReturns(closes []float64) []float64 {
	result := make([]float64, len(closes))
	result[0] = 0

	for i := 1; i < len(closes); i++ {
		if closes[i-1] > 0 && closes[i] > 0 {
			result[i] = math.Log(closes[i] / closes[i-1])
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculatePriceChange(closes []float64) []float64 {
	result := make([]float64, len(closes))
	result[0] = 0

	for i := 1; i < len(closes); i++ {
		result[i] = closes[i] - closes[i-1]
	}
	return result
}

func (e *MLFeatureEngine) calculateVolatility(highs, lows []float64) []float64 {
	result := make([]float64, len(highs))
	for i := range highs {
		result[i] = highs[i] - lows[i]
	}
	return result
}

func (e *MLFeatureEngine) calculateGaps(opens, closes []float64) []float64 {
	result := make([]float64, len(opens))
	result[0] = 0

	for i := 1; i < len(opens); i++ {
		result[i] = opens[i] - closes[i-1]
	}
	return result
}

func (e *MLFeatureEngine) calculateBodyRatio(opens, highs, lows, closes []float64) []float64 {
	result := make([]float64, len(opens))
	for i := range opens {
		range_ := highs[i] - lows[i]
		if range_ > 0 {
			result[i] = (closes[i] - opens[i]) / range_
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateRangePct(highs, lows, closes []float64) []float64 {
	result := make([]float64, len(highs))
	for i := range highs {
		if closes[i] > 0 {
			result[i] = (highs[i] - lows[i]) / closes[i]
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateUpperWick(opens, highs, closes []float64) []float64 {
	result := make([]float64, len(opens))
	for i := range opens {
		maxBody := math.Max(opens[i], closes[i])
		result[i] = highs[i] - maxBody
	}
	return result
}

func (e *MLFeatureEngine) calculateLowerWick(opens, lows, closes []float64) []float64 {
	result := make([]float64, len(opens))
	for i := range opens {
		minBody := math.Min(opens[i], closes[i])
		result[i] = minBody - lows[i]
	}
	return result
}

// ==================== Target Calculations ====================

func (e *MLFeatureEngine) calculateFutureReturns(closes []float64, period int) []float64 {
	result := make([]float64, len(closes))
	for i := 0; i < len(closes)-period; i++ {
		if closes[i] != 0 {
			result[i] = (closes[i+period] - closes[i]) / closes[i]
		} else {
			result[i] = math.NaN()
		}
	}
	// Fill remaining with NaN
	for i := len(closes) - period; i < len(closes); i++ {
		result[i] = math.NaN()
	}
	return result
}

func (e *MLFeatureEngine) calculateFutureDirection(closes []float64, period int) []float64 {
	result := make([]float64, len(closes))
	for i := 0; i < len(closes)-period; i++ {
		if closes[i+period] > closes[i] {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	for i := len(closes) - period; i < len(closes); i++ {
		result[i] = math.NaN()
	}
	return result
}

func (e *MLFeatureEngine) calculateFutureClass(closes []float64, period int, bins []float64) []float64 {
	returns := e.calculateFutureReturns(closes, period)
	result := make([]float64, len(returns))

	for i, ret := range returns {
		if math.IsNaN(ret) {
			result[i] = math.NaN()
			continue
		}

		// Find bin
		class := float64(len(bins))
		for j, threshold := range bins {
			if ret < threshold {
				class = float64(j)
				break
			}
		}
		result[i] = class
	}
	return result
}

func (e *MLFeatureEngine) calculateFutureVolatility(highs, lows, closes []float64, period int) []float64 {
	result := make([]float64, len(closes))

	for i := 0; i < len(closes)-period; i++ {
		// Calculate average true range over next period
		sum := 0.0
		for j := i + 1; j <= i+period; j++ {
			tr := math.Max(highs[j]-lows[j], math.Max(math.Abs(highs[j]-closes[j-1]), math.Abs(lows[j]-closes[j-1])))
			sum += tr
		}
		result[i] = sum / float64(period)
	}

	for i := len(closes) - period; i < len(closes); i++ {
		result[i] = math.NaN()
	}
	return result
}

// ==================== Cross Feature Calculations ====================

func (e *MLFeatureEngine) calculateBBPosition(closes, bbUpper, bbLower []float64) []float64 {
	result := make([]float64, len(closes))
	for i := range closes {
		if math.IsNaN(bbUpper[i]) || math.IsNaN(bbLower[i]) {
			result[i] = math.NaN()
		} else {
			range_ := bbUpper[i] - bbLower[i]
			if range_ > 0 {
				result[i] = (closes[i] - bbLower[i]) / range_
			} else {
				result[i] = 0.5
			}
		}
	}
	return result
}

func (e *MLFeatureEngine) calculatePriceVsMA(closes, ma []float64) []float64 {
	result := make([]float64, len(closes))
	for i := range closes {
		if math.IsNaN(ma[i]) || ma[i] == 0 {
			result[i] = math.NaN()
		} else {
			result[i] = (closes[i] - ma[i]) / ma[i]
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateMACrossover(fast, slow []float64) []float64 {
	result := make([]float64, len(fast))
	for i := range fast {
		if math.IsNaN(fast[i]) || math.IsNaN(slow[i]) {
			result[i] = math.NaN()
		} else if fast[i] > slow[i] {
			result[i] = 1
		} else if fast[i] < slow[i] {
			result[i] = -1
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateRSIOversold(rsi []float64) []float64 {
	result := make([]float64, len(rsi))
	for i := range rsi {
		if math.IsNaN(rsi[i]) {
			result[i] = math.NaN()
		} else if rsi[i] < 30 {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) calculateRSIOverbought(rsi []float64) []float64 {
	result := make([]float64, len(rsi))
	for i := range rsi {
		if math.IsNaN(rsi[i]) {
			result[i] = math.NaN()
		} else if rsi[i] > 70 {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return result
}

// ==================== Lag and Rolling Calculations ====================

func (e *MLFeatureEngine) lagSeries(values []float64, lag int) []float64 {
	result := make([]float64, len(values))
	for i := 0; i < lag; i++ {
		result[i] = math.NaN()
	}
	for i := lag; i < len(values); i++ {
		result[i] = values[i-lag]
	}
	return result
}

func (e *MLFeatureEngine) rollingMean(values []float64, window int) []float64 {
	result := make([]float64, len(values))
	for i := 0; i < window-1; i++ {
		result[i] = math.NaN()
	}
	for i := window - 1; i < len(values); i++ {
		sum := 0.0
		count := 0
		for j := i - window + 1; j <= i; j++ {
			if !math.IsNaN(values[j]) {
				sum += values[j]
				count++
			}
		}
		if count > 0 {
			result[i] = sum / float64(count)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

func (e *MLFeatureEngine) rollingStd(values []float64, window int) []float64 {
	means := e.rollingMean(values, window)
	result := make([]float64, len(values))

	for i := 0; i < window-1; i++ {
		result[i] = math.NaN()
	}
	for i := window - 1; i < len(values); i++ {
		if math.IsNaN(means[i]) {
			result[i] = math.NaN()
			continue
		}

		sumSq := 0.0
		count := 0
		for j := i - window + 1; j <= i; j++ {
			if !math.IsNaN(values[j]) {
				diff := values[j] - means[i]
				sumSq += diff * diff
				count++
			}
		}
		if count > 1 {
			result[i] = math.Sqrt(sumSq / float64(count-1))
		} else {
			result[i] = 0
		}
	}
	return result
}

func (e *MLFeatureEngine) rollingMin(values []float64, window int) []float64 {
	result := make([]float64, len(values))
	for i := 0; i < window-1; i++ {
		result[i] = math.NaN()
	}
	for i := window - 1; i < len(values); i++ {
		min := math.Inf(1)
		for j := i - window + 1; j <= i; j++ {
			if !math.IsNaN(values[j]) && values[j] < min {
				min = values[j]
			}
		}
		if math.IsInf(min, 1) {
			result[i] = math.NaN()
		} else {
			result[i] = min
		}
	}
	return result
}

func (e *MLFeatureEngine) rollingMax(values []float64, window int) []float64 {
	result := make([]float64, len(values))
	for i := 0; i < window-1; i++ {
		result[i] = math.NaN()
	}
	for i := window - 1; i < len(values); i++ {
		max := math.Inf(-1)
		for j := i - window + 1; j <= i; j++ {
			if !math.IsNaN(values[j]) && values[j] > max {
				max = values[j]
			}
		}
		if math.IsInf(max, -1) {
			result[i] = math.NaN()
		} else {
			result[i] = max
		}
	}
	return result
}

func (e *MLFeatureEngine) rollingMedian(values []float64, window int) []float64 {
	result := make([]float64, len(values))
	for i := 0; i < window-1; i++ {
		result[i] = math.NaN()
	}
	for i := window - 1; i < len(values); i++ {
		windowVals := make([]float64, 0, window)
		for j := i - window + 1; j <= i; j++ {
			if !math.IsNaN(values[j]) {
				windowVals = append(windowVals, values[j])
			}
		}
		if len(windowVals) == 0 {
			result[i] = math.NaN()
		} else {
			sort.Float64s(windowVals)
			mid := len(windowVals) / 2
			if len(windowVals)%2 == 0 {
				result[i] = (windowVals[mid-1] + windowVals[mid]) / 2
			} else {
				result[i] = windowVals[mid]
			}
		}
	}
	return result
}

// ==================== Helper Functions ====================

func extractFloats(candles []models.Candle, getter func(c models.Candle) float64) []float64 {
	result := make([]float64, len(candles))
	for i, c := range candles {
		result[i] = getter(c)
	}
	return result
}

func extractIndicator(candles []models.Candle, getter func(ind models.Indicators) *float64) []float64 {
	result := make([]float64, len(candles))
	for i, c := range candles {
		val := getter(c.Indicators)
		if val != nil {
			result[i] = *val
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

func float64SliceFromInt64(vals []int64) []float64 {
	result := make([]float64, len(vals))
	for i, v := range vals {
		result[i] = float64(v)
	}
	return result
}

func countNaN(values []float64) int {
	count := 0
	for _, v := range values {
		if math.IsNaN(v) {
			count++
		}
	}
	return count
}

type statsResult struct {
	min, max, mean, std float64
}

func calculateStats(values []float64) statsResult {
	if len(values) == 0 {
		return statsResult{}
	}

	min := math.Inf(1)
	max := math.Inf(-1)
	sum := 0.0
	count := 0

	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
		count++
	}

	if count == 0 {
		return statsResult{}
	}

	mean := sum / float64(count)

	// Calculate std
	sumSq := 0.0
	for _, v := range values {
		if !math.IsNaN(v) {
			diff := v - mean
			sumSq += diff * diff
		}
	}
	std := 0.0
	if count > 1 {
		std = math.Sqrt(sumSq / float64(count-1))
	}

	return statsResult{min: min, max: max, mean: mean, std: std}
}
