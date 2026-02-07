'use client';

import { useState } from 'react';
import { Calendar, Play, RotateCcw, TrendingUp, BarChart3, Save, X } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import Button from '@/components/ui/Button';
import StatCard from '@/components/ui/StatCard';
import ErrorAlert from '@/components/ui/ErrorAlert';
import PriceChart from '@/components/Charts/PriceChart';
import ResponsiveTable from '@/components/Tables/ResponsiveTable';
import apiClient from '@/lib/apiClient';
import { formatCurrency, formatPercent, formatNumber } from '@/lib/formatters';
import { getPnLColor, getStatCardVariant } from '@/lib/colorHelpers';

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
  const [savedBacktests, setSavedBacktests] = useState<BacktestResult[]>([]);
  const [showComparison, setShowComparison] = useState(false);

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

  const handleSaveBacktest = () => {
    if (results) {
      setSavedBacktests(prev => [...prev, { ...results, created_at: Date.now() }]);
      setShowComparison(true);
    }
  };

  const handleRemoveBacktest = (index: number) => {
    setSavedBacktests(prev => prev.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <PageHeader 
        title="Strategy Backtest" 
        description="Test your trading strategy on historical data"
      />

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
              <ErrorAlert message={error} />
            )}

            {/* Action Buttons */}
            <div className="flex gap-2 pt-2">
              <Button
                variant="primary"
                icon={<Play className="w-4 h-4" />}
                loading={isLoading}
                onClick={handleRunBacktest}
                className="flex-1"
              >
                {isLoading ? 'Running...' : 'Run Backtest'}
              </Button>
              <Button
                variant="secondary"
                icon={<RotateCcw className="w-4 h-4" />}
                onClick={handleReset}
              />
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
                  <div>
                    <h3 className="text-lg font-bold text-white">Backtest Results</h3>
                    <span className="text-sm text-slate-400">
                      {results.start_date} to {results.end_date}
                    </span>
                  </div>
                  <div className="flex gap-2">
                    {savedBacktests.length > 0 && (
                      <Button
                        variant="secondary"
                        icon={<BarChart3 className="w-4 h-4" />}
                        onClick={() => setShowComparison(!showComparison)}
                      >
                        {showComparison ? 'Hide' : 'Show'} Comparison ({savedBacktests.length})
                      </Button>
                    )}
                    <Button
                      variant="primary"
                      icon={<Save className="w-4 h-4" />}
                      onClick={handleSaveBacktest}
                    >
                      Save for Comparison
                    </Button>
                  </div>
                </div>

                {/* Main Metrics */}
                <div className="grid grid-cols-2 gap-4">
                  <StatCard
                    label="Total Return"
                    value={`${results.total_return_pct.toFixed(2)}%`}
                    variant={getStatCardVariant(results.total_return_pct)}
                  />
                  <StatCard
                    label="Win Rate"
                    value={`${results.win_rate.toFixed(2)}%`}
                    variant="neutral"
                  />
                  <StatCard
                    label="Largest Win"
                    value={formatCurrency(results.largest_win)}
                    variant="positive"
                  />
                  <StatCard
                    label="Largest Loss"
                    value={formatCurrency(Math.abs(results.largest_loss))}
                    variant="negative"
                  />
                </div>
              </div>

              {/* Comparison Table */}
              {showComparison && savedBacktests.length > 0 && (
                <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-bold text-white">Backtest Comparison</h3>
                    <Button
                      variant="secondary"
                      icon={<X className="w-4 h-4" />}
                      onClick={() => setSavedBacktests([])}
                    >
                      Clear All
                    </Button>
                  </div>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-slate-700">
                          <th className="text-left p-3 text-slate-400 font-semibold">Metric</th>
                          {savedBacktests.map((bt, idx) => (
                            <th key={idx} className="text-right p-3">
                              <div className="flex flex-col items-end gap-1">
                                <span className="text-white font-bold">{bt.symbol}</span>
                                <span className="text-xs text-slate-400">{bt.start_date}</span>
                                <button
                                  onClick={() => handleRemoveBacktest(idx)}
                                  className="text-red-400 hover:text-red-300 text-xs"
                                >
                                  Remove
                                </button>
                              </div>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Total Return</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className={`p-3 text-right font-semibold ${getPnLColor(bt.total_return_pct)}`}>
                              {bt.total_return_pct.toFixed(2)}%
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Win Rate</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-white font-semibold">
                              {bt.win_rate.toFixed(2)}%
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Total Trades</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-white font-semibold">
                              {bt.total_trades}
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Winning Trades</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-green-400 font-semibold">
                              {bt.winning_trades}
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Losing Trades</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-red-400 font-semibold">
                              {bt.losing_trades}
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Largest Win</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-green-400 font-semibold">
                              {formatCurrency(bt.largest_win)}
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Largest Loss</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className="p-3 text-right text-red-400 font-semibold">
                              {formatCurrency(Math.abs(bt.largest_loss))}
                            </td>
                          ))}
                        </tr>
                        <tr className="border-b border-slate-700/50">
                          <td className="p-3 text-slate-300">Final Balance</td>
                          {savedBacktests.map((bt, idx) => (
                            <td key={idx} className={`p-3 text-right font-semibold ${bt.final_balance >= bt.initial_capital ? 'text-green-400' : 'text-red-400'}`}>
                              {formatCurrency(bt.final_balance)}
                            </td>
                          ))}
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

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
                <>
                  <PriceChart
                    data={results.historical_bars}
                    title="Price Movement"
                    daysLabel={`${results.historical_bars.length} days`}
                    showVolume={true}
                    height={400}
                  />

                  {/* Price Statistics */}
                  <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
                    <StatCard
                      label="Highest"
                      value={formatCurrency(Math.max(...results.historical_bars.map(b => b.high)))}
                      variant="positive"
                    />
                    <StatCard
                      label="Lowest"
                      value={formatCurrency(Math.min(...results.historical_bars.map(b => b.low)))}
                      variant="negative"
                    />
                    <StatCard
                      label="Avg Volume"
                      value={formatNumber(results.historical_bars.reduce((sum, b) => sum + b.volume, 0) / results.historical_bars.length / 1000000, 1) + 'M'}
                    />
                    <StatCard
                      label="Range"
                      value={formatCurrency(Math.max(...results.historical_bars.map(b => b.high)) - Math.min(...results.historical_bars.map(b => b.low)))}
                    />
                    <StatCard
                      label="% Change"
                      value={formatPercent((results.historical_bars[results.historical_bars.length - 1].close - results.historical_bars[0].open) / results.historical_bars[0].open)}
                      variant={getStatCardVariant(results.historical_bars[results.historical_bars.length - 1].close - results.historical_bars[0].open)}
                    />
                  </div>

                  {/* Trades Table */}
                  {results.trades && results.trades.length > 0 && (
                    <div className="mt-8">
                      <h3 className="text-white font-semibold mb-4">Trade Details</h3>
                      <ResponsiveTable
                        data={results.trades}
                        keyExtractor={(trade, idx) => idx}
                        columns={[
                          {
                            key: 'trade_num' as any,
                            label: '#',
                            width: '50px',
                            align: 'left',
                          },
                          {
                            key: 'entry_time' as any,
                            label: 'Entry Date',
                            mobileHidden: true,
                          },
                          {
                            key: 'entry_price' as any,
                            label: 'Entry Price',
                            align: 'right',
                            render: (val) => formatCurrency(val),
                          },
                          {
                            key: 'exit_time' as any,
                            label: 'Exit Date',
                            mobileHidden: true,
                          },
                          {
                            key: 'exit_price' as any,
                            label: 'Exit Price',
                            align: 'right',
                            render: (val) => formatCurrency(val),
                          },
                          {
                            key: 'quantity' as any,
                            label: 'Qty',
                            align: 'right',
                            render: (val) => val.toFixed(2),
                          },
                          {
                            key: 'pnl' as any,
                            label: 'P&L',
                            align: 'right',
                            render: (val) => (
                              <span className={getPnLColor(val)}>
                                {formatCurrency(val)}
                              </span>
                            ),
                          },
                          {
                            key: 'return_pct' as any,
                            label: 'Return %',
                            align: 'right',
                            render: (val) => (
                              <span className={getPnLColor(val)}>
                                {formatPercent(val / 100)}
                              </span>
                            ),
                          },
                        ]}
                        emptyMessage="No trades"
                      />
                    </div>
                  )}
                </>
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
