package indicators

import (
	"fmt"
	"log"

	"github.com/yourusername/datacollector/internal/models"
)

// Service handles indicator calculations for candles
type Service struct{}

// NewService creates a new indicator service
func NewService() *Service {
	return &Service{}
}

// IndicatorConfig defines which indicators to calculate and their parameters
type IndicatorConfig struct {
	// RSI configurations
	CalculateRSI bool
	RSIPeriods   []int // e.g., [6, 14, 24]

	// EMA configurations
	CalculateEMA bool
	EMAPeriods   []int // e.g., [12, 26]

	// SMA configurations
	CalculateSMA bool
	SMAPeriods   []int // e.g., [20, 50, 200]

	// MACD configuration
	CalculateMACD  bool
	MACDFast       int // e.g., 12
	MACDSlow       int // e.g., 26
	MACDSignal     int // e.g., 9

	// Bollinger Bands configuration
	CalculateBollinger bool
	BollingerPeriod    int     // e.g., 20
	BollingerStdDev    float64 // e.g., 2.0

	// ATR configuration
	CalculateATR bool
	ATRPeriod    int // e.g., 14
}

// DefaultConfig returns a default indicator configuration
func DefaultConfig() *IndicatorConfig {
	return &IndicatorConfig{
		CalculateRSI: true,
		RSIPeriods:   []int{6, 14, 24},

		CalculateEMA: true,
		EMAPeriods:   []int{12, 26},

		CalculateSMA: true,
		SMAPeriods:   []int{20, 50},

		CalculateMACD: true,
		MACDFast:      12,
		MACDSlow:      26,
		MACDSignal:    9,

		CalculateBollinger: true,
		BollingerPeriod:    20,
		BollingerStdDev:    2.0,

		CalculateATR: true,
		ATRPeriod:    14,
	}
}

// CalculateAll calculates all configured indicators and updates the candles
// Note: Candles should be ordered with newest first (index 0)
func (s *Service) CalculateAll(candles []models.Candle, config *IndicatorConfig) ([]models.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	// Reverse candles for calculation (indicators expect oldest first)
	reversed := make([]models.Candle, len(candles))
	for i := range candles {
		reversed[i] = candles[len(candles)-1-i]
	}

	log.Printf("[INDICATORS] Calculating indicators for %d candles", len(reversed))

	// Initialize indicators map for each candle
	for i := range reversed {
		if reversed[i].Indicators == (models.Indicators{}) {
			reversed[i].Indicators = models.Indicators{}
		}
	}

	// Calculate RSI
	if config.CalculateRSI {
		for _, period := range config.RSIPeriods {
			rsi := CalculateRSI(reversed, period, "close")
			for i, val := range rsi {
				if isValidValue(val) {
					switch period {
					case 6:
						reversed[i].Indicators.RSI6 = &val
					case 14:
						reversed[i].Indicators.RSI14 = &val
					case 24:
						reversed[i].Indicators.RSI24 = &val
					}
				}
			}
			log.Printf("[INDICATORS] Calculated RSI(%d)", period)
		}
	}

	// Calculate EMA
	if config.CalculateEMA {
		for _, period := range config.EMAPeriods {
			ema := CalculateEMA(reversed, period, "close")
			for i, val := range ema {
				if isValidValue(val) {
					switch period {
					case 12:
						reversed[i].Indicators.EMA12 = &val
					case 26:
						reversed[i].Indicators.EMA26 = &val
					}
				}
			}
			log.Printf("[INDICATORS] Calculated EMA(%d)", period)
		}
	}

	// Calculate MACD
	if config.CalculateMACD {
		macd := CalculateMACD(reversed, config.MACDFast, config.MACDSlow, config.MACDSignal, "close")
		for i := range reversed {
			if isValidValue(macd.MACD[i]) {
				reversed[i].Indicators.MACD = &macd.MACD[i]
			}
			if isValidValue(macd.Signal[i]) {
				reversed[i].Indicators.MACDSignal = &macd.Signal[i]
			}
			if isValidValue(macd.Histogram[i]) {
				reversed[i].Indicators.MACDHist = &macd.Histogram[i]
			}
		}
		log.Printf("[INDICATORS] Calculated MACD(%d,%d,%d)", config.MACDFast, config.MACDSlow, config.MACDSignal)
	}

	// Reverse back to newest-first order
	for i := range candles {
		candles[i] = reversed[len(reversed)-1-i]
	}

	log.Printf("[INDICATORS] Indicator calculation complete")
	return candles, nil
}

// CalculateMinimumCandles calculates the minimum number of candles needed
// for all configured indicators
func (s *Service) CalculateMinimumCandles(config *IndicatorConfig) int {
	minCandles := 0

	if config.CalculateRSI {
		for _, period := range config.RSIPeriods {
			if period+1 > minCandles {
				minCandles = period + 1
			}
		}
	}

	if config.CalculateEMA {
		for _, period := range config.EMAPeriods {
			if period > minCandles {
				minCandles = period
			}
		}
	}

	if config.CalculateSMA {
		for _, period := range config.SMAPeriods {
			if period > minCandles {
				minCandles = period
			}
		}
	}

	if config.CalculateMACD {
		required := config.MACDSlow + config.MACDSignal - 1
		if required > minCandles {
			minCandles = required
		}
	}

	if config.CalculateBollinger {
		if config.BollingerPeriod > minCandles {
			minCandles = config.BollingerPeriod
		}
	}

	if config.CalculateATR {
		if config.ATRPeriod > minCandles {
			minCandles = config.ATRPeriod
		}
	}

	return minCandles
}

// ValidateConfig validates the indicator configuration
func (s *Service) ValidateConfig(config *IndicatorConfig) error {
	if config.CalculateRSI {
		for _, period := range config.RSIPeriods {
			if period < 2 || period > 100 {
				return fmt.Errorf("RSI period %d out of range (2-100)", period)
			}
		}
	}

	if config.CalculateMACD {
		if config.MACDFast >= config.MACDSlow {
			return fmt.Errorf("MACD fast period (%d) must be less than slow period (%d)", config.MACDFast, config.MACDSlow)
		}
		if config.MACDSignal < 2 {
			return fmt.Errorf("MACD signal period must be at least 2")
		}
	}

	if config.CalculateBollinger {
		if config.BollingerPeriod < 2 {
			return fmt.Errorf("Bollinger period must be at least 2")
		}
		if config.BollingerStdDev <= 0 {
			return fmt.Errorf("Bollinger stdDev must be positive")
		}
	}

	return nil
}
