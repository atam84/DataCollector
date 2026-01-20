import { InformationCircleIcon } from '@heroicons/react/24/outline'

/**
 * IndicatorsInfo Component
 *
 * Comprehensive documentation page for all 29 technical indicators.
 * Organized by category with descriptions, use cases, and parameter explanations.
 */
function IndicatorsInfo() {
  const indicators = {
    trend: [
      {
        name: 'SMA (Simple Moving Average)',
        description: 'The average price over a specified period. It smooths out price data to identify trends by filtering out noise.',
        useCase: 'Trend identification, support/resistance levels, golden/death cross strategies.',
        parameters: [
          { name: 'Periods', description: 'Common values: 20 (short-term), 50 (medium-term), 200 (long-term)' }
        ],
        calculation: 'Sum of closing prices ÷ Number of periods'
      },
      {
        name: 'EMA (Exponential Moving Average)',
        description: 'Similar to SMA but gives more weight to recent prices, making it more responsive to new information.',
        useCase: 'Faster trend detection, MACD calculation, dynamic support/resistance.',
        parameters: [
          { name: 'Periods', description: 'Common values: 12, 26 (MACD), 50 (trend)' }
        ],
        calculation: 'Weighted average with exponentially decreasing weights'
      },
      {
        name: 'DEMA (Double Exponential Moving Average)',
        description: 'Reduces lag by applying EMA twice, providing faster response to price changes.',
        useCase: 'Reducing lag in trend signals, faster entry/exit points.',
        parameters: [
          { name: 'Period', description: 'Typically 20-30 periods' }
        ]
      },
      {
        name: 'TEMA (Triple Exponential Moving Average)',
        description: 'Further reduces lag by applying EMA three times, offering the most responsive moving average.',
        useCase: 'Ultra-fast trend detection, scalping strategies.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      },
      {
        name: 'WMA (Weighted Moving Average)',
        description: 'Assigns linearly decreasing weights to older prices, more responsive than SMA but less than EMA.',
        useCase: 'Balance between responsiveness and stability.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      },
      {
        name: 'HMA (Hull Moving Average)',
        description: 'Combines WMA calculations to reduce lag while maintaining smoothness.',
        useCase: 'Low-lag trend identification, reducing false signals.',
        parameters: [
          { name: 'Period', description: 'Typically 9 periods for responsiveness' }
        ],
        calculation: 'WMA(2×WMA(n/2) - WMA(n)), using WMA(√n)'
      },
      {
        name: 'VWMA (Volume Weighted Moving Average)',
        description: 'Weights prices by volume, giving more importance to high-volume periods.',
        useCase: 'Identifying price levels with strong volume support.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      },
      {
        name: 'Ichimoku Cloud',
        description: 'Multi-component indicator providing support/resistance levels, trend direction, and momentum.',
        useCase: 'Complete trend analysis system, cloud provides support/resistance zones.',
        parameters: [
          { name: 'Tenkan', description: '(9) Conversion Line - short-term trend' },
          { name: 'Kijun', description: '(26) Base Line - medium-term trend' },
          { name: 'Senkou B', description: '(52) Leading Span B - long-term trend' }
        ]
      },
      {
        name: 'ADX (Average Directional Index)',
        description: 'Measures trend strength on a scale of 0-100, regardless of direction.',
        useCase: 'Determining if market is trending or ranging. Above 25 indicates trending.',
        parameters: [
          { name: 'Period', description: 'Standard is 14 periods' }
        ],
        calculation: 'Smoothed averages of +DI and -DI differences'
      },
      {
        name: 'SuperTrend',
        description: 'Trend-following indicator that provides buy/sell signals based on ATR and price position.',
        useCase: 'Clear trend following with dynamic stop-loss levels.',
        parameters: [
          { name: 'Period', description: 'ATR period, typically 10' },
          { name: 'Multiplier', description: 'ATR multiplier, typically 3.0' }
        ]
      }
    ],
    momentum: [
      {
        name: 'RSI (Relative Strength Index)',
        description: 'Measures speed and magnitude of price changes on a 0-100 scale.',
        useCase: 'Overbought (>70) and oversold (<30) conditions, divergence analysis.',
        parameters: [
          { name: 'Periods', description: 'Common: 6 (fast), 14 (standard), 24 (slow)' }
        ],
        calculation: '100 - (100 / (1 + RS)), where RS = Avg Gain / Avg Loss'
      },
      {
        name: 'Stochastic Oscillator',
        description: 'Compares closing price to price range over a period, showing momentum.',
        useCase: 'Overbought/oversold identification, divergence signals.',
        parameters: [
          { name: 'K Period', description: 'Typically 14 - lookback period' },
          { name: 'D Period', description: 'Typically 3 - smoothing of %K' },
          { name: 'Smooth', description: 'Typically 3 - additional smoothing' }
        ]
      },
      {
        name: 'MACD (Moving Average Convergence Divergence)',
        description: 'Shows relationship between two EMAs, with signal line and histogram.',
        useCase: 'Trend changes, momentum shifts, divergence analysis.',
        parameters: [
          { name: 'Fast', description: '12-period EMA' },
          { name: 'Slow', description: '26-period EMA' },
          { name: 'Signal', description: '9-period EMA of MACD line' }
        ],
        calculation: 'MACD Line = EMA(12) - EMA(26), Signal = EMA(9) of MACD'
      },
      {
        name: 'ROC (Rate of Change)',
        description: 'Measures percentage change in price over a specified period.',
        useCase: 'Momentum measurement, divergence detection.',
        parameters: [
          { name: 'Period', description: 'Typically 12 periods' }
        ],
        calculation: '((Current Price - Price n periods ago) / Price n periods ago) × 100'
      },
      {
        name: 'CCI (Commodity Channel Index)',
        description: 'Measures deviation from average price, identifying cyclical trends.',
        useCase: 'Overbought (>100), oversold (<-100), trend identification.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      },
      {
        name: 'Williams %R',
        description: 'Momentum indicator comparing closing price to high-low range.',
        useCase: 'Overbought (-20 to 0), oversold (-80 to -100) conditions.',
        parameters: [
          { name: 'Period', description: 'Typically 14 periods' }
        ],
        calculation: '((Highest High - Close) / (Highest High - Lowest Low)) × -100'
      },
      {
        name: 'Momentum',
        description: 'Simple rate of change showing how much price has changed over a period.',
        useCase: 'Measuring strength of price movements.',
        parameters: [
          { name: 'Period', description: 'Typically 10 periods' }
        ],
        calculation: 'Current Price - Price n periods ago'
      }
    ],
    volatility: [
      {
        name: 'Bollinger Bands',
        description: 'Price channel consisting of middle (SMA), upper, and lower bands based on standard deviation.',
        useCase: 'Volatility measurement, price extremes, breakout detection.',
        parameters: [
          { name: 'Period', description: 'Typically 20 for SMA' },
          { name: 'Std Dev', description: 'Typically 2.0 standard deviations' }
        ],
        calculation: 'Middle = SMA(20), Upper/Lower = Middle ± (2 × StdDev)'
      },
      {
        name: 'ATR (Average True Range)',
        description: 'Measures market volatility by calculating average of true ranges.',
        useCase: 'Volatility assessment, stop-loss placement, position sizing.',
        parameters: [
          { name: 'Period', description: 'Typically 14 periods' }
        ],
        calculation: 'Average of True Range = max(High-Low, |High-PrevClose|, |Low-PrevClose|)'
      },
      {
        name: 'Keltner Channels',
        description: 'Volatility-based envelopes using EMA and ATR.',
        useCase: 'Trend identification, breakout detection.',
        parameters: [
          { name: 'EMA Period', description: 'Typically 20' },
          { name: 'ATR Period', description: 'Typically 10' },
          { name: 'Multiplier', description: 'Typically 2.0' }
        ],
        calculation: 'Middle = EMA, Upper/Lower = EMA ± (Multiplier × ATR)'
      },
      {
        name: 'Donchian Channels',
        description: 'Price channel showing highest high and lowest low over a period.',
        useCase: 'Breakout trading, trend following.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ],
        calculation: 'Upper = Highest High(n), Lower = Lowest Low(n), Middle = Average'
      },
      {
        name: 'Standard Deviation',
        description: 'Statistical measure of price volatility and dispersion.',
        useCase: 'Volatility measurement, risk assessment.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      }
    ],
    volume: [
      {
        name: 'OBV (On-Balance Volume)',
        description: 'Cumulative volume indicator adding volume on up days and subtracting on down days.',
        useCase: 'Confirming trends, detecting divergences, volume flow analysis.',
        parameters: [],
        calculation: 'Running total: Add volume if close > prev close, subtract if close < prev close'
      },
      {
        name: 'VWAP (Volume Weighted Average Price)',
        description: 'Average price weighted by volume, often resets daily.',
        useCase: 'Intraday benchmarking, institutional trader reference.',
        parameters: [],
        calculation: 'Cumulative(Price × Volume) / Cumulative(Volume)'
      },
      {
        name: 'MFI (Money Flow Index)',
        description: 'Volume-weighted RSI measuring buying and selling pressure.',
        useCase: 'Overbought (>80), oversold (<20), divergence analysis.',
        parameters: [
          { name: 'Period', description: 'Typically 14 periods' }
        ],
        calculation: '100 - (100 / (1 + Money Flow Ratio))'
      },
      {
        name: 'CMF (Chaikin Money Flow)',
        description: 'Measures money flow volume over a specific period.',
        useCase: 'Accumulation/distribution, trend confirmation.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ],
        calculation: 'Sum of Money Flow Volume / Sum of Volume'
      },
      {
        name: 'Volume SMA',
        description: 'Simple moving average of volume.',
        useCase: 'Identifying high/low volume periods, volume trend analysis.',
        parameters: [
          { name: 'Period', description: 'Typically 20 periods' }
        ]
      }
    ]
  }

  const IndicatorCard = ({ indicator }) => (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition">
      <h3 className="text-lg font-bold text-gray-900 mb-2">{indicator.name}</h3>
      <p className="text-sm text-gray-700 mb-3">{indicator.description}</p>

      <div className="mb-3">
        <span className="text-xs font-semibold text-blue-600 uppercase">Use Case:</span>
        <p className="text-sm text-gray-600 mt-1">{indicator.useCase}</p>
      </div>

      {indicator.parameters && indicator.parameters.length > 0 && (
        <div className="mb-3">
          <span className="text-xs font-semibold text-purple-600 uppercase">Parameters:</span>
          <ul className="mt-1 space-y-1">
            {indicator.parameters.map((param, idx) => (
              <li key={idx} className="text-sm text-gray-600">
                <span className="font-medium">{param.name}:</span> {param.description}
              </li>
            ))}
          </ul>
        </div>
      )}

      {indicator.calculation && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <span className="text-xs font-semibold text-green-600 uppercase">Calculation:</span>
          <p className="text-xs text-gray-500 mt-1 font-mono">{indicator.calculation}</p>
        </div>
      )}
    </div>
  )

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <div className="flex items-center space-x-3 mb-4">
          <InformationCircleIcon className="w-8 h-8 text-blue-500" />
          <h1 className="text-3xl font-bold text-gray-900">Technical Indicators Reference</h1>
        </div>
        <p className="text-gray-600">
          Comprehensive guide to all 29 technical indicators available in the DataCollector platform.
          Each indicator can be configured at the connector or job level to customize your data collection strategy.
        </p>
      </div>

      {/* Trend Indicators */}
      <div className="mb-10">
        <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center">
          <span className="mr-2">📈</span> Trend Indicators ({indicators.trend.length})
        </h2>
        <p className="text-gray-600 mb-4">
          Trend indicators help identify the direction and strength of market trends. They smooth price data
          to filter out noise and reveal underlying momentum.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {indicators.trend.map((indicator, idx) => (
            <IndicatorCard key={idx} indicator={indicator} />
          ))}
        </div>
      </div>

      {/* Momentum Indicators */}
      <div className="mb-10">
        <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center">
          <span className="mr-2">⚡</span> Momentum Indicators ({indicators.momentum.length})
        </h2>
        <p className="text-gray-600 mb-4">
          Momentum indicators measure the speed and strength of price movements. They help identify
          overbought/oversold conditions and potential trend reversals.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {indicators.momentum.map((indicator, idx) => (
            <IndicatorCard key={idx} indicator={indicator} />
          ))}
        </div>
      </div>

      {/* Volatility Indicators */}
      <div className="mb-10">
        <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center">
          <span className="mr-2">📊</span> Volatility Indicators ({indicators.volatility.length})
        </h2>
        <p className="text-gray-600 mb-4">
          Volatility indicators measure price fluctuation and market uncertainty. They help with risk
          management, stop-loss placement, and identifying breakout opportunities.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {indicators.volatility.map((indicator, idx) => (
            <IndicatorCard key={idx} indicator={indicator} />
          ))}
        </div>
      </div>

      {/* Volume Indicators */}
      <div className="mb-10">
        <h2 className="text-2xl font-bold text-gray-900 mb-4 flex items-center">
          <span className="mr-2">📦</span> Volume Indicators ({indicators.volume.length})
        </h2>
        <p className="text-gray-600 mb-4">
          Volume indicators analyze trading volume to confirm trends and identify potential reversals.
          They provide insight into the strength behind price movements.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {indicators.volume.map((indicator, idx) => (
            <IndicatorCard key={idx} indicator={indicator} />
          ))}
        </div>
      </div>

      {/* Quick Reference */}
      <div className="bg-blue-50 rounded-lg p-6 border-2 border-blue-200">
        <h3 className="text-lg font-bold text-blue-900 mb-3">💡 Quick Tips</h3>
        <ul className="space-y-2 text-sm text-blue-800">
          <li>• <strong>Start Simple:</strong> Begin with basic indicators like SMA, EMA, RSI, and MACD before exploring advanced ones.</li>
          <li>• <strong>Configuration Levels:</strong> Set defaults at connector level, override specific indicators at job level.</li>
          <li>• <strong>Performance:</strong> More indicators = longer calculation time. Only enable what you need.</li>
          <li>• <strong>Recalculation:</strong> After changing config, use the recalculate button to apply changes to existing data.</li>
          <li>• <strong>Combinations:</strong> Use multiple indicator types (trend + momentum + volume) for better confirmation.</li>
        </ul>
      </div>
    </div>
  )
}

export default IndicatorsInfo
