'use client';

import { useState } from 'react';
import { Calendar, Play, RotateCcw, TrendingUp, BarChart3, AlertCircle } from 'lucide-react';
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
import apiClient from '@/lib/apiClient';

interface Trade {
  trade_num: number;
  entry_price: number;
  exit_price: number;
  entry_time: string;
  exit_time: string;
  pnl: number;
  return_pct: number;
  quantity: number;
}

interface BacktestResult {
  backtest_id: string;
  symbol: string;
  status: string;
  start_date: string;
  end_date: string;
  initial_capital: number;
  final_balance: number;
  total_return_pct: number;
  win_rate: number;
  total_trades: number;
  winning_trades: number;
  losing_trades: number;
  largest_win: number;
  largest_loss: number;
  created_at: number;
  historical_bars?: Array<{
    date: string;
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
    timestamp: number;
  }>;
  trades?: Trade[];
}

export default function BacktestPage() {
  const [symbol, setSymbol] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [results, setResults] = useState<BacktestResult | null>(null);

  const handleRunBacktest = async () => {
    if (!symbol.trim() || !startDate || !endDate) {
      setError('Please fill in all fields');
      return;
    }

    if (new Date(startDate) >= new Date(endDate)) {
      setError('Start date must be before end date');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        symbol: symbol.toUpperCase(),
        start_date: startDate,
        end_date: endDate,
        capital: '100000', // Default capital
      });

      const data = await apiClient.get(`/backtest?${params.toString()}`) as BacktestResult;
      console.log('Backtest result:', data);
      setResults(data);
    } catch (err: any) {
      console.error('Backtest error:', err);
      setError(err.message || 'Failed to run backtest. Please check your inputs and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleReset = () => {
    setSymbol('');
    setStartDate('');
    setEndDate('');
    setResults(null);
    setError(null);
  };

  const formatPercent = (value: number) => {
    return `${(value * 100).toFixed(2)}%`;
  };

  const formatCurrency = (value: number) => {
    return `$${value.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')}`
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">Strategy Backtest</h1>
        <p className="text-slate-400">Test your trading strategy on historical data</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Input Panel */}
        <div className="lg:col-span-1">
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-6 sticky top-4 space-y-4">
            <h2 className="text-lg font-bold text-white">Backtest Setup</h2>

            {/* Symbol Input */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Stock Symbol
              </label>
              <input
                type="text"
                placeholder="e.g., TSLA"
                value={symbol}
                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
              />
            </div>

            {/* Start Date */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Start Date
              </label>
              <div className="relative">
                <Calendar className="absolute left-3 top-3 w-4 h-4 text-slate-400" />
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>

            {/* End Date */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                End Date
              </label>
              <div className="relative">
                <Calendar className="absolute left-3 top-3 w-4 h-4 text-slate-400" />
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-3 flex gap-2 text-sm text-red-400">
                <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                {error}
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex gap-2 pt-2">
              <button
                onClick={handleRunBacktest}
                disabled={isLoading}
                className="flex-1 flex items-center justify-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 text-white px-4 py-2 rounded-lg font-semibold transition"
              >
                <Play className="w-4 h-4" />
                {isLoading ? 'Running...' : 'Run Backtest'}
              </button>
              <button
                onClick={handleReset}
                className="flex items-center justify-center gap-2 bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg font-semibold transition"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Results Panel */}
        <div className="lg:col-span-2">
          {isLoading ? (
            <div className="bg-slate-800 rounded-lg border border-slate-700 p-8 space-y-4">
              <div className="flex items-center justify-center h-96">
                <div className="text-center">
                  <div className="inline-flex items-center justify-center w-12 h-12 bg-blue-600/20 rounded-full mb-4">
                    <div className="w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                  </div>
                  <p className="text-slate-300 font-semibold">Running backtest...</p>
                  <p className="text-slate-400 text-sm mt-1">This may take a moment</p>
                </div>
              </div>
            </div>
          ) : results ? (
            <div className="space-y-4">
              {/* Summary Card */}
              <div className="bg-gradient-to-br from-slate-800 to-slate-750 rounded-lg border border-slate-700 p-6 space-y-4">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-white">Backtest Results</h3>
                  <span className="text-sm text-slate-400">
                    {results.start_date} to {results.end_date}
                  </span>
                </div>

                {/* Main Metrics */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Total Return</p>
                    <p className={`text-2xl font-bold ${
                      results.total_return_pct >= 0 ? 'text-green-400' : 'text-red-400'
                    }`}>
                      {results.total_return_pct.toFixed(2)}%
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Win Rate</p>
                    <p className="text-2xl font-bold text-blue-400">
                      {results.win_rate.toFixed(2)}%
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Largest Win</p>
                    <p className="text-2xl font-bold text-green-400">
                      {formatCurrency(results.largest_win)}
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Largest Loss</p>
                    <p className="text-2xl font-bold text-red-400">
                      {formatCurrency(Math.abs(results.largest_loss))}
                    </p>
                  </div>
                </div>
              </div>

              {/* Detailed Metrics */}
              <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
                <div className="flex items-center gap-2 mb-3">
                  <TrendingUp className="w-5 h-5 text-blue-400" />
                  <h4 className="font-semibold text-white">Performance</h4>
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-slate-400">Starting Balance:</span>
                    <span className="text-white font-semibold">{formatCurrency(results.initial_capital)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Ending Balance:</span>
                    <span className={`font-semibold ${results.final_balance >= results.initial_capital ? 'text-green-400' : 'text-red-400'}`}>
                      {formatCurrency(results.final_balance)}
                    </span>
                  </div>
                  <div className="flex justify-between pt-2 border-t border-slate-700">
                    <span className="text-slate-400">Total Trades:</span>
                    <span className="text-white font-semibold">{results.total_trades}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Winning Trades:</span>
                    <span className="text-green-400 font-semibold">{results.winning_trades}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Losing Trades:</span>
                    <span className="text-red-400 font-semibold">{results.losing_trades}</span>
                  </div>
                </div>
              </div>

              {/* Historical Bars Chart */}
              {results.historical_bars && results.historical_bars.length > 0 && (
                <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
                  <h3 className="text-lg font-semibold text-white mb-4">Price Movement</h3>
                  <div className="w-full h-96">
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={results.historical_bars}
                        margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                        <XAxis 
                          dataKey="date" 
                          stroke="#64748b"
                          tick={{ fontSize: 12 }}
                          interval={Math.floor(results.historical_bars.length / 10)}
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
                            return [value, name];
                          }}
                          cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
                        />
                        <Legend 
                          wrapperStyle={{ paddingTop: '20px' }}
                          iconType="line"
                        />
                        <Line 
                          type="monotone" 
                          dataKey="close" 
                          stroke="#22c55e" 
                          dot={false}
                          strokeWidth={2}
                          name="Close"
                          isAnimationActive={false}
                        />
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
                      </LineChart>
                    </ResponsiveContainer>
                  </div>

                  {/* Price Statistics */}
                  <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mt-6 text-sm">
                    <div className="bg-slate-700/30 rounded p-3">
                      <p className="text-slate-400 text-xs mb-1">Highest</p>
                      <p className="font-bold text-green-400 text-lg">
                        ${Math.max(...results.historical_bars.map(b => b.high)).toFixed(2)}
                      </p>
                    </div>
                    <div className="bg-slate-700/30 rounded p-3">
                      <p className="text-slate-400 text-xs mb-1">Lowest</p>
                      <p className="font-bold text-red-400 text-lg">
                        ${Math.min(...results.historical_bars.map(b => b.low)).toFixed(2)}
                      </p>
                    </div>
                    <div className="bg-slate-700/30 rounded p-3">
                      <p className="text-slate-400 text-xs mb-1">Avg Volume</p>
                      <p className="font-bold text-white text-lg">
                        {(results.historical_bars.reduce((sum, b) => sum + b.volume, 0) / results.historical_bars.length / 1000000).toFixed(1)}M
                      </p>
                    </div>
                    <div className="bg-slate-700/30 rounded p-3">
                      <p className="text-slate-400 text-xs mb-1">Range</p>
                      <p className="font-bold text-white text-lg">
                        ${(Math.max(...results.historical_bars.map(b => b.high)) - Math.min(...results.historical_bars.map(b => b.low))).toFixed(2)}
                      </p>
                    </div>
                    <div className="bg-slate-700/30 rounded p-3">
                      <p className="text-slate-400 text-xs mb-1">% Change</p>
                      <p className={`font-bold text-lg ${
                        results.historical_bars[results.historical_bars.length - 1].close >= results.historical_bars[0].open
                          ? 'text-green-400'
                          : 'text-red-400'
                      }`}>
                        {(((results.historical_bars[results.historical_bars.length - 1].close - results.historical_bars[0].open) / results.historical_bars[0].open) * 100).toFixed(2)}%
                      </p>
                    </div>
                  </div>

                  {/* Trades Table */}
                  {results.trades && results.trades.length > 0 && (
                    <div className="mt-8 bg-slate-700/20 rounded-lg border border-slate-600">
                      <div className="p-4 border-b border-slate-600">
                        <h3 className="text-white font-semibold">Trade Details</h3>
                      </div>
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b border-slate-600 text-slate-400 text-xs uppercase">
                              <th className="px-4 py-3 text-left">#</th>
                              <th className="px-4 py-3 text-left">Entry Date</th>
                              <th className="px-4 py-3 text-right">Entry Price</th>
                              <th className="px-4 py-3 text-left">Exit Date</th>
                              <th className="px-4 py-3 text-right">Exit Price</th>
                              <th className="px-4 py-3 text-right">Qty</th>
                              <th className="px-4 py-3 text-right">P&L</th>
                              <th className="px-4 py-3 text-right">Return %</th>
                            </tr>
                          </thead>
                          <tbody>
                            {results.trades.map((trade, idx) => (
                              <tr key={idx} className="border-b border-slate-700 hover:bg-slate-700/20">
                                <td className="px-4 py-3 text-white">{trade.trade_num}</td>
                                <td className="px-4 py-3 text-slate-300">{trade.entry_time}</td>
                                <td className="px-4 py-3 text-right text-slate-300">${trade.entry_price.toFixed(2)}</td>
                                <td className="px-4 py-3 text-slate-300">{trade.exit_time}</td>
                                <td className="px-4 py-3 text-right text-slate-300">${trade.exit_price.toFixed(2)}</td>
                                <td className="px-4 py-3 text-right text-slate-400">{trade.quantity.toFixed(2)}</td>
                                <td className={`px-4 py-3 text-right font-semibold ${
                                  trade.pnl >= 0 ? 'text-green-400' : 'text-red-400'
                                }`}>
                                  ${trade.pnl.toFixed(2)}
                                </td>
                                <td className={`px-4 py-3 text-right font-semibold ${
                                  trade.return_pct >= 0 ? 'text-green-400' : 'text-red-400'
                                }`}>
                                  {trade.return_pct.toFixed(2)}%
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="bg-slate-800 rounded-lg border border-slate-700 p-12 text-center h-96 flex items-center justify-center">
              <div>
                <BarChart3 className="w-16 h-16 text-slate-500 mx-auto mb-4 opacity-50" />
                <p className="text-slate-400 mb-2">No backtest results yet</p>
                <p className="text-slate-500 text-sm">Fill in the parameters and run a backtest to see results here</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
