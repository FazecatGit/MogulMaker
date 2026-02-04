'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, TrendingUp, TrendingDown, Clock, Search, ArrowUp, ArrowDown } from 'lucide-react';
import { useTrades } from '@/hooks/useTrades';
import { useTradeStatistics } from '@/hooks/useTradeStatistics';
import apiClient from '@/lib/apiClient';

interface Position {
  symbol: string;
  unrealized_pl: number;
  unrealized_plpc: number;
  current_price: number;
  [key: string]: any;
}

export default function TradesPage() {
  const { data, isLoading, error, isError } = useTrades();
  const { data: statsData, isLoading: statsLoading } = useTradeStatistics();
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'open' | 'closed'>('all');
  const [sortBy, setSortBy] = useState<'recent' | 'pnl' | 'duration'>('recent');
  const [tradeSymbol, setTradeSymbol] = useState('');
  const [tradeQuantity, setTradeQuantity] = useState(1);
  const [tradingLoading, setTradingLoading] = useState(false);
  const [tradingMessage, setTradingMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);

  // Fetch positions for unrealized P&L
  useEffect(() => {
    const fetchPositions = async () => {
      try {
        const response = await apiClient.get('/risk');
        const backendData = response?.data || response;
        if (backendData?.positions) {
          setPositions(backendData.positions);
        }
      } catch (err) {
        console.error('Failed to fetch positions:', err);
      }
    };

    fetchPositions();
    const interval = setInterval(fetchPositions, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  // Create position map for quick lookup
  const positionMap = new Map(positions.map(p => [p.symbol, p]));

  // Filter trades
  const filteredTrades = (data?.trades || []).filter((trade) => {
    const matchesSearch = trade.symbol.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || trade.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  // Sort trades
  const sortedTrades = [...filteredTrades].sort((a, b) => {
    switch (sortBy) {
      case 'pnl': {
        const pnlA = a.status === 'open' 
          ? (positionMap.get(a.symbol)?.unrealized_pl || 0)
          : (a.realized_pl || 0);
        const pnlB = b.status === 'open'
          ? (positionMap.get(b.symbol)?.unrealized_pl || 0)
          : (b.realized_pl || 0);
        return pnlB - pnlA;
      }
      case 'duration':
        const durationA = a.duration_ms || 0;
        const durationB = b.duration_ms || 0;
        return durationB - durationA;
      case 'recent':
      default:
        return new Date(b.entry_time).getTime() - new Date(a.entry_time).getTime();
    }
  });

  // Loading state
  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Trade History</h1>
          <p className="text-slate-400">View and analyze your trading history</p>
        </div>

        {/* Skeleton loaders */}
        <div className="space-y-2">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="bg-slate-800 rounded-lg h-16 animate-pulse border border-slate-700"
            />
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (isError) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Trade History</h1>
          <p className="text-slate-400">View and analyze your trading history</p>
        </div>

        <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
          <div>
            <p className="text-red-400 font-semibold">Failed to load trade history</p>
            <p className="text-red-300 text-sm">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  const trades = sortedTrades;
  const openTrades = trades.filter((t) => t.status === 'open').length;
  const closedTrades = trades.filter((t) => t.status === 'closed').length;
  const totalPnL = trades.reduce((sum, t) => sum + (typeof t.realized_pl === 'number' ? t.realized_pl : parseFloat(t.realized_pl || '0')), 0);

  const formatDuration = (ms?: number) => {
    if (!ms || ms <= 0) return 'Ongoing';
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d`;
    if (hours > 0) return `${hours}h`;
    if (minutes > 0) return `${minutes}m`;
    return `${seconds}s`;
  };

  const handleLongTrade = async () => {
    if (!tradeSymbol.trim()) {
      setTradingMessage({ type: 'error', text: 'Please enter a symbol' });
      return;
    }
    setTradingLoading(true);
    setTradingMessage(null);
    try {
      await apiClient.post('/execute-trade', {
        symbol: tradeSymbol.toUpperCase(),
        side: 'buy',
        quantity: tradeQuantity,
      });
      setTradingMessage({ type: 'success', text: `Long trade executed for ${tradeSymbol.toUpperCase()}` });
      setTradeSymbol('');
      setTradeQuantity(1);
    } catch (err) {
      setTradingMessage({ type: 'error', text: 'Failed to execute long trade' });
    } finally {
      setTradingLoading(false);
    }
  };

  const handleShortTrade = async () => {
    if (!tradeSymbol.trim()) {
      setTradingMessage({ type: 'error', text: 'Please enter a symbol' });
      return;
    }
    setTradingLoading(true);
    setTradingMessage(null);
    try {
      await apiClient.post('/execute-trade', {
        symbol: tradeSymbol.toUpperCase(),
        side: 'sell',
        quantity: tradeQuantity,
      });
      setTradingMessage({ type: 'success', text: `Short trade executed for ${tradeSymbol.toUpperCase()}` });
      setTradeSymbol('');
      setTradeQuantity(1);
    } catch (err) {
      setTradingMessage({ type: 'error', text: 'Failed to execute short trade' });
    } finally {
      setTradingLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Trades</h1>
        <p className="text-slate-400">
          {trades.length} trade{trades.length !== 1 ? 's' : ''} • {openTrades} open • {closedTrades} closed
        </p>
      </div>

      {/* Trade Statistics */}
      {statsData && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
            <p className="text-slate-400 text-sm mb-1">Win Rate</p>
            <p className="text-2xl font-bold text-white">{statsData.win_rate.toFixed(1)}%</p>
            <p className="text-xs text-slate-500 mt-1">{statsData.winning_trades}W / {statsData.losing_trades}L</p>
          </div>

          <div className={`bg-slate-800 rounded-lg border border-slate-700 p-4 ${statsData.total_pnl >= 0 ? 'border-green-600/30' : 'border-red-600/30'}`}>
            <p className="text-slate-400 text-sm mb-1">Realized P&L</p>
            <p className={`text-2xl font-bold ${statsData.total_pnl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              ${statsData.total_pnl.toFixed(2)}
            </p>
            <p className="text-xs text-slate-500 mt-1">Avg: ${statsData.avg_pnl.toFixed(2)}</p>
          </div>

          <div className={`bg-slate-800 rounded-lg border border-slate-700 p-4 ${statsData.open_pnl >= 0 ? 'border-green-600/30' : 'border-red-600/30'}`}>
            <p className="text-slate-400 text-sm mb-1">Unrealized P&L</p>
            <p className={`text-2xl font-bold ${statsData.open_pnl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              ${statsData.open_pnl.toFixed(2)}
            </p>
            <p className="text-xs text-slate-500 mt-1">{statsData.open_positions} open position(s)</p>
          </div>

          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
            <p className="text-slate-400 text-sm mb-1">Sharpe Ratio</p>
            <p className="text-2xl font-bold text-white">{statsData.sharpe_ratio.toFixed(2)}</p>
            <p className="text-xs text-slate-500 mt-1">Win: ${statsData.largest_win.toFixed(2)} / Loss: ${Math.abs(statsData.largest_loss).toFixed(2)}</p>
          </div>
        </div>
      )}

      {/* Trading Panel */}
      <div className="bg-slate-800 rounded-lg p-6 border border-slate-700 space-y-4">
        <h2 className="text-xl font-bold text-white flex items-center gap-2">
          <Search className="w-5 h-5" />
          Execute Trade
        </h2>
        
        <div className="space-y-4">
          {tradingMessage && (
            <div
              className={`rounded-lg p-4 ${
                tradingMessage.type === 'success'
                  ? 'bg-green-900/30 border border-green-600 text-green-400'
                  : 'bg-red-900/30 border border-red-600 text-red-400'
              }`}
            >
              {tradingMessage.text}
            </div>
          )}

          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <label className="block text-slate-300 text-sm font-semibold mb-2">Symbol</label>
              <input
                type="text"
                placeholder="e.g., TSLA, AAPL"
                value={tradeSymbol}
                onChange={(e) => setTradeSymbol(e.target.value.toUpperCase())}
                className="w-full bg-slate-700 text-white placeholder-slate-400 rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
              />
            </div>

            <div className="w-32">
              <label className="block text-slate-300 text-sm font-semibold mb-2">Quantity</label>
              <input
                type="number"
                min="1"
                value={tradeQuantity}
                onChange={(e) => setTradeQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                className="w-full bg-slate-700 text-white placeholder-slate-400 rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
              />
            </div>
          </div>

          <div className="flex gap-3">
            <button
              onClick={handleLongTrade}
              disabled={tradingLoading}
              className="flex-1 flex items-center justify-center gap-2 bg-green-600 hover:bg-green-700 disabled:bg-slate-600 text-white font-semibold py-3 rounded-lg transition"
            >
              <ArrowUp className="w-5 h-5" />
              {tradingLoading ? 'Processing...' : 'Long'}
            </button>
            <button
              onClick={handleShortTrade}
              disabled={tradingLoading}
              className="flex-1 flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 disabled:bg-slate-600 text-white font-semibold py-3 rounded-lg transition"
            >
              <ArrowDown className="w-5 h-5" />
              {tradingLoading ? 'Processing...' : 'Short'}
            </button>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="bg-slate-800 rounded-lg p-4 border border-slate-700 space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Search */}
          <input
            type="text"
            placeholder="Search by symbol..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="flex-1 bg-slate-700 text-white placeholder-slate-400 rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
          />

          {/* Status Filter */}
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value as any)}
            className="bg-slate-700 text-white rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
          >
            <option value="all">All Trades</option>
            <option value="open">Open</option>
            <option value="closed">Closed</option>
          </select>

          {/* Sort */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="bg-slate-700 text-white rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
          >
            <option value="recent">Sort by Recent</option>
            <option value="pnl">Sort by P&L</option>
            <option value="duration">Sort by Duration</option>
          </select>
        </div>
      </div>

      {/* Trades Table */}
      {trades.length > 0 ? (
        <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
          {/* Desktop View - Table */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50 border-b border-slate-700">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-slate-300">
                    Symbol
                  </th>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Side
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Entry Price
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Exit Price
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Quantity
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-slate-300">
                    Entry Time
                  </th>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Duration
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    P&L
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Return %
                  </th>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Status
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {trades.map((trade) => {
                  const position = positionMap.get(trade.symbol);
                  const pnl = trade.status === 'open' 
                    ? (position?.unrealized_pl || 0)
                    : (trade.realized_pl || 0);
                  const pnlPercent = trade.status === 'open'
                    ? ((position?.unrealized_plpc || 0) * 100)
                    : ((trade.realized_plpc || 0) * 100);
                  const isPositive = pnl >= 0;

                  return (
                    <tr key={trade.id} className="hover:bg-slate-700/30 transition">
                      <td className="px-6 py-4">
                        <div className="font-semibold text-white">{trade.symbol}</div>
                        <div className="text-xs text-slate-400">{trade.exchange}</div>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span className={`px-3 py-1 rounded text-sm font-semibold ${
                          trade.side === 'buy'
                            ? 'bg-green-900/30 text-green-400'
                            : 'bg-red-900/30 text-red-400'
                        }`}>
                          {trade.side.toUpperCase()}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        ${typeof trade.entry_price === 'number' ? trade.entry_price.toFixed(2) : parseFloat(trade.entry_price).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        {trade.exit_price ? `$${typeof trade.exit_price === 'number' ? trade.exit_price.toFixed(2) : parseFloat(trade.exit_price).toFixed(2)}` : '-'}
                      </td>
                      <td className="px-6 py-4 text-right text-white">
                        {typeof trade.qty === 'number' ? trade.qty.toFixed(0) : parseFloat(trade.qty).toFixed(0)}
                      </td>
                      <td className="px-6 py-4 text-left text-slate-300 text-sm">
                        {new Date(trade.entry_time).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </td>
                      <td className="px-6 py-4 text-center text-slate-300 text-sm">
                        <div className="flex items-center justify-center gap-1">
                          <Clock className="w-4 h-4" />
                          {formatDuration(trade.duration_ms)}
                        </div>
                      </td>
                      <td
                        className={`px-6 py-4 text-right font-semibold ${
                          isPositive ? 'text-green-400' : 'text-red-400'
                        }`}
                      >
                        <div className="flex items-center justify-end gap-1">
                          {isPositive ? (
                            <TrendingUp className="w-4 h-4" />
                          ) : (
                            <TrendingDown className="w-4 h-4" />
                          )}
                          {isPositive ? '+' : ''}${Math.abs(pnl).toFixed(2)}
                        </div>
                      </td>
                      <td
                        className={`px-6 py-4 text-right font-semibold ${
                          isPositive ? 'text-green-400' : 'text-red-400'
                        }`}
                      >
                        {isPositive ? '+' : ''}{pnlPercent.toFixed(2)}%
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span className={`px-3 py-1 rounded text-xs font-semibold ${
                          trade.status === 'closed'
                            ? 'bg-slate-700 text-slate-300'
                            : 'bg-blue-900/30 text-blue-400'
                        }`}>
                          {trade.status.charAt(0).toUpperCase() + trade.status.slice(1)}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Mobile View - Cards */}
          <div className="md:hidden space-y-3 p-4">
            {trades.map((trade) => {
              const position = positionMap.get(trade.symbol);
              const pnl = trade.status === 'open' 
                ? (position?.unrealized_pl || 0)
                : (trade.realized_pl || 0);
              const pnlPercent = trade.status === 'open'
                ? ((position?.unrealized_plpc || 0) * 100)
                : ((trade.realized_plpc || 0) * 100);
              const isPositive = pnl >= 0;

              return (
                <div key={trade.id} className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <div className="font-semibold text-white text-lg">{trade.symbol}</div>
                      <div className="text-xs text-slate-400">{trade.exchange}</div>
                    </div>
                    <span className={`px-2 py-1 rounded text-xs font-semibold ${
                      trade.side === 'buy'
                        ? 'bg-green-900/30 text-green-400'
                        : 'bg-red-900/30 text-red-400'
                    }`}>
                      {trade.side.toUpperCase()}
                    </span>
                  </div>

                  <div className="space-y-2 text-sm mb-3">
                    <div className="flex justify-between">
                      <span className="text-slate-400">Entry Price:</span>
                      <span className="text-white">${typeof trade.entry_price === 'number' ? trade.entry_price.toFixed(2) : parseFloat(trade.entry_price).toFixed(2)}</span>
                    </div>
                    {trade.exit_price && (
                      <div className="flex justify-between">
                        <span className="text-slate-400">Exit Price:</span>
                        <span className="text-white">${typeof trade.exit_price === 'number' ? trade.exit_price.toFixed(2) : parseFloat(trade.exit_price).toFixed(2)}</span>
                      </div>
                    )}
                    <div className="flex justify-between">
                      <span className="text-slate-400">Quantity:</span>
                      <span className="text-white font-semibold">{typeof trade.qty === 'number' ? trade.qty.toFixed(0) : parseFloat(trade.qty).toFixed(0)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Entry Time:</span>
                      <span className="text-white text-xs">
                        {new Date(trade.entry_time).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Duration:</span>
                      <span className="text-white flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {formatDuration(trade.duration_ms)}
                      </span>
                    </div>
                  </div>

                  <div className={`p-3 rounded bg-slate-600/30 border mb-2 ${isPositive ? 'border-green-700/30' : 'border-red-700/30'}`}>
                    <div className="flex justify-between">
                      <span className="text-slate-300">P&L:</span>
                      <span className={`font-semibold flex items-center gap-1 ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
                        {isPositive ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                        {isPositive ? '+' : ''}${Math.abs(pnl).toFixed(2)} ({isPositive ? '+' : ''}
                        {pnlPercent.toFixed(2)}%)
                      </span>
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <span className={`px-2 py-1 rounded text-xs font-semibold ${
                      trade.status === 'closed'
                        ? 'bg-slate-700 text-slate-300'
                        : 'bg-blue-900/30 text-blue-400'
                    }`}>
                      {trade.status.charAt(0).toUpperCase() + trade.status.slice(1)}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="bg-slate-800 rounded-lg p-12 border border-slate-700 text-center">
          <p className="text-slate-400">No trades found</p>
        </div>
      )}

      {/* Summary */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Total Trades</p>
          <p className="text-2xl font-bold text-white">{trades.length}</p>
        </div>
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Open Trades</p>
          <p className="text-2xl font-bold text-blue-400">{openTrades}</p>
        </div>
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Closed Trades</p>
          <p className="text-2xl font-bold text-white">{closedTrades}</p>
        </div>
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Total P&L</p>
          <p className={`text-2xl font-bold ${totalPnL >= 0 ? 'text-green-400' : 'text-red-400'}`}>
            {totalPnL >= 0 ? '+' : ''}${Math.abs(totalPnL).toFixed(2)}
          </p>
        </div>
      </div>
    </div>
  );
}
