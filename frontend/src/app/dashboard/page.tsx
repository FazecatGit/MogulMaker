'use client';

import { usePortfolio } from '@/hooks/usePortfolio';
import { useTradeStatistics } from '@/hooks/useTradeStatistics';
import { useTrades } from '@/hooks/useTrades';
import { usePositionsTable } from '@/hooks/usePositionsTable';
import PageHeader from '@/components/PageHeader';
import { Loader2, AlertCircle, TrendingUp, TrendingDown } from 'lucide-react';
import { formatCurrency, formatPercent } from '@/lib/formatters';
import {
  LineChart,
  Line,
  PieChart,
  Pie,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';

/**
 * Dashboard Page - Real Data Version
 * 
 * FLOW:
 * 1. Component calls usePortfolio() and useTradeStatistics() hooks
 * 2. Hooks fetch data from API gateway
 * 3. Process data for charts (time-series, pie chart by symbol)
 * 4. Display real data with visualizations
 */

export default function DashboardPage() {
  const { data, isLoading, error, isError } = usePortfolio();
  const { data: statsData, isLoading: statsLoading } = useTradeStatistics();
  const { data: tradesData, isLoading: tradesLoading } = useTrades();
  const { data: positionsData, isLoading: positionsLoading } = usePositionsTable();

  // Process trade data for visualizations
  const processTradeData = () => {
    if (!tradesData?.trades) return { bySymbol: [], timeSeries: [], topBuys: [], topSells: [] };

    // Group trades by symbol and calculate P&L
    const bySymbol: { [key: string]: { symbol: string; pnl: number; trades: number; avgReturn: number } } = {};
    
    tradesData.trades.forEach((trade: any) => {
      if (!bySymbol[trade.symbol]) {
        bySymbol[trade.symbol] = { symbol: trade.symbol, pnl: 0, trades: 0, avgReturn: 0 };
      }
      bySymbol[trade.symbol].pnl += trade.realized_pl || 0;
      bySymbol[trade.symbol].trades += 1;
      bySymbol[trade.symbol].avgReturn += (trade.realized_plpc || 0) * 100;
    });

    // Convert to array and calculate avg return
    const bySymbolArray = Object.values(bySymbol).map(item => ({
      ...item,
      avgReturn: item.avgReturn / item.trades,
    }));

    // Time-series earnings (aggregate by date) - SORTED chronologically
    const timeSeriesMap: { [date: string]: { pnl: number; dateObj: Date } } = {};
    tradesData.trades.forEach((trade: any) => {
      const entryDate = new Date(trade.entry_time);
      const date = entryDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      if (!timeSeriesMap[date]) {
        timeSeriesMap[date] = { pnl: 0, dateObj: entryDate };
      }
      timeSeriesMap[date].pnl += trade.realized_pl || 0;
    });

    // Sort by actual date (past to present)
    const timeSeries = Object.entries(timeSeriesMap)
      .sort(([, a], [, b]) => a.dateObj.getTime() - b.dateObj.getTime())
      .map(([date, data]) => ({
        date,
        pnl: parseFloat(data.pnl.toFixed(2)),
      }));

    // Top Buys (sort by quantity)
    const topBuys = tradesData.trades
      .filter((t: any) => t.side === 'buy')
      .sort((a: any, b: any) => (parseFloat(b.qty) || 0) - (parseFloat(a.qty) || 0))
      .slice(0, 5)
      .map((trade: any) => ({
        symbol: trade.symbol,
        qty: parseFloat(trade.qty),
        price: parseFloat(trade.entry_price),
      }));

    // Top Sells
    const topSells = tradesData.trades
      .filter((t: any) => t.side === 'sell')
      .sort((a: any, b: any) => (parseFloat(b.qty) || 0) - (parseFloat(a.qty) || 0))
      .slice(0, 5)
      .map((trade: any) => ({
        symbol: trade.symbol,
        qty: parseFloat(trade.qty),
        price: parseFloat(trade.exit_price || trade.entry_price),
      }));

    return { bySymbol: bySymbolArray, timeSeries, topBuys, topSells };
  };

  // Process current positions for pie chart
  const processCurrentPositions = () => {
    if (!positionsData?.positions) return [];
    
    return positionsData.positions.map((pos: any) => ({
      symbol: pos.symbol,
      marketValue: parseFloat(pos.market_value),
      unrealizedPl: parseFloat(pos.unrealized_pl),
      qty: parseFloat(pos.qty),
      currentPrice: parseFloat(pos.current_price),
    }));
  };

  const chartData = processTradeData();
  const currentPositions = processCurrentPositions();
  const COLORS = ['#22c55e', '#ef4444', '#3b82f6', '#f59e0b', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316'];

  // Loading state - show spinners
  if (isLoading || statsLoading || tradesLoading || positionsLoading) {
    return (
      <div>
        <PageHeader title="Portfolio Dashboard" />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="bg-slate-800 rounded-lg p-6 border border-slate-700 animate-pulse"
            >
              <div className="h-4 bg-slate-700 rounded mb-4 w-20"></div>
              <div className="h-8 bg-slate-700 rounded mb-2 w-32"></div>
              <div className="h-4 bg-slate-700 rounded w-24"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (isError) {
    return (
      <div>
        <h1 className="text-3xl font-bold mb-8">Portfolio Dashboard</h1>
        <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-400" />
          <div>
            <p className="text-red-400 font-semibold">Failed to load portfolio</p>
            <p className="text-red-300 text-sm">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Success state - display real data
  return (
    <div>
      <PageHeader title="Portfolio Dashboard" />

      {/* Real Data Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Total P&L Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Total P&L</p>
          <p
            className={`text-3xl font-bold ${
              data && data.totalPnL >= 0 ? 'text-green-400' : 'text-red-400'
            }`}
          >
            {data ? `${data.totalPnL >= 0 ? '+' : ''}${formatCurrency(data.totalPnL)}` : 'Loading...'}
          </p>
          <p
            className={`text-sm mt-2 ${
              data && data.dailyPnLPercent >= 0
                ? 'text-green-400'
                : 'text-red-400'
            }`}
          >
            {data ? `${data.dailyPnLPercent >= 0 ? '+' : ''}${formatPercent(data.dailyPnLPercent / 100, 1)}` : 'Loading...'}
          </p>
        </div>

        {/* Portfolio Value Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Portfolio Value</p>
          <p className="text-3xl font-bold text-white">
            {data ? formatCurrency(data.portfolioValue) : 'Loading...'}
          </p>
          <p className="text-sm text-slate-400 mt-2">
            {data ? `${data.openPositions} open positions` : 'Loading...'}
          </p>
        </div>

        {/* Win Rate Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Win Rate</p>
          <p className="text-3xl font-bold text-blue-400">
            {statsData ? formatPercent(statsData.win_rate / 100, 1) : 'Loading...'}
          </p>
          <p className="text-sm text-slate-400 mt-2">
            {statsData ? `${statsData.winning_trades}W / ${statsData.losing_trades}L` : 'Loading...'}
          </p>
        </div>
      </div>

      {/* Portfolio Earnings Over Time */}
      {chartData.timeSeries.length > 0 && (
        <div className="mt-8 bg-slate-800 rounded-lg p-6 border border-slate-700">
          <h2 className="text-xl font-bold text-white mb-4">Earnings Timeline</h2>
          <ResponsiveContainer width="100%" height={500}>
            <LineChart data={chartData.timeSeries}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="date" stroke="#64748b" tick={{ fontSize: 12 }} />
              <YAxis stroke="#64748b" tick={{ fontSize: 12 }} label={{ value: 'P&L ($)', angle: -90, position: 'insideLeft' }} />
              <Tooltip 
                contentStyle={{
                  backgroundColor: '#1e293b',
                  border: '1px solid #475569',
                  borderRadius: '6px',
                }}
                labelStyle={{ color: '#e2e8f0' }}
                formatter={(value: any) => [`$${value.toFixed(2)}`, 'P&L']}
              />
              <Line 
                type="monotone" 
                dataKey="pnl" 
                stroke="#22c55e" 
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Positions Held - Pie Chart */}
      {currentPositions && currentPositions.length > 0 && (
        <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
            <h2 className="text-xl font-bold text-white mb-4">Positions Held (Current)</h2>
            <ResponsiveContainer width="100%" height={500}>
              <PieChart>
                <Pie
                  data={currentPositions}
                  dataKey="marketValue"
                  nameKey="symbol"
                  cx="50%"
                  cy="50%"
                  outerRadius={160}
                  fill="#8884d8"
                  label={(entry: any) => `${entry.symbol}: $${entry.marketValue.toFixed(0)}`}
                >
                  {currentPositions.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip 
                  contentStyle={{
                    backgroundColor: '#1e293b',
                    border: '1px solid #475569',
                    borderRadius: '6px',
                  }}
                  labelStyle={{ color: '#e2e8f0' }}
                  formatter={(value: any, name: string | undefined) => {
                    if (name === 'marketValue') return [`$${value?.toFixed(2) || '0'}`, 'Market Value'];
                    return [value, name];
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Positions List - Sorted by Market Value */}
          <div className="bg-slate-800 rounded-lg p-6 border border-slate-70 flex flex-col h-[600px]">
            <h2 className="text-xl font-bold text-white mb-4">Position Details</h2>
            <div className="space-y-2 flex-1 overflow-y-auto pr-2">
              {currentPositions
                .sort((a, b) => b.marketValue - a.marketValue)
                .map((pos, index) => (
                  <div key={pos.symbol} className="bg-slate-700/30 rounded p-3 border border-slate-600 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div 
                        className="w-3 h-3 rounded-full" 
                        style={{ backgroundColor: COLORS[index % COLORS.length] }}
                      />
                      <div>
                        <p className="font-bold text-white">{pos.symbol}</p>
                        <p className="text-xs text-slate-400">{pos.qty.toFixed(0)} shares @ ${pos.currentPrice.toFixed(2)}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-white">${pos.marketValue.toFixed(2)}</p>
                      <p className={`text-sm ${pos.unrealizedPl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                        {pos.unrealizedPl >= 0 ? '+' : ''}${pos.unrealizedPl.toFixed(2)}
                      </p>
                    </div>
                  </div>
                ))}
            </div>
          </div>
        </div>
      )}

      {/* Top Buys and Sells */}
      {(chartData.topBuys.length > 0 || chartData.topSells.length > 0) && (
        <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Top Buys */}
          {chartData.topBuys.length > 0 && (
            <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
              <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-green-400" />
                Biggest Buys
              </h2>
              <div className="space-y-3">
                {chartData.topBuys.map((trade, index) => (
                  <div key={index} className="bg-green-900/20 border border-green-700/30 rounded p-3">
                    <div className="flex justify-between items-center mb-2">
                      <span className="font-bold text-green-400">{trade.symbol}</span>
                      <span className="text-white font-semibold">{trade.qty.toFixed(0)} shares</span>
                    </div>
                    <div className="text-sm text-slate-300">
                      Entry: ${trade.price.toFixed(2)} | Total: ${(trade.qty * trade.price).toFixed(2)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Top Sells */}
          {chartData.topSells.length > 0 && (
            <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
              <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <TrendingDown className="w-5 h-5 text-red-400" />
                Biggest Sells
              </h2>
              <div className="space-y-3">
                {chartData.topSells.map((trade, index) => (
                  <div key={index} className="bg-red-900/20 border border-red-700/30 rounded p-3">
                    <div className="flex justify-between items-center mb-2">
                      <span className="font-bold text-red-400">{trade.symbol}</span>
                      <span className="text-white font-semibold">{trade.qty.toFixed(0)} shares</span>
                    </div>
                    <div className="text-sm text-slate-300">
                      Exit: ${trade.price.toFixed(2)} | Total: ${(trade.qty * trade.price).toFixed(2)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
