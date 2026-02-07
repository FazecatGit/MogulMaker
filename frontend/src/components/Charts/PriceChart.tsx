import React from 'react';
import {
  ComposedChart,
  Line,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';

interface PriceBar {
  date?: string;
  timestamp?: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
  rsi?: number;
  atr?: number;
  sma_20?: number;
  [key: string]: any;
}

interface PriceChartProps {
  data: PriceBar[];
  title: string;
  daysLabel?: string;
  showVolume?: boolean;
  showHigh?: boolean;
  showLow?: boolean;
  showRSI?: boolean;
  showATR?: boolean;
  showSMA?: boolean;
  chartType?: 'line' | 'candlestick';
  height?: number;
}

export default function PriceChart({
  data,
  title,
  daysLabel,
  showVolume = false,
  showHigh = true,
  showLow = true,
  showRSI = false,
  showATR = false,
  showSMA = false,
  chartType = 'line',
  height = 400,
}: PriceChartProps) {
  const formattedData = data.map((bar) => ({
    ...bar,
    date: bar.date || (bar.timestamp ? new Date(bar.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''),
    fullDate: bar.timestamp ? new Date(bar.timestamp * 1000).toLocaleDateString() : bar.date,
    // Scale volume to fit on chart (divide by 1M for readability)
    volumeScaled: bar.volume ? bar.volume / 1000000 : 0,
  }));

  const interval = Math.floor(formattedData.length / 10);

  // Calculate dynamic height based on number of sub-charts
  const numCharts = 1 + (showVolume ? 1 : 0) + (showRSI ? 1 : 0) + (showATR ? 1 : 0);
  const totalHeight = height * numCharts * 0.7; // Scale height for multiple charts

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
      <div className="flex justify-between items-center mb-2">
        <h3 className="text-xl font-bold">{title}</h3>
        {daysLabel && (
          <div className="text-sm text-slate-400">
            {daysLabel}
          </div>
        )}
      </div>

      {/* Main Price Chart */}
      <div className="bg-slate-900/50 rounded p-4 overflow-x-auto mb-4">
        <ResponsiveContainer width="100%" height={height}>
          <ComposedChart
            data={formattedData}
            margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
          >
            <defs>
              <linearGradient id="closeGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#22c55e" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="volumeGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.6} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.1} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis
              dataKey="date"
              stroke="#64748b"
              tick={{ fontSize: 12 }}
              interval={interval > 0 ? interval : 0}
            />
            <YAxis
              yAxisId="price"
              stroke="#64748b"
              tick={{ fontSize: 12 }}
              domain={['dataMin - 5', 'dataMax + 5']}
              label={{ value: 'Price ($)', angle: -90, position: 'insideLeft', style: { fill: '#64748b' } }}
            />
            {showVolume && (
              <YAxis
                yAxisId="volume"
                orientation="right"
                stroke="#3b82f6"
                tick={{ fontSize: 12 }}
                label={{ value: 'Volume (M)', angle: 90, position: 'insideRight', style: { fill: '#3b82f6' } }}
              />
            )}
            <Tooltip
              contentStyle={{
                backgroundColor: '#1e293b',
                border: '1px solid #475569',
                borderRadius: '6px',
              }}
              labelStyle={{ color: '#e2e8f0' }}
              formatter={(value: any, name: string | undefined) => {
                if (!name) return [value, name];
                if (['open', 'high', 'low', 'close'].includes(name)) {
                  return [`$${value.toFixed(2)}`, name.toUpperCase()];
                }
                if (name === 'volumeScaled') {
                  return [`${value.toFixed(1)}M`, 'Volume'];
                }
                if (name === 'rsi') {
                  return [value.toFixed(2), 'RSI'];
                }
                if (name === 'atr') {
                  return [value.toFixed(2), 'ATR'];
                }
                if (name === 'sma_20') {
                  return [`$${value.toFixed(2)}`, 'SMA 20'];
                }
                return [value, name];
              }}
              cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
            />
            <Legend wrapperStyle={{ paddingTop: '10px' }} iconType="line" />

            {/* Volume Bars (if enabled) */}
            {showVolume && (
              <Bar
                yAxisId="volume"
                dataKey="volumeScaled"
                fill="url(#volumeGradient)"
                opacity={0.4}
                name="Volume (M)"
                isAnimationActive={false}
              />
            )}

            {/* Price Lines */}
            <Line
              yAxisId="price"
              type="monotone"
              dataKey="close"
              stroke="#22c55e"
              dot={false}
              strokeWidth={2}
              name="Close"
              isAnimationActive={false}
            />
            {showHigh && (
              <Line
                yAxisId="price"
                type="monotone"
                dataKey="high"
                stroke="#6b7280"
                dot={false}
                strokeWidth={1}
                name="High"
                strokeDasharray="5 5"
                isAnimationActive={false}
              />
            )}
            {showLow && (
              <Line
                yAxisId="price"
                type="monotone"
                dataKey="low"
                stroke="#6b7280"
                dot={false}
                strokeWidth={1}
                name="Low"
                strokeDasharray="5 5"
                isAnimationActive={false}
              />
            )}

            {/* Technical Indicators */}
            {showSMA && (
              <Line
                yAxisId="price"
                type="monotone"
                dataKey="sma_20"
                stroke="#f59e0b"
                dot={false}
                strokeWidth={2}
                name="SMA 20"
                strokeDasharray="3 3"
                isAnimationActive={false}
              />
            )}
          </ComposedChart>
        </ResponsiveContainer>
      </div>

      {/* RSI Chart (if enabled) */}
      {showRSI && (
        <div className="bg-slate-900/50 rounded p-4 overflow-x-auto mb-4">
          <ResponsiveContainer width="100%" height={150}>
            <ComposedChart
              data={formattedData}
              margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis
                dataKey="date"
                stroke="#64748b"
                tick={{ fontSize: 12 }}
                interval={interval > 0 ? interval : 0}
              />
              <YAxis
                stroke="#64748b"
                tick={{ fontSize: 12 }}
                domain={[0, 100]}
                label={{ value: 'RSI', angle: -90, position: 'insideLeft', style: { fill: '#64748b' } }}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1e293b',
                  border: '1px solid #475569',
                  borderRadius: '6px',
                }}
                labelStyle={{ color: '#e2e8f0' }}
                formatter={(value: any) => [value.toFixed(2), 'RSI']}
                cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
              />
              <Legend wrapperStyle={{ paddingTop: '10px' }} iconType="line" />
              
              {/* Overbought/Oversold Reference Lines */}
              <ReferenceLine y={70} stroke="#ef4444" strokeDasharray="3 3" label={{ value: 'Overbought (70)', position: 'right', fill: '#ef4444', fontSize: 10 }} />
              <ReferenceLine y={30} stroke="#22c55e" strokeDasharray="3 3" label={{ value: 'Oversold (30)', position: 'right', fill: '#22c55e', fontSize: 10 }} />
              
              <Line
                type="monotone"
                dataKey="rsi"
                stroke="#8b5cf6"
                dot={false}
                strokeWidth={2}
                name="RSI"
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* ATR Chart (if enabled) */}
      {showATR && (
        <div className="bg-slate-900/50 rounded p-4 overflow-x-auto">
          <ResponsiveContainer width="100%" height={150}>
            <ComposedChart
              data={formattedData}
              margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis
                dataKey="date"
                stroke="#64748b"
                tick={{ fontSize: 12 }}
                interval={interval > 0 ? interval : 0}
              />
              <YAxis
                stroke="#64748b"
                tick={{ fontSize: 12 }}
                domain={['auto', 'auto']}
                label={{ value: 'ATR', angle: -90, position: 'insideLeft', style: { fill: '#64748b' } }}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1e293b',
                  border: '1px solid #475569',
                  borderRadius: '6px',
                }}
                labelStyle={{ color: '#e2e8f0' }}
                formatter={(value: any) => [value.toFixed(2), 'ATR']}
                cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
              />
              <Legend wrapperStyle={{ paddingTop: '10px' }} iconType="line" />
              
              <Line
                type="monotone"
                dataKey="atr"
                stroke="#06b6d4"
                dot={false}
                strokeWidth={2}
                name="ATR (Volatility)"
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
