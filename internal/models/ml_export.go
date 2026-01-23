package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MLExportFormat represents supported export formats
type MLExportFormat string

const (
	MLExportFormatCSV     MLExportFormat = "csv"
	MLExportFormatParquet MLExportFormat = "parquet"
	MLExportFormatNumpy   MLExportFormat = "numpy"
	MLExportFormatJSONL   MLExportFormat = "jsonl"
)

// MLExportStatus represents the status of an export job
type MLExportStatus string

const (
	MLExportStatusPending   MLExportStatus = "pending"
	MLExportStatusRunning   MLExportStatus = "running"
	MLExportStatusCompleted MLExportStatus = "completed"
	MLExportStatusFailed    MLExportStatus = "failed"
	MLExportStatusCancelled MLExportStatus = "cancelled"
)

// NormalizationType represents normalization methods
type NormalizationType string

const (
	NormalizationNone   NormalizationType = "none"
	NormalizationMinMax NormalizationType = "minmax"
	NormalizationZScore NormalizationType = "zscore"
	NormalizationRobust NormalizationType = "robust"
)

// NaNHandlingType represents NaN handling strategies
type NaNHandlingType string

const (
	NaNHandlingDrop        NaNHandlingType = "drop"
	NaNHandlingForwardFill NaNHandlingType = "forward_fill"
	NaNHandlingBackwardFill NaNHandlingType = "backward_fill"
	NaNHandlingInterpolate NaNHandlingType = "interpolate"
	NaNHandlingZero        NaNHandlingType = "zero"
)

// TargetType represents target variable types
type TargetType string

const (
	TargetTypeFutureReturns   TargetType = "future_returns"
	TargetTypeFutureDirection TargetType = "future_direction"
	TargetTypeFutureClass     TargetType = "future_class"
	TargetTypeFutureVolatility TargetType = "future_volatility"
)

// MLExportConfig represents configuration for ML data export
type MLExportConfig struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description,omitempty" json:"description,omitempty"`
	IsPreset      bool               `bson:"is_preset" json:"is_preset"`
	IsDefault     bool               `bson:"is_default" json:"is_default"`
	Format        MLExportFormat     `bson:"format" json:"format"`
	Features      FeatureConfig      `bson:"features" json:"features"`
	Target        TargetConfig       `bson:"target" json:"target"`
	Preprocessing PreprocessConfig   `bson:"preprocessing" json:"preprocessing"`
	Split         SplitConfig        `bson:"split" json:"split"`
	Sequence      SequenceConfig     `bson:"sequence" json:"sequence"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// FeatureConfig defines which features to include in export
type FeatureConfig struct {
	// Base data
	IncludeOHLCV    bool `bson:"include_ohlcv" json:"include_ohlcv"`
	IncludeVolume   bool `bson:"include_volume" json:"include_volume"`
	IncludeTimestamp bool `bson:"include_timestamp" json:"include_timestamp"`

	// Indicator selection
	IncludeAllIndicators bool     `bson:"include_all_indicators" json:"include_all_indicators"`
	IndicatorCategories  []string `bson:"indicator_categories,omitempty" json:"indicator_categories,omitempty"` // trend, momentum, volatility, volume
	SpecificIndicators   []string `bson:"specific_indicators,omitempty" json:"specific_indicators,omitempty"`   // sma20, rsi14, etc.
	ExcludeIndicators    []string `bson:"exclude_indicators,omitempty" json:"exclude_indicators,omitempty"`

	// Price-based features
	PriceFeatures []string `bson:"price_features,omitempty" json:"price_features,omitempty"` // returns, log_returns, volatility, price_change, gaps, body_ratio, range_pct

	// Lagged features
	LaggedFeatures LagConfig `bson:"lagged_features" json:"lagged_features"`

	// Rolling statistics
	RollingFeatures RollingConfig `bson:"rolling_features" json:"rolling_features"`

	// Temporal features
	TemporalFeatures []string `bson:"temporal_features,omitempty" json:"temporal_features,omitempty"` // hour, hour_sin, hour_cos, day_of_week, dow_sin, dow_cos, month, is_weekend

	// Cross-indicator features
	CrossFeatures []string `bson:"cross_features,omitempty" json:"cross_features,omitempty"` // bb_position, price_vs_sma20, ma_crossover, rsi_divergence
}

// LagConfig defines lagged feature configuration
type LagConfig struct {
	Enabled      bool     `bson:"enabled" json:"enabled"`
	LagPeriods   []int    `bson:"lag_periods,omitempty" json:"lag_periods,omitempty"`     // e.g., [1, 5, 10, 20]
	LagFeatures  []string `bson:"lag_features,omitempty" json:"lag_features,omitempty"`   // Which features to lag, empty = all
}

// RollingConfig defines rolling statistics configuration
type RollingConfig struct {
	Enabled         bool     `bson:"enabled" json:"enabled"`
	Windows         []int    `bson:"windows,omitempty" json:"windows,omitempty"`           // e.g., [5, 10, 20]
	Stats           []string `bson:"stats,omitempty" json:"stats,omitempty"`               // mean, std, min, max, median
	RollingFeatures []string `bson:"rolling_features,omitempty" json:"rolling_features,omitempty"` // Which features, empty = close only
}

// TargetConfig defines target variable generation
type TargetConfig struct {
	Enabled            bool       `bson:"enabled" json:"enabled"`
	Type               TargetType `bson:"type" json:"type"`
	LookaheadPeriods   []int      `bson:"lookahead_periods,omitempty" json:"lookahead_periods,omitempty"` // e.g., [1, 5, 10]
	ClassificationBins []float64  `bson:"classification_bins,omitempty" json:"classification_bins,omitempty"` // For multi-class: [-0.02, -0.01, 0.01, 0.02]
}

// PreprocessConfig defines preprocessing options
type PreprocessConfig struct {
	Normalization   NormalizationType `bson:"normalization" json:"normalization"`
	NaNHandling     NaNHandlingType   `bson:"nan_handling" json:"nan_handling"`
	RemoveNaNRows   bool              `bson:"remove_nan_rows" json:"remove_nan_rows"`
	ClipOutliers    bool              `bson:"clip_outliers" json:"clip_outliers"`
	OutlierStdDev   float64           `bson:"outlier_stddev,omitempty" json:"outlier_stddev,omitempty"` // Clip at N std devs
	InfHandling     string            `bson:"inf_handling,omitempty" json:"inf_handling,omitempty"`     // drop, replace_nan, clip
}

// SplitConfig defines train/validation/test split
type SplitConfig struct {
	Enabled         bool    `bson:"enabled" json:"enabled"`
	TrainRatio      float64 `bson:"train_ratio" json:"train_ratio"`
	ValidationRatio float64 `bson:"validation_ratio" json:"validation_ratio"`
	TestRatio       float64 `bson:"test_ratio" json:"test_ratio"`
	TimeBased       bool    `bson:"time_based" json:"time_based"` // True = chronological split (no look-ahead bias)
	Shuffle         bool    `bson:"shuffle" json:"shuffle"`       // Only if TimeBased is false
}

// SequenceConfig defines sequence generation for RNN/LSTM/Transformer
type SequenceConfig struct {
	Enabled       bool `bson:"enabled" json:"enabled"`
	Length        int  `bson:"length" json:"length"`               // Sequence length (e.g., 60 candles)
	Stride        int  `bson:"stride" json:"stride"`               // Step between sequences (1 = sliding window)
	IncludeTarget bool `bson:"include_target" json:"include_target"` // Include target in sequence output
}

// MLExportJob represents a background ML export job
type MLExportJob struct {
	ID     primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Status MLExportStatus       `bson:"status" json:"status"`
	JobIDs []primitive.ObjectID `bson:"job_ids" json:"job_ids"` // Source data collection jobs
	Config MLExportConfig       `bson:"config" json:"config"`

	// Progress tracking
	Progress         float64 `bson:"progress" json:"progress"`                   // 0-100
	TotalRecords     int64   `bson:"total_records" json:"total_records"`
	ProcessedRecords int64   `bson:"processed_records" json:"processed_records"`
	CurrentPhase     string  `bson:"current_phase,omitempty" json:"current_phase,omitempty"` // loading, features, preprocessing, writing

	// Results
	OutputPath    string   `bson:"output_path,omitempty" json:"output_path,omitempty"`
	OutputFiles   []string `bson:"output_files,omitempty" json:"output_files,omitempty"` // For split outputs
	FileSizeBytes int64    `bson:"file_size_bytes" json:"file_size_bytes"`
	FeatureCount  int      `bson:"feature_count" json:"feature_count"`
	ColumnNames   []string `bson:"column_names,omitempty" json:"column_names,omitempty"`
	RowCount      int64    `bson:"row_count" json:"row_count"`

	// Metadata for reproducibility
	Metadata MLExportMetadata `bson:"metadata" json:"metadata"`

	// Timestamps
	StartedAt   *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	ExpiresAt   *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"` // Auto-cleanup time
	LastError   string     `bson:"last_error,omitempty" json:"last_error,omitempty"`
	Errors      []string   `bson:"errors,omitempty" json:"errors,omitempty"`
	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
}

// MLExportMetadata stores metadata for reproducibility
type MLExportMetadata struct {
	Version             string                  `bson:"version" json:"version"`
	ExportedAt          time.Time               `bson:"exported_at" json:"exported_at"`
	DataRange           DataRange               `bson:"data_range" json:"data_range"`
	SourceJobs          []SourceJobInfo         `bson:"source_jobs" json:"source_jobs"`
	FeatureSchema       []FeatureSchema         `bson:"feature_schema" json:"feature_schema"`
	NormalizationParams map[string]NormParams   `bson:"normalization_params,omitempty" json:"normalization_params,omitempty"`
	SplitInfo           *SplitInfo              `bson:"split_info,omitempty" json:"split_info,omitempty"`
	SequenceInfo        *SequenceInfo           `bson:"sequence_info,omitempty" json:"sequence_info,omitempty"`
}

// DataRange describes the time range of exported data
type DataRange struct {
	StartTime  time.Time `bson:"start_time" json:"start_time"`
	EndTime    time.Time `bson:"end_time" json:"end_time"`
	TotalBars  int64     `bson:"total_bars" json:"total_bars"`
	Symbols    []string  `bson:"symbols" json:"symbols"`
	Timeframes []string  `bson:"timeframes" json:"timeframes"`
	Exchanges  []string  `bson:"exchanges" json:"exchanges"`
}

// SourceJobInfo describes a source data collection job
type SourceJobInfo struct {
	JobID      primitive.ObjectID `bson:"job_id" json:"job_id"`
	ExchangeID string             `bson:"exchange_id" json:"exchange_id"`
	Symbol     string             `bson:"symbol" json:"symbol"`
	Timeframe  string             `bson:"timeframe" json:"timeframe"`
	BarCount   int64              `bson:"bar_count" json:"bar_count"`
	StartTime  time.Time          `bson:"start_time" json:"start_time"`
	EndTime    time.Time          `bson:"end_time" json:"end_time"`
}

// FeatureSchema describes a feature column
type FeatureSchema struct {
	Name        string  `bson:"name" json:"name"`
	Type        string  `bson:"type" json:"type"` // float64, int64, bool, string
	Source      string  `bson:"source" json:"source"` // ohlcv, indicator, price_feature, lagged, rolling, temporal, cross, target
	Description string  `bson:"description,omitempty" json:"description,omitempty"`
	NaNCount    int64   `bson:"nan_count" json:"nan_count"`
	NaNPercent  float64 `bson:"nan_percent" json:"nan_percent"`
	Min         float64 `bson:"min,omitempty" json:"min,omitempty"`
	Max         float64 `bson:"max,omitempty" json:"max,omitempty"`
	Mean        float64 `bson:"mean,omitempty" json:"mean,omitempty"`
	Std         float64 `bson:"std,omitempty" json:"std,omitempty"`
}

// NormParams stores normalization parameters for a feature
type NormParams struct {
	Method string  `bson:"method" json:"method"`
	Mean   float64 `bson:"mean" json:"mean"`
	Std    float64 `bson:"std" json:"std"`
	Min    float64 `bson:"min" json:"min"`
	Max    float64 `bson:"max" json:"max"`
	Median float64 `bson:"median,omitempty" json:"median,omitempty"`
	IQR    float64 `bson:"iqr,omitempty" json:"iqr,omitempty"` // For robust scaling
}

// SplitInfo describes train/val/test split details
type SplitInfo struct {
	TrainStart time.Time `bson:"train_start" json:"train_start"`
	TrainEnd   time.Time `bson:"train_end" json:"train_end"`
	TrainRows  int64     `bson:"train_rows" json:"train_rows"`
	ValStart   time.Time `bson:"val_start,omitempty" json:"val_start,omitempty"`
	ValEnd     time.Time `bson:"val_end,omitempty" json:"val_end,omitempty"`
	ValRows    int64     `bson:"val_rows" json:"val_rows"`
	TestStart  time.Time `bson:"test_start,omitempty" json:"test_start,omitempty"`
	TestEnd    time.Time `bson:"test_end,omitempty" json:"test_end,omitempty"`
	TestRows   int64     `bson:"test_rows" json:"test_rows"`
}

// SequenceInfo describes sequence generation details
type SequenceInfo struct {
	Length         int   `bson:"length" json:"length"`
	Stride         int   `bson:"stride" json:"stride"`
	TotalSequences int64 `bson:"total_sequences" json:"total_sequences"`
	TrainSequences int64 `bson:"train_sequences,omitempty" json:"train_sequences,omitempty"`
	ValSequences   int64 `bson:"val_sequences,omitempty" json:"val_sequences,omitempty"`
	TestSequences  int64 `bson:"test_sequences,omitempty" json:"test_sequences,omitempty"`
}

// MLDataset represents a combined dataset from multiple jobs
type MLDataset struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	JobIDs      []primitive.ObjectID `bson:"job_ids" json:"job_ids"`
	ExportJobID primitive.ObjectID   `bson:"export_job_id,omitempty" json:"export_job_id,omitempty"`
	Config      MLExportConfig       `bson:"config" json:"config"`
	Metadata    MLExportMetadata     `bson:"metadata" json:"metadata"`
	OutputPath  string               `bson:"output_path,omitempty" json:"output_path,omitempty"`
	Status      MLExportStatus       `bson:"status" json:"status"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

// FeatureMatrix represents computed features ready for export
type FeatureMatrix struct {
	Columns     []string             `json:"columns"`
	Data        [][]float64          `json:"data"`
	Timestamps  []int64              `json:"timestamps"`
	RowCount    int                  `json:"row_count"`
	ColumnCount int                  `json:"column_count"`
	Schema      []FeatureSchema      `json:"schema"`
	SplitLabels []string             `json:"split_labels,omitempty"` // train, validation, test per row
	Sequences   [][][]float64        `json:"sequences,omitempty"`    // For sequence output
}

// DefaultMLExportConfig returns a default configuration
func DefaultMLExportConfig() MLExportConfig {
	return MLExportConfig{
		Format: MLExportFormatCSV,
		Features: FeatureConfig{
			IncludeOHLCV:         true,
			IncludeVolume:        true,
			IncludeTimestamp:     true,
			IncludeAllIndicators: false,
			IndicatorCategories:  []string{"trend", "momentum"},
			PriceFeatures:        []string{"returns", "log_returns", "volatility"},
			LaggedFeatures: LagConfig{
				Enabled:    true,
				LagPeriods: []int{1, 5, 10},
			},
			RollingFeatures: RollingConfig{
				Enabled: true,
				Windows: []int{5, 10, 20},
				Stats:   []string{"mean", "std"},
			},
			TemporalFeatures: []string{"hour_sin", "hour_cos", "day_of_week"},
		},
		Target: TargetConfig{
			Enabled:          true,
			Type:             TargetTypeFutureReturns,
			LookaheadPeriods: []int{1, 5},
		},
		Preprocessing: PreprocessConfig{
			Normalization: NormalizationZScore,
			NaNHandling:   NaNHandlingForwardFill,
			RemoveNaNRows: true,
		},
		Split: SplitConfig{
			Enabled:         true,
			TrainRatio:      0.7,
			ValidationRatio: 0.15,
			TestRatio:       0.15,
			TimeBased:       true,
		},
		Sequence: SequenceConfig{
			Enabled: false,
			Length:  60,
			Stride:  1,
		},
	}
}

// GetBuiltinPresets returns built-in export presets
func GetBuiltinPresets() []MLExportConfig {
	return []MLExportConfig{
		{
			Name:        "basic",
			Description: "Basic OHLCV with returns and volume",
			IsPreset:    true,
			Format:      MLExportFormatCSV,
			Features: FeatureConfig{
				IncludeOHLCV:     true,
				IncludeVolume:    true,
				IncludeTimestamp: true,
				PriceFeatures:    []string{"returns", "log_returns"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureReturns,
				LookaheadPeriods: []int{1},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationNone,
				NaNHandling:   NaNHandlingDrop,
			},
		},
		{
			Name:        "technical",
			Description: "All technical indicators, no lag features",
			IsPreset:    true,
			Format:      MLExportFormatCSV,
			Features: FeatureConfig{
				IncludeOHLCV:         true,
				IncludeVolume:        true,
				IncludeTimestamp:     true,
				IncludeAllIndicators: true,
				PriceFeatures:        []string{"returns", "log_returns", "volatility", "gaps"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureReturns,
				LookaheadPeriods: []int{1, 5, 10},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationZScore,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
			},
		},
		{
			Name:        "full_features",
			Description: "Everything enabled - maximum feature set",
			IsPreset:    true,
			Format:      MLExportFormatParquet,
			Features: FeatureConfig{
				IncludeOHLCV:         true,
				IncludeVolume:        true,
				IncludeTimestamp:     true,
				IncludeAllIndicators: true,
				PriceFeatures:        []string{"returns", "log_returns", "volatility", "price_change", "gaps", "body_ratio", "range_pct"},
				LaggedFeatures: LagConfig{
					Enabled:    true,
					LagPeriods: []int{1, 2, 3, 5, 10, 20},
				},
				RollingFeatures: RollingConfig{
					Enabled: true,
					Windows: []int{5, 10, 20, 50},
					Stats:   []string{"mean", "std", "min", "max"},
				},
				TemporalFeatures: []string{"hour", "hour_sin", "hour_cos", "day_of_week", "dow_sin", "dow_cos", "month", "is_weekend"},
				CrossFeatures:    []string{"bb_position", "price_vs_sma20", "price_vs_sma50", "ma_crossover"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureReturns,
				LookaheadPeriods: []int{1, 5, 10, 20},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationZScore,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
				ClipOutliers:  true,
				OutlierStdDev: 5.0,
			},
			Split: SplitConfig{
				Enabled:         true,
				TrainRatio:      0.7,
				ValidationRatio: 0.15,
				TestRatio:       0.15,
				TimeBased:       true,
			},
		},
		{
			Name:        "lstm_ready",
			Description: "Sequences for LSTM/Transformer models",
			IsPreset:    true,
			Format:      MLExportFormatNumpy,
			Features: FeatureConfig{
				IncludeOHLCV:         true,
				IncludeVolume:        true,
				IncludeTimestamp:     false, // Not needed in sequences
				IncludeAllIndicators: true,
				PriceFeatures:        []string{"returns", "log_returns"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureDirection,
				LookaheadPeriods: []int{1},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationMinMax,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
			},
			Split: SplitConfig{
				Enabled:         true,
				TrainRatio:      0.7,
				ValidationRatio: 0.15,
				TestRatio:       0.15,
				TimeBased:       true,
			},
			Sequence: SequenceConfig{
				Enabled:       true,
				Length:        60,
				Stride:        1,
				IncludeTarget: true,
			},
		},
		{
			Name:        "classification",
			Description: "Multi-class classification with direction bins",
			IsPreset:    true,
			Format:      MLExportFormatCSV,
			Features: FeatureConfig{
				IncludeOHLCV:        true,
				IncludeVolume:       true,
				IndicatorCategories: []string{"momentum", "volatility"},
				PriceFeatures:       []string{"returns", "volatility"},
				LaggedFeatures: LagConfig{
					Enabled:    true,
					LagPeriods: []int{1, 5, 10},
				},
			},
			Target: TargetConfig{
				Enabled:            true,
				Type:               TargetTypeFutureClass,
				LookaheadPeriods:   []int{5},
				ClassificationBins: []float64{-0.02, -0.01, 0.01, 0.02}, // Strong down, down, neutral, up, strong up
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationZScore,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
			},
			Split: SplitConfig{
				Enabled:         true,
				TrainRatio:      0.7,
				ValidationRatio: 0.15,
				TestRatio:       0.15,
				TimeBased:       true,
			},
		},
		{
			Name:        "scalping",
			Description: "Short-term features for scalping strategies",
			IsPreset:    true,
			Format:      MLExportFormatCSV,
			Features: FeatureConfig{
				IncludeOHLCV:        true,
				IncludeVolume:       true,
				IndicatorCategories: []string{"momentum"},
				SpecificIndicators:  []string{"rsi14", "stoch_k", "stoch_d", "macd", "macd_signal", "bb_upper", "bb_lower"},
				PriceFeatures:       []string{"returns", "volatility", "gaps"},
				LaggedFeatures: LagConfig{
					Enabled:    true,
					LagPeriods: []int{1, 2, 3, 5},
				},
				RollingFeatures: RollingConfig{
					Enabled: true,
					Windows: []int{3, 5, 10},
					Stats:   []string{"mean", "std"},
				},
				TemporalFeatures: []string{"hour_sin", "hour_cos"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureReturns,
				LookaheadPeriods: []int{1, 3, 5},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationZScore,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
			},
		},
		{
			Name:        "swing_trading",
			Description: "Longer-term features for swing trading",
			IsPreset:    true,
			Format:      MLExportFormatParquet,
			Features: FeatureConfig{
				IncludeOHLCV:        true,
				IncludeVolume:       true,
				IndicatorCategories: []string{"trend", "volume"},
				SpecificIndicators:  []string{"sma20", "sma50", "sma200", "adx", "obv", "atr"},
				PriceFeatures:       []string{"returns", "log_returns"},
				LaggedFeatures: LagConfig{
					Enabled:    true,
					LagPeriods: []int{1, 5, 10, 20, 50},
				},
				RollingFeatures: RollingConfig{
					Enabled: true,
					Windows: []int{10, 20, 50},
					Stats:   []string{"mean", "std", "min", "max"},
				},
				CrossFeatures: []string{"price_vs_sma20", "price_vs_sma50", "ma_crossover"},
			},
			Target: TargetConfig{
				Enabled:          true,
				Type:             TargetTypeFutureReturns,
				LookaheadPeriods: []int{5, 10, 20},
			},
			Preprocessing: PreprocessConfig{
				Normalization: NormalizationZScore,
				NaNHandling:   NaNHandlingForwardFill,
				RemoveNaNRows: true,
			},
			Split: SplitConfig{
				Enabled:         true,
				TrainRatio:      0.7,
				ValidationRatio: 0.15,
				TestRatio:       0.15,
				TimeBased:       true,
			},
		},
	}
}
