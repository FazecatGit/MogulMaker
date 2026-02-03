'use client';

import { useState } from 'react';
import { Calendar, Play, RotateCcw, TrendingUp, BarChart3, AlertCircle } from 'lucide-react';
import apiClient from '@/lib/apiClient';

interface BacktestResult {
  symbol: string;
  start_date: string;
  end_date: string;
  total_return: number;
  trades_count: number;
  win_rate: number;
  max_drawdown: number;
  sharpe_ratio: number;
  starting_balance: number;
  ending_balance: number;
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
      });

      const data = await apiClient.get(`/backtest?${params.toString()}`);
      setResults(data.data as BacktestResult);
    } catch (err: any) {
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
                      results.total_return >= 0 ? 'text-green-400' : 'text-red-400'
                    }`}>
                      {formatPercent(results.total_return)}
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Win Rate</p>
                    <p className="text-2xl font-bold text-blue-400">
                      {formatPercent(results.win_rate)}
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Trades</p>
                    <p className="text-2xl font-bold text-white">
                      {results.trades_count}
                    </p>
                  </div>

                  <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                    <p className="text-slate-400 text-xs mb-1">Max Drawdown</p>
                    <p className="text-2xl font-bold text-orange-400">
                      {formatPercent(results.max_drawdown)}
                    </p>
                  </div>
                </div>
              </div>

              {/* Detailed Metrics */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <TrendingUp className="w-5 h-5 text-blue-400" />
                    <h4 className="font-semibold text-white">Performance</h4>
                  </div>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-slate-400">Starting Balance:</span>
                      <span className="text-white font-semibold">{formatCurrency(results.starting_balance)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Ending Balance:</span>
                      <span className={`font-semibold ${results.ending_balance >= results.starting_balance ? 'text-green-400' : 'text-red-400'}`}>
                        {formatCurrency(results.ending_balance)}
                      </span>
                    </div>
                    <div className="flex justify-between pt-2 border-t border-slate-700">
                      <span className="text-slate-400">Sharpe Ratio:</span>
                      <span className="text-white font-semibold">{results.sharpe_ratio.toFixed(2)}</span>
                    </div>
                  </div>
                </div>

                <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <BarChart3 className="w-5 h-5 text-green-400" />
                    <h4 className="font-semibold text-white">Trade Statistics</h4>
                  </div>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-slate-400">Total Trades:</span>
                      <span className="text-white font-semibold">{results.trades_count}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Winning Trades:</span>
                      <span className="text-green-400 font-semibold">
                        {Math.round(results.trades_count * results.win_rate)}
                      </span>
                    </div>
                    <div className="flex justify-between pt-2 border-t border-slate-700">
                      <span className="text-slate-400">Max Drawdown:</span>
                      <span className="text-orange-400 font-semibold">{formatPercent(results.max_drawdown)}</span>
                    </div>
                  </div>
                </div>
              </div>
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
