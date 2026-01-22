package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IndicatorConfig stores configuration for indicator calculations
type IndicatorConfig struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`           // Config name (e.g., "default", "scalping", "swing")
	IsDefault bool               `bson:"is_default" json:"is_default"` // Whether this is the default config

	// Enable/disable indicator categories
	EnableTrend      bool `bson:"enable_trend" json:"enable_trend"`
	EnableMomentum   bool `bson:"enable_momentum" json:"enable_momentum"`
	EnableVolatility bool `bson:"enable_volatility" json:"enable_volatility"`
	EnableVolume     bool `bson:"enable_volume" json:"enable_volume"`

	// Trend indicator settings
	Trend TrendConfig `bson:"trend" json:"trend"`

	// Momentum indicator settings
	Momentum MomentumConfig `bson:"momentum" json:"momentum"`

	// Volatility indicator settings
	Volatility VolatilityConfig `bson:"volatility" json:"volatility"`

	// Volume indicator settings
	Volume VolumeConfig `bson:"volume" json:"volume"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// TrendConfig holds configuration for trend indicators
type TrendConfig struct {
	// SMA settings
	SMAEnabled bool  `bson:"sma_enabled" json:"sma_enabled"`
	SMAPeriods []int `bson:"sma_periods" json:"sma_periods"` // Default: [20, 50, 200]

	// EMA settings
	EMAEnabled bool  `bson:"ema_enabled" json:"ema_enabled"`
	EMAPeriods []int `bson:"ema_periods" json:"ema_periods"` // Default: [12, 26, 50]

	// DEMA settings
	DEMAEnabled bool `bson:"dema_enabled" json:"dema_enabled"`
	DEMAPeriod  int  `bson:"dema_period" json:"dema_period"` // Default: 20

	// TEMA settings
	TEMAEnabled bool `bson:"tema_enabled" json:"tema_enabled"`
	TEMAPeriod  int  `bson:"tema_period" json:"tema_period"` // Default: 20

	// WMA settings
	WMAEnabled bool `bson:"wma_enabled" json:"wma_enabled"`
	WMAPeriod  int  `bson:"wma_period" json:"wma_period"` // Default: 20

	// HMA settings
	HMAEnabled bool `bson:"hma_enabled" json:"hma_enabled"`
	HMAPeriod  int  `bson:"hma_period" json:"hma_period"` // Default: 9

	// VWMA settings
	VWMAEnabled bool `bson:"vwma_enabled" json:"vwma_enabled"`
	VWMAPeriod  int  `bson:"vwma_period" json:"vwma_period"` // Default: 20

	// Ichimoku settings
	IchimokuEnabled      bool `bson:"ichimoku_enabled" json:"ichimoku_enabled"`
	IchimokuTenkan       int  `bson:"ichimoku_tenkan" json:"ichimoku_tenkan"`             // Default: 9
	IchimokuKijun        int  `bson:"ichimoku_kijun" json:"ichimoku_kijun"`               // Default: 26
	IchimokuSenkouB      int  `bson:"ichimoku_senkou_b" json:"ichimoku_senkou_b"`         // Default: 52
	IchimokuDisplacement int  `bson:"ichimoku_displacement" json:"ichimoku_displacement"` // Default: 26

	// ADX settings
	ADXEnabled bool `bson:"adx_enabled" json:"adx_enabled"`
	ADXPeriod  int  `bson:"adx_period" json:"adx_period"` // Default: 14

	// SuperTrend settings
	SuperTrendEnabled    bool    `bson:"supertrend_enabled" json:"supertrend_enabled"`
	SuperTrendPeriod     int     `bson:"supertrend_period" json:"supertrend_period"`         // Default: 10
	SuperTrendMultiplier float64 `bson:"supertrend_multiplier" json:"supertrend_multiplier"` // Default: 3.0
}

// MomentumConfig holds configuration for momentum indicators
type MomentumConfig struct {
	// RSI settings
	RSIEnabled bool  `bson:"rsi_enabled" json:"rsi_enabled"`
	RSIPeriods []int `bson:"rsi_periods" json:"rsi_periods"` // Default: [6, 14, 24]

	// Stochastic settings
	StochEnabled bool `bson:"stoch_enabled" json:"stoch_enabled"`
	StochK       int  `bson:"stoch_k" json:"stoch_k"`           // Default: 14
	StochD       int  `bson:"stoch_d" json:"stoch_d"`           // Default: 3
	StochSmooth  int  `bson:"stoch_smooth" json:"stoch_smooth"` // Default: 3

	// MACD settings
	MACDEnabled bool `bson:"macd_enabled" json:"macd_enabled"`
	MACDFast    int  `bson:"macd_fast" json:"macd_fast"`     // Default: 12
	MACDSlow    int  `bson:"macd_slow" json:"macd_slow"`     // Default: 26
	MACDSignal  int  `bson:"macd_signal" json:"macd_signal"` // Default: 9

	// ROC settings
	ROCEnabled bool `bson:"roc_enabled" json:"roc_enabled"`
	ROCPeriod  int  `bson:"roc_period" json:"roc_period"` // Default: 12

	// CCI settings
	CCIEnabled bool `bson:"cci_enabled" json:"cci_enabled"`
	CCIPeriod  int  `bson:"cci_period" json:"cci_period"` // Default: 20

	// Williams %R settings
	WilliamsREnabled bool `bson:"williams_r_enabled" json:"williams_r_enabled"`
	WilliamsRPeriod  int  `bson:"williams_r_period" json:"williams_r_period"` // Default: 14

	// Momentum settings
	MomentumEnabled bool `bson:"momentum_enabled" json:"momentum_enabled"`
	MomentumPeriod  int  `bson:"momentum_period" json:"momentum_period"` // Default: 10
}

// VolatilityConfig holds configuration for volatility indicators
type VolatilityConfig struct {
	// Bollinger Bands settings
	BollingerEnabled bool    `bson:"bollinger_enabled" json:"bollinger_enabled"`
	BollingerPeriod  int     `bson:"bollinger_period" json:"bollinger_period"`   // Default: 20
	BollingerStdDev  float64 `bson:"bollinger_stddev" json:"bollinger_stddev"`   // Default: 2.0

	// ATR settings
	ATREnabled bool `bson:"atr_enabled" json:"atr_enabled"`
	ATRPeriod  int  `bson:"atr_period" json:"atr_period"` // Default: 14

	// Keltner Channels settings
	KeltnerEnabled    bool    `bson:"keltner_enabled" json:"keltner_enabled"`
	KeltnerPeriod     int     `bson:"keltner_period" json:"keltner_period"`         // Default: 20
	KeltnerATRPeriod  int     `bson:"keltner_atr_period" json:"keltner_atr_period"` // Default: 10
	KeltnerMultiplier float64 `bson:"keltner_multiplier" json:"keltner_multiplier"` // Default: 2.0

	// Donchian Channels settings
	DonchianEnabled bool `bson:"donchian_enabled" json:"donchian_enabled"`
	DonchianPeriod  int  `bson:"donchian_period" json:"donchian_period"` // Default: 20

	// Standard Deviation settings
	StdDevEnabled bool `bson:"stddev_enabled" json:"stddev_enabled"`
	StdDevPeriod  int  `bson:"stddev_period" json:"stddev_period"` // Default: 20
}

// VolumeConfig holds configuration for volume indicators
type VolumeConfig struct {
	// OBV settings
	OBVEnabled bool `bson:"obv_enabled" json:"obv_enabled"`

	// VWAP settings
	VWAPEnabled bool `bson:"vwap_enabled" json:"vwap_enabled"`

	// MFI settings
	MFIEnabled bool `bson:"mfi_enabled" json:"mfi_enabled"`
	MFIPeriod  int  `bson:"mfi_period" json:"mfi_period"` // Default: 14

	// CMF settings
	CMFEnabled bool `bson:"cmf_enabled" json:"cmf_enabled"`
	CMFPeriod  int  `bson:"cmf_period" json:"cmf_period"` // Default: 20

	// Volume SMA settings
	VolumeSMAEnabled bool `bson:"volume_sma_enabled" json:"volume_sma_enabled"`
	VolumeSMAPeriod  int  `bson:"volume_sma_period" json:"volume_sma_period"` // Default: 20
}

// DefaultIndicatorConfig returns a configuration with all defaults enabled
func DefaultIndicatorConfig() *IndicatorConfig {
	return &IndicatorConfig{
		Name:             "default",
		IsDefault:        true,
		EnableTrend:      true,
		EnableMomentum:   true,
		EnableVolatility: true,
		EnableVolume:     true,
		Trend:            DefaultTrendConfig(),
		Momentum:         DefaultMomentumConfig(),
		Volatility:       DefaultVolatilityConfig(),
		Volume:           DefaultVolumeConfig(),
	}
}

// DefaultTrendConfig returns default trend indicator configuration
func DefaultTrendConfig() TrendConfig {
	return TrendConfig{
		SMAEnabled:           true,
		SMAPeriods:           []int{20, 50, 200},
		EMAEnabled:           true,
		EMAPeriods:           []int{12, 26, 50},
		DEMAEnabled:          true,
		DEMAPeriod:           20,
		TEMAEnabled:          true,
		TEMAPeriod:           20,
		WMAEnabled:           true,
		WMAPeriod:            20,
		HMAEnabled:           true,
		HMAPeriod:            9,
		VWMAEnabled:          true,
		VWMAPeriod:           20,
		IchimokuEnabled:      true,
		IchimokuTenkan:       9,
		IchimokuKijun:        26,
		IchimokuSenkouB:      52,
		IchimokuDisplacement: 26,
		ADXEnabled:           true,
		ADXPeriod:            14,
		SuperTrendEnabled:    true,
		SuperTrendPeriod:     10,
		SuperTrendMultiplier: 3.0,
	}
}

// DefaultMomentumConfig returns default momentum indicator configuration
func DefaultMomentumConfig() MomentumConfig {
	return MomentumConfig{
		RSIEnabled:       true,
		RSIPeriods:       []int{6, 14, 24},
		StochEnabled:     true,
		StochK:           14,
		StochD:           3,
		StochSmooth:      3,
		MACDEnabled:      true,
		MACDFast:         12,
		MACDSlow:         26,
		MACDSignal:       9,
		ROCEnabled:       true,
		ROCPeriod:        12,
		CCIEnabled:       true,
		CCIPeriod:        20,
		WilliamsREnabled: true,
		WilliamsRPeriod:  14,
		MomentumEnabled:  true,
		MomentumPeriod:   10,
	}
}

// DefaultVolatilityConfig returns default volatility indicator configuration
func DefaultVolatilityConfig() VolatilityConfig {
	return VolatilityConfig{
		BollingerEnabled:  true,
		BollingerPeriod:   20,
		BollingerStdDev:   2.0,
		ATREnabled:        true,
		ATRPeriod:         14,
		KeltnerEnabled:    true,
		KeltnerPeriod:     20,
		KeltnerATRPeriod:  10,
		KeltnerMultiplier: 2.0,
		DonchianEnabled:   true,
		DonchianPeriod:    20,
		StdDevEnabled:     true,
		StdDevPeriod:      20,
	}
}

// DefaultVolumeConfig returns default volume indicator configuration
func DefaultVolumeConfig() VolumeConfig {
	return VolumeConfig{
		OBVEnabled:       true,
		VWAPEnabled:      true,
		MFIEnabled:       true,
		MFIPeriod:        14,
		CMFEnabled:       true,
		CMFPeriod:        20,
		VolumeSMAEnabled: true,
		VolumeSMAPeriod:  20,
	}
}

// IndicatorConfigCreateRequest for creating a new config
type IndicatorConfigCreateRequest struct {
	Name             string           `json:"name" validate:"required"`
	IsDefault        bool             `json:"is_default"`
	EnableTrend      bool             `json:"enable_trend"`
	EnableMomentum   bool             `json:"enable_momentum"`
	EnableVolatility bool             `json:"enable_volatility"`
	EnableVolume     bool             `json:"enable_volume"`
	Trend            *TrendConfig     `json:"trend,omitempty"`
	Momentum         *MomentumConfig  `json:"momentum,omitempty"`
	Volatility       *VolatilityConfig `json:"volatility,omitempty"`
	Volume           *VolumeConfig    `json:"volume,omitempty"`
}

// IndicatorConfigUpdateRequest for updating a config
type IndicatorConfigUpdateRequest struct {
	Name             *string           `json:"name,omitempty"`
	IsDefault        *bool             `json:"is_default,omitempty"`
	EnableTrend      *bool             `json:"enable_trend,omitempty"`
	EnableMomentum   *bool             `json:"enable_momentum,omitempty"`
	EnableVolatility *bool             `json:"enable_volatility,omitempty"`
	EnableVolume     *bool             `json:"enable_volume,omitempty"`
	Trend            *TrendConfig      `json:"trend,omitempty"`
	Momentum         *MomentumConfig   `json:"momentum,omitempty"`
	Volatility       *VolatilityConfig `json:"volatility,omitempty"`
	Volume           *VolumeConfig     `json:"volume,omitempty"`
}

// Validation constants
const (
	MinPeriod        = 1
	MaxPeriod        = 500
	MinMultiplier    = 0.1
	MaxMultiplier    = 10.0
	MinStdDev        = 0.5
	MaxStdDev        = 5.0
	MaxPeriodsInList = 10
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ConfigValidationResult holds all validation errors
type ConfigValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validate validates the entire indicator configuration
func (c *IndicatorConfig) Validate() ConfigValidationResult {
	result := ConfigValidationResult{Valid: true, Errors: []ValidationError{}}

	if c.Name == "" {
		result.addError("name", "Config name is required")
	}

	if c.EnableTrend {
		trendErrors := c.Trend.Validate()
		for _, e := range trendErrors {
			e.Field = "trend." + e.Field
			result.Errors = append(result.Errors, e)
		}
	}

	if c.EnableMomentum {
		momentumErrors := c.Momentum.Validate()
		for _, e := range momentumErrors {
			e.Field = "momentum." + e.Field
			result.Errors = append(result.Errors, e)
		}
	}

	if c.EnableVolatility {
		volatilityErrors := c.Volatility.Validate()
		for _, e := range volatilityErrors {
			e.Field = "volatility." + e.Field
			result.Errors = append(result.Errors, e)
		}
	}

	if c.EnableVolume {
		volumeErrors := c.Volume.Validate()
		for _, e := range volumeErrors {
			e.Field = "volume." + e.Field
			result.Errors = append(result.Errors, e)
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

func (r *ConfigValidationResult) addError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
	r.Valid = false
}

// Validate validates trend indicator configuration
func (c *TrendConfig) Validate() []ValidationError {
	var errors []ValidationError

	if c.SMAEnabled {
		errors = append(errors, validatePeriodList("sma_periods", c.SMAPeriods)...)
	}

	if c.EMAEnabled {
		errors = append(errors, validatePeriodList("ema_periods", c.EMAPeriods)...)
	}

	if c.DEMAEnabled {
		if err := validatePeriod("dema_period", c.DEMAPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.TEMAEnabled {
		if err := validatePeriod("tema_period", c.TEMAPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.WMAEnabled {
		if err := validatePeriod("wma_period", c.WMAPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.HMAEnabled {
		if err := validatePeriod("hma_period", c.HMAPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.VWMAEnabled {
		if err := validatePeriod("vwma_period", c.VWMAPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.IchimokuEnabled {
		if err := validatePeriod("ichimoku_tenkan", c.IchimokuTenkan, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("ichimoku_kijun", c.IchimokuKijun, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("ichimoku_senkou_b", c.IchimokuSenkouB, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("ichimoku_displacement", c.IchimokuDisplacement, 1); err != nil {
			errors = append(errors, *err)
		}
		if c.IchimokuTenkan >= c.IchimokuKijun {
			errors = append(errors, ValidationError{
				Field:   "ichimoku_tenkan",
				Message: "Tenkan period must be less than Kijun period",
			})
		}
	}

	if c.ADXEnabled {
		if err := validatePeriod("adx_period", c.ADXPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.SuperTrendEnabled {
		if err := validatePeriod("supertrend_period", c.SuperTrendPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validateMultiplier("supertrend_multiplier", c.SuperTrendMultiplier); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// Validate validates momentum indicator configuration
func (c *MomentumConfig) Validate() []ValidationError {
	var errors []ValidationError

	if c.RSIEnabled {
		errors = append(errors, validatePeriodList("rsi_periods", c.RSIPeriods)...)
	}

	if c.StochEnabled {
		if err := validatePeriod("stoch_k", c.StochK, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("stoch_d", c.StochD, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("stoch_smooth", c.StochSmooth, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.MACDEnabled {
		if err := validatePeriod("macd_fast", c.MACDFast, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("macd_slow", c.MACDSlow, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("macd_signal", c.MACDSignal, 1); err != nil {
			errors = append(errors, *err)
		}
		if c.MACDFast >= c.MACDSlow {
			errors = append(errors, ValidationError{
				Field:   "macd_fast",
				Message: "MACD fast period must be less than slow period",
			})
		}
	}

	if c.ROCEnabled {
		if err := validatePeriod("roc_period", c.ROCPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.CCIEnabled {
		if err := validatePeriod("cci_period", c.CCIPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.WilliamsREnabled {
		if err := validatePeriod("williams_r_period", c.WilliamsRPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.MomentumEnabled {
		if err := validatePeriod("momentum_period", c.MomentumPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// Validate validates volatility indicator configuration
func (c *VolatilityConfig) Validate() []ValidationError {
	var errors []ValidationError

	if c.BollingerEnabled {
		if err := validatePeriod("bollinger_period", c.BollingerPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
		if err := validateStdDev("bollinger_stddev", c.BollingerStdDev); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.ATREnabled {
		if err := validatePeriod("atr_period", c.ATRPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.KeltnerEnabled {
		if err := validatePeriod("keltner_period", c.KeltnerPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validatePeriod("keltner_atr_period", c.KeltnerATRPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
		if err := validateMultiplier("keltner_multiplier", c.KeltnerMultiplier); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.DonchianEnabled {
		if err := validatePeriod("donchian_period", c.DonchianPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.StdDevEnabled {
		if err := validatePeriod("stddev_period", c.StdDevPeriod, 2); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// Validate validates volume indicator configuration
func (c *VolumeConfig) Validate() []ValidationError {
	var errors []ValidationError

	if c.MFIEnabled {
		if err := validatePeriod("mfi_period", c.MFIPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.CMFEnabled {
		if err := validatePeriod("cmf_period", c.CMFPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	if c.VolumeSMAEnabled {
		if err := validatePeriod("volume_sma_period", c.VolumeSMAPeriod, 1); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// Helper functions for validation

func validatePeriod(field string, value int, minVal int) *ValidationError {
	if value < minVal {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d", minVal),
		}
	}
	if value > MaxPeriod {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at most %d", MaxPeriod),
		}
	}
	return nil
}

func validatePeriodList(field string, values []int) []ValidationError {
	var errors []ValidationError

	if len(values) == 0 {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: "at least one period is required when indicator is enabled",
		})
		return errors
	}

	if len(values) > MaxPeriodsInList {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("cannot have more than %d periods", MaxPeriodsInList),
		})
	}

	for i, v := range values {
		if v < MinPeriod {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s[%d]", field, i),
				Message: fmt.Sprintf("must be at least %d", MinPeriod),
			})
		}
		if v > MaxPeriod {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s[%d]", field, i),
				Message: fmt.Sprintf("must be at most %d", MaxPeriod),
			})
		}
	}

	return errors
}

func validateMultiplier(field string, value float64) *ValidationError {
	if value < MinMultiplier {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %.1f", MinMultiplier),
		}
	}
	if value > MaxMultiplier {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at most %.1f", MaxMultiplier),
		}
	}
	return nil
}

func validateStdDev(field string, value float64) *ValidationError {
	if value < MinStdDev {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %.1f", MinStdDev),
		}
	}
	if value > MaxStdDev {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at most %.1f", MaxStdDev),
		}
	}
	return nil
}
