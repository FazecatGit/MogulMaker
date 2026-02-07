import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

interface PriceBar {
  date?: string;
  timestamp?: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
  [key: string]: any;
}

interface PriceChartProps {
  data: PriceBar[];
  title: string;
  daysLabel?: string;
  showVolume?: boolean;
  showHigh?: boolean;
  showLow?: boolean;
  height?: number;
}

export default function PriceChart({
  data,
  title,
  daysLabel,
  showVolume = false,
  showHigh = true,
  showLow = true,
  height = 400,
}: PriceChartProps) {
  const formattedData = data.map((bar) => ({
    ...bar,
    date: bar.date || (bar.timestamp ? new Date(bar.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''),
    fullDate: bar.timestamp ? new Date(bar.timestamp * 1000).toLocaleDateString() : bar.date,
  }));

  const interval = Math.floor(formattedData.length / 10);

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

      {/* Chart Container */}
      <div className="bg-slate-900/50 rounded p-4 overflow-x-auto mb-6">
        <ResponsiveContainer width="100%" height={height}>
          <LineChart
            data={formattedData}
            margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
          >
            <defs>
              <linearGradient id="closeGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#22c55e" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
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
              stroke="#64748b"
              tick={{ fontSize: 12 }}
              domain={['dataMin - 5', 'dataMax + 5']}
              label={{ value: 'Price ($)', angle: -90, position: 'insideLeft' }}
            />
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
                if (name === 'volume') {
                  return [`${(value / 1000000).toFixed(1)}M`, 'Volume'];
                }
                if (['rsi', 'atr', 'sma_20'].includes(name)) {
                  return [value.toFixed(2), name.toUpperCase()];
                }
                return [value, name];
              }}
              cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
            />
            <Legend wrapperStyle={{ paddingTop: '20px' }} iconType="line" />
            <Line
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
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
