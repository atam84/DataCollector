package models

import (
	"testing"
)

func TestDefaultIndicatorConfig(t *testing.T) {
	config := DefaultIndicatorConfig()

	if config.Name != "default" {
		t.Errorf("expected name 'default', got '%s'", config.Name)
	}
	if !config.IsDefault {
		t.Error("expected IsDefault to be true")
	}
	if !config.EnableTrend {
		t.Error("expected EnableTrend to be true")
	}
	if !config.EnableMomentum {
		t.Error("expected EnableMomentum to be true")
	}
	if !config.EnableVolatility {
		t.Error("expected EnableVolatility to be true")
	}
	if !config.EnableVolume {
		t.Error("expected EnableVolume to be true")
	}

	// Validate default config should pass
	result := config.Validate()
	if !result.Valid {
		t.Errorf("default config should be valid, got errors: %v", result.Errors)
	}
}

func TestDefaultTrendConfig(t *testing.T) {
	config := DefaultTrendConfig()

	if !config.SMAEnabled {
		t.Error("expected SMAEnabled to be true")
	}
	if len(config.SMAPeriods) != 3 {
		t.Errorf("expected 3 SMA periods, got %d", len(config.SMAPeriods))
	}
	if config.DEMAPeriod != 20 {
		t.Errorf("expected DEMA period 20, got %d", config.DEMAPeriod)
	}
	if config.SuperTrendMultiplier != 3.0 {
		t.Errorf("expected SuperTrend multiplier 3.0, got %f", config.SuperTrendMultiplier)
	}
}

func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		minVal   int
		hasError bool
	}{
		{"valid period", 14, 1, false},
		{"minimum period", 1, 1, false},
		{"maximum period", 500, 1, false},
		{"below minimum", 0, 1, true},
		{"above maximum", 501, 1, true},
		{"negative period", -1, 1, true},
		{"below custom minimum", 1, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePeriod("test_field", tt.value, tt.minVal)
			if tt.hasError && err == nil {
				t.Errorf("expected error for period %d with min %d", tt.value, tt.minVal)
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error for period %d with min %d: %s", tt.value, tt.minVal, err.Message)
			}
		})
	}
}

func TestValidatePeriodList(t *testing.T) {
	tests := []struct {
		name       string
		values     []int
		errorCount int
	}{
		{"valid periods", []int{10, 20, 50}, 0},
		{"single period", []int{14}, 0},
		{"empty list", []int{}, 1},
		{"contains zero", []int{0, 14, 20}, 1},
		{"contains negative", []int{-5, 14, 20}, 1},
		{"contains too high", []int{10, 501}, 1},
		{"multiple invalid", []int{0, -1, 501}, 3},
		{"max periods", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0},
		{"too many periods", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validatePeriodList("test_field", tt.values)
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestValidateMultiplier(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		hasError bool
	}{
		{"valid multiplier", 3.0, false},
		{"minimum multiplier", 0.1, false},
		{"maximum multiplier", 10.0, false},
		{"below minimum", 0.05, true},
		{"above maximum", 10.5, true},
		{"zero", 0.0, true},
		{"negative", -1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultiplier("test_field", tt.value)
			if tt.hasError && err == nil {
				t.Errorf("expected error for multiplier %f", tt.value)
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error for multiplier %f: %s", tt.value, err.Message)
			}
		})
	}
}

func TestValidateStdDev(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		hasError bool
	}{
		{"valid stddev", 2.0, false},
		{"minimum stddev", 0.5, false},
		{"maximum stddev", 5.0, false},
		{"below minimum", 0.4, true},
		{"above maximum", 5.5, true},
		{"zero", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStdDev("test_field", tt.value)
			if tt.hasError && err == nil {
				t.Errorf("expected error for stddev %f", tt.value)
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error for stddev %f: %s", tt.value, err.Message)
			}
		})
	}
}

func TestTrendConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		config     TrendConfig
		errorCount int
	}{
		{
			"valid default config",
			DefaultTrendConfig(),
			0,
		},
		{
			"invalid SMA periods",
			TrendConfig{
				SMAEnabled: true,
				SMAPeriods: []int{0, 501},
			},
			2, // two invalid periods
		},
		{
			"invalid DEMA period",
			TrendConfig{
				DEMAEnabled: true,
				DEMAPeriod:  1, // needs at least 2
			},
			1,
		},
		{
			"invalid SuperTrend multiplier",
			TrendConfig{
				SuperTrendEnabled:    true,
				SuperTrendPeriod:     10,
				SuperTrendMultiplier: 0.05, // too low
			},
			1,
		},
		{
			"Ichimoku tenkan >= kijun",
			TrendConfig{
				IchimokuEnabled: true,
				IchimokuTenkan:  26,
				IchimokuKijun:   26, // should be less
				IchimokuSenkouB: 52,
				IchimokuDisplacement: 26,
			},
			1, // tenkan must be < kijun
		},
		{
			"disabled indicator with invalid params",
			TrendConfig{
				SMAEnabled: false, // disabled, so invalid periods shouldn't matter
				SMAPeriods: []int{0, -1},
			},
			0, // no errors since disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestMomentumConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		config     MomentumConfig
		errorCount int
	}{
		{
			"valid default config",
			DefaultMomentumConfig(),
			0,
		},
		{
			"invalid RSI periods",
			MomentumConfig{
				RSIEnabled: true,
				RSIPeriods: []int{},
			},
			1, // empty list
		},
		{
			"MACD fast >= slow",
			MomentumConfig{
				MACDEnabled: true,
				MACDFast:    26,
				MACDSlow:    12, // fast should be < slow
				MACDSignal:  9,
			},
			1,
		},
		{
			"MACD fast == slow",
			MomentumConfig{
				MACDEnabled: true,
				MACDFast:    12,
				MACDSlow:    12,
				MACDSignal:  9,
			},
			1,
		},
		{
			"valid MACD config",
			MomentumConfig{
				MACDEnabled: true,
				MACDFast:    12,
				MACDSlow:    26,
				MACDSignal:  9,
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestVolatilityConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		config     VolatilityConfig
		errorCount int
	}{
		{
			"valid default config",
			DefaultVolatilityConfig(),
			0,
		},
		{
			"invalid Bollinger stddev",
			VolatilityConfig{
				BollingerEnabled: true,
				BollingerPeriod:  20,
				BollingerStdDev:  0.3, // too low (min 0.5)
			},
			1,
		},
		{
			"invalid Bollinger period",
			VolatilityConfig{
				BollingerEnabled: true,
				BollingerPeriod:  1, // needs at least 2 for stddev
				BollingerStdDev:  2.0,
			},
			1,
		},
		{
			"invalid Keltner multiplier",
			VolatilityConfig{
				KeltnerEnabled:    true,
				KeltnerPeriod:     20,
				KeltnerATRPeriod:  10,
				KeltnerMultiplier: 15.0, // too high (max 10)
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestVolumeConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		config     VolumeConfig
		errorCount int
	}{
		{
			"valid default config",
			DefaultVolumeConfig(),
			0,
		},
		{
			"invalid MFI period",
			VolumeConfig{
				MFIEnabled: true,
				MFIPeriod:  0, // must be >= 1
			},
			1,
		},
		{
			"OBV no validation needed",
			VolumeConfig{
				OBVEnabled: true,
			},
			0, // OBV has no period param
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestFullIndicatorConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *IndicatorConfig
		isValid bool
	}{
		{
			"valid default config",
			DefaultIndicatorConfig(),
			true,
		},
		{
			"missing name",
			&IndicatorConfig{
				Name:        "",
				EnableTrend: true,
				Trend:       DefaultTrendConfig(),
			},
			false,
		},
		{
			"all categories disabled",
			&IndicatorConfig{
				Name:             "empty",
				EnableTrend:      false,
				EnableMomentum:   false,
				EnableVolatility: false,
				EnableVolume:     false,
			},
			true, // valid to have all disabled
		},
		{
			"invalid trend with enabled",
			&IndicatorConfig{
				Name:        "bad_trend",
				EnableTrend: true,
				Trend: TrendConfig{
					SMAEnabled: true,
					SMAPeriods: []int{}, // empty not allowed
				},
			},
			false,
		},
		{
			"invalid trend but disabled",
			&IndicatorConfig{
				Name:        "disabled_trend",
				EnableTrend: false, // disabled, so invalid params ok
				Trend: TrendConfig{
					SMAEnabled: true,
					SMAPeriods: []int{},
				},
			},
			true,
		},
		{
			"multiple category errors",
			&IndicatorConfig{
				Name:             "multi_error",
				EnableTrend:      true,
				EnableMomentum:   true,
				EnableVolatility: true,
				Trend: TrendConfig{
					SMAEnabled: true,
					SMAPeriods: []int{0}, // invalid
				},
				Momentum: MomentumConfig{
					MACDEnabled: true,
					MACDFast:    26,
					MACDSlow:    12, // fast >= slow
					MACDSignal:  9,
				},
				Volatility: VolatilityConfig{
					BollingerEnabled: true,
					BollingerPeriod:  20,
					BollingerStdDev:  0.1, // too low
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()
			if result.Valid != tt.isValid {
				t.Errorf("expected Valid=%v, got Valid=%v with errors: %v", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestValidationErrorFieldPrefixes(t *testing.T) {
	config := &IndicatorConfig{
		Name:             "test",
		EnableTrend:      true,
		EnableMomentum:   true,
		EnableVolatility: true,
		EnableVolume:     true,
		Trend: TrendConfig{
			SMAEnabled: true,
			SMAPeriods: []int{0},
		},
		Momentum: MomentumConfig{
			RSIEnabled: true,
			RSIPeriods: []int{-1},
		},
		Volatility: VolatilityConfig{
			ATREnabled: true,
			ATRPeriod:  0,
		},
		Volume: VolumeConfig{
			MFIEnabled: true,
			MFIPeriod:  0,
		},
	}

	result := config.Validate()

	// Check that field names have proper prefixes
	fieldPrefixes := map[string]bool{
		"trend.":      false,
		"momentum.":   false,
		"volatility.": false,
		"volume.":     false,
	}

	for _, err := range result.Errors {
		for prefix := range fieldPrefixes {
			if len(err.Field) >= len(prefix) && err.Field[:len(prefix)] == prefix {
				fieldPrefixes[prefix] = true
			}
		}
	}

	for prefix, found := range fieldPrefixes {
		if !found {
			t.Errorf("expected to find error with prefix '%s'", prefix)
		}
	}
}

func TestConfigValidationResultAddError(t *testing.T) {
	result := ConfigValidationResult{Valid: true, Errors: []ValidationError{}}

	result.addError("field1", "error1")

	if result.Valid {
		t.Error("expected Valid to be false after adding error")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Field != "field1" {
		t.Errorf("expected field 'field1', got '%s'", result.Errors[0].Field)
	}

	result.addError("field2", "error2")
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}
