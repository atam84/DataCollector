package models

// IndicatorConfig defines the configuration for all technical indicators
// This can be attached to Connectors (default for all jobs) or Jobs (override connector defaults)
type IndicatorConfig struct {
	// Trend Indicators
	SMA        *SMAConfig        `bson:"sma,omitempty" json:"sma,omitempty"`
	EMA        *EMAConfig        `bson:"ema,omitempty" json:"ema,omitempty"`
	DEMA       *DEMAConfig       `bson:"dema,omitempty" json:"dema,omitempty"`
	TEMA       *TEMAConfig       `bson:"tema,omitempty" json:"tema,omitempty"`
	WMA        *WMAConfig        `bson:"wma,omitempty" json:"wma,omitempty"`
	HMA        *HMAConfig        `bson:"hma,omitempty" json:"hma,omitempty"`
	VWMA       *VWMAConfig       `bson:"vwma,omitempty" json:"vwma,omitempty"`
	Ichimoku   *IchimokuConfig   `bson:"ichimoku,omitempty" json:"ichimoku,omitempty"`
	ADX        *ADXConfig        `bson:"adx,omitempty" json:"adx,omitempty"`
	SuperTrend *SuperTrendConfig `bson:"supertrend,omitempty" json:"supertrend,omitempty"`

	// Momentum Indicators
	RSI        *RSIConfig        `bson:"rsi,omitempty" json:"rsi,omitempty"`
	Stochastic *StochasticConfig `bson:"stochastic,omitempty" json:"stochastic,omitempty"`
	MACD       *MACDConfig       `bson:"macd,omitempty" json:"macd,omitempty"`
	ROC        *ROCConfig        `bson:"roc,omitempty" json:"roc,omitempty"`
	CCI        *CCIConfig        `bson:"cci,omitempty" json:"cci,omitempty"`
	WilliamsR  *WilliamsRConfig  `bson:"williams_r,omitempty" json:"williams_r,omitempty"`
	Momentum   *MomentumConfig   `bson:"momentum,omitempty" json:"momentum,omitempty"`

	// Volatility Indicators
	Bollinger *BollingerConfig `bson:"bollinger,omitempty" json:"bollinger,omitempty"`
	ATR       *ATRConfig       `bson:"atr,omitempty" json:"atr,omitempty"`
	Keltner   *KeltnerConfig   `bson:"keltner,omitempty" json:"keltner,omitempty"`
	Donchian  *DonchianConfig  `bson:"donchian,omitempty" json:"donchian,omitempty"`
	StdDev    *StdDevConfig    `bson:"stddev,omitempty" json:"stddev,omitempty"`

	// Volume Indicators
	OBV       *OBVConfig       `bson:"obv,omitempty" json:"obv,omitempty"`
	VWAP      *VWAPConfig      `bson:"vwap,omitempty" json:"vwap,omitempty"`
	MFI       *MFIConfig       `bson:"mfi,omitempty" json:"mfi,omitempty"`
	CMF       *CMFConfig       `bson:"cmf,omitempty" json:"cmf,omitempty"`
	VolumeSMA *VolumeSMAConfig `bson:"volume_sma,omitempty" json:"volume_sma,omitempty"`
}

// ==================== Trend Indicator Configs ====================

type SMAConfig struct {
	Enabled bool  `bson:"enabled" json:"enabled"`
	Periods []int `bson:"periods" json:"periods"` // e.g., [20, 50, 200]
}

type EMAConfig struct {
	Enabled bool  `bson:"enabled" json:"enabled"`
	Periods []int `bson:"periods" json:"periods"` // e.g., [12, 26, 50]
}

type DEMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type TEMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type WMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type HMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 9
}

type VWMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type IchimokuConfig struct {
	Enabled         bool `bson:"enabled" json:"enabled"`
	TenkanPeriod    int  `bson:"tenkan_period" json:"tenkan_period"`       // Conversion Line, default: 9
	KijunPeriod     int  `bson:"kijun_period" json:"kijun_period"`         // Base Line, default: 26
	SenkouBPeriod   int  `bson:"senkou_b_period" json:"senkou_b_period"`   // Leading Span B, default: 52
	DisplacementFwd int  `bson:"displacement_fwd" json:"displacement_fwd"` // Forward displacement for Senkou, default: 26
	DisplacementBck int  `bson:"displacement_bck" json:"displacement_bck"` // Backward displacement for Chikou, default: 26
}

type ADXConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 14
}

type SuperTrendConfig struct {
	Enabled    bool    `bson:"enabled" json:"enabled"`
	Period     int     `bson:"period" json:"period"`         // ATR period, e.g., 10
	Multiplier float64 `bson:"multiplier" json:"multiplier"` // ATR multiplier, e.g., 3.0
}

// ==================== Momentum Indicator Configs ====================

type RSIConfig struct {
	Enabled bool  `bson:"enabled" json:"enabled"`
	Periods []int `bson:"periods" json:"periods"` // e.g., [6, 14, 24]
}

type StochasticConfig struct {
	Enabled  bool `bson:"enabled" json:"enabled"`
	KPeriod  int  `bson:"k_period" json:"k_period"`   // %K period, e.g., 14
	DPeriod  int  `bson:"d_period" json:"d_period"`   // %D period (SMA of %K), e.g., 3
	Smooth   int  `bson:"smooth" json:"smooth"`       // Smoothing period, e.g., 3
}

type MACDConfig struct {
	Enabled      bool `bson:"enabled" json:"enabled"`
	FastPeriod   int  `bson:"fast_period" json:"fast_period"`     // e.g., 12
	SlowPeriod   int  `bson:"slow_period" json:"slow_period"`     // e.g., 26
	SignalPeriod int  `bson:"signal_period" json:"signal_period"` // e.g., 9
}

type ROCConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 12
}

type CCIConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type WilliamsRConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 14
}

type MomentumConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 10
}

// ==================== Volatility Indicator Configs ====================

type BollingerConfig struct {
	Enabled    bool    `bson:"enabled" json:"enabled"`
	Period     int     `bson:"period" json:"period"`         // SMA period, e.g., 20
	StdDev     float64 `bson:"std_dev" json:"std_dev"`       // Standard deviations, e.g., 2.0
}

type ATRConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 14
}

type KeltnerConfig struct {
	Enabled    bool    `bson:"enabled" json:"enabled"`
	Period     int     `bson:"period" json:"period"`         // EMA period, e.g., 20
	ATRPeriod  int     `bson:"atr_period" json:"atr_period"` // ATR period, e.g., 10
	Multiplier float64 `bson:"multiplier" json:"multiplier"` // ATR multiplier, e.g., 2.0
}

type DonchianConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type StdDevConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

// ==================== Volume Indicator Configs ====================

type OBVConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	// OBV has no period, it's cumulative
}

type VWAPConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	// VWAP resets daily or is calculated from start
}

type MFIConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 14
}

type CMFConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}

type VolumeSMAConfig struct {
	Enabled bool `bson:"enabled" json:"enabled"`
	Period  int  `bson:"period" json:"period"` // e.g., 20
}
