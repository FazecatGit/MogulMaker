'use client';

import { useState } from 'react';
import { AlertCircle, TrendingUp, TrendingDown, Plus, Minus } from 'lucide-react';
import { usePositionsTable, type Position } from '@/hooks/usePositionsTable';
import apiClient from '@/lib/apiClient';

export default function PositionsPage() {
  const { data, isLoading, error, isError, refetch } = usePositionsTable();
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'symbol' | 'pnl' | 'value'>('symbol');
  const [closingSymbol, setClosingSymbol] = useState<string | null>(null);
  const [buyingSymbol, setBuyingSymbol] = useState<string | null>(null);

  const handleClosePosition = async (symbol: string, quantity: number) => {
    // Ask for confirmation before selling
    const confirmed = window.confirm(
      `Are you sure you want to sell all ${quantity} shares of ${symbol}? This action cannot be undone.`
    );
    if (!confirmed) return;

    setClosingSymbol(symbol);
    try {
      // Sell all shares for this position
      await apiClient.post('/trades', {
        symbol,
        side: 'sell',
        quantity: quantity,
      });
      await refetch();
    } catch (err) {
      console.error('Failed to close position:', err);
    } finally {
      setClosingSymbol(null);
    }
  };

  const handleBuyMore = async (symbol: string) => {
    const existingPosition = data?.positions.find(p => p.symbol === symbol);
    const pendingOrders = data?.pending_orders?.filter(o => o.symbol === symbol && o.side === 'buy') || [];
    
    const existingQty = existingPosition ? parseFloat(existingPosition.qty) : 0;
    const pendingQty = pendingOrders.reduce((sum, order) => sum + parseFloat(order.qty), 0);

    if (existingQty > 0 || pendingQty > 0) {
      let confirmMessage = `You currently have:\n`;
      if (existingQty > 0) {
        confirmMessage += `• ${existingQty} shares of ${symbol} (filled)\n`;
      }
      if (pendingQty > 0) {
        confirmMessage += `• ${pendingQty} shares of ${symbol} (pending)\n`;
      }
      confirmMessage += `\nAre you sure you want to buy more ${symbol}?`;
      
      if (!window.confirm(confirmMessage)) {
        return;
      }
    }
    
    const quantityStr = prompt(`How many shares of ${symbol} do you want to buy?`, '1');
    if (!quantityStr || isNaN(parseFloat(quantityStr)) || parseFloat(quantityStr) <= 0) {
      return;
    }
    
    setBuyingSymbol(symbol);
    try {
      await apiClient.post('/trades', {
        symbol,
        side: 'buy',
        quantity: parseFloat(quantityStr),
      });
      await refetch();
    } catch (err) {
      console.error('Failed to buy more shares:', err);
    } finally {
      setBuyingSymbol(null);
    }
  };

  // Filter positions by search term
  const filteredPositions = (data?.positions || []).filter((pos) =>
    pos.symbol.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Sort positions
  const sortedPositions = [...filteredPositions].sort((a, b) => {
    switch (sortBy) {
      case 'pnl':
        return parseFloat(b.unrealized_pl) - parseFloat(a.unrealized_pl);
      case 'value':
        return parseFloat(b.market_value) - parseFloat(a.market_value);
      case 'symbol':
      default:
        return a.symbol.localeCompare(b.symbol);
    }
  });

  // Loading state
  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Positions</h1>
          <p className="text-slate-400">View and manage your open positions</p>
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
          <h1 className="text-3xl font-bold text-white mb-2">Positions</h1>
          <p className="text-slate-400">View and manage your open positions</p>
        </div>

        <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
          <div>
            <p className="text-red-400 font-semibold">Failed to load positions</p>
            <p className="text-red-300 text-sm">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  const positions = sortedPositions;
  const pendingOrders = data?.pending_orders || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Positions</h1>
        <p className="text-slate-400">
          {positions.length} open position{positions.length !== 1 ? 's' : ''}
          {pendingOrders.length > 0 && ` • ${pendingOrders.length} pending order${pendingOrders.length !== 1 ? 's' : ''}`}
        </p>
      </div>

      {/* Pending Orders Alert */}
      {pendingOrders.length > 0 && (
        <div className="bg-yellow-900/10 border-2 border-yellow-500 rounded-lg p-6 shadow-lg">
          <div className="flex items-start gap-4">
            <div className="bg-yellow-500 rounded-full p-2">
              <AlertCircle className="w-6 h-6 text-slate-900 flex-shrink-0" />
            </div>
            <div className="flex-1">
              <p className="text-yellow-400 font-bold text-lg mb-3">
                Pending Orders ({pendingOrders.length})
              </p>
              <div className="space-y-2">
                {pendingOrders.map((order) => (
                  <div key={order.id} className="bg-slate-800/50 rounded px-3 py-2 border border-yellow-500/30">
                    <span className="font-bold text-yellow-300 text-base">{order.symbol}</span>
                    {' • '}
                    <span className="capitalize text-white font-medium">{order.side}</span>
                    {' • '}
                    <span className="text-white">{parseFloat(order.qty)} shares</span>
                    {' • '}
                    <span className="text-yellow-400 font-semibold uppercase text-xs">{order.status}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

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

          {/* Sort */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="bg-slate-700 text-white rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none"
          >
            <option value="symbol">Sort by Symbol</option>
            <option value="pnl">Sort by P&L</option>
            <option value="value">Sort by Value</option>
          </select>
        </div>
      </div>

      {/* Positions Table */}
      {positions.length > 0 ? (
        <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
          {/* Desktop View - Table */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50 border-b border-slate-700">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-slate-300">
                    Symbol
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Quantity
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Entry Price
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Current Price
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Market Value
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    P&L
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Return %
                  </th>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {positions.map((pos) => {
                  const pnl = parseFloat(pos.unrealized_pl);
                  const pnlPercent = parseFloat(pos.unrealized_plpc) * 100;
                  const isPositive = pnl >= 0;

                  return (
                    <tr key={pos.asset_id} className="hover:bg-slate-700/30 transition">
                      <td className="px-6 py-4">
                        <div className="font-semibold text-white">{pos.symbol}</div>
                        <div className="text-xs text-slate-400">{pos.exchange}</div>
                      </td>
                      <td className="px-6 py-4 text-right text-white">
                        {parseFloat(pos.qty).toFixed(0)}
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        ${parseFloat(pos.avg_entry_price).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        ${parseFloat(pos.current_price).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 text-right text-white">
                        ${parseFloat(pos.market_value).toFixed(2)}
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
                      <td className="px-6 py-4 text-center flex gap-2 justify-center">
                        <button 
                          onClick={() => handleBuyMore(pos.symbol)}
                          disabled={buyingSymbol === pos.symbol}
                          className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-sm font-medium text-white transition flex items-center gap-1"
                        >
                          <Plus className="w-4 h-4" />
                          {buyingSymbol === pos.symbol ? 'Buying...' : 'Buy'}
                        </button>
                        <button 
                          onClick={() => handleClosePosition(pos.symbol, parseFloat(pos.qty))}
                          disabled={closingSymbol === pos.symbol}
                          className="px-3 py-1 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-sm font-medium text-white transition flex items-center gap-1"
                        >
                          <Minus className="w-4 h-4" />
                          {closingSymbol === pos.symbol ? 'Selling...' : 'Sell'}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Mobile View - Cards */}
          <div className="md:hidden space-y-3 p-4">
            {positions.map((pos) => {
              const pnl = parseFloat(pos.unrealized_pl);
              const pnlPercent = parseFloat(pos.unrealized_plpc) * 100;
              const isPositive = pnl >= 0;

              return (
                <div key={pos.asset_id} className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <div className="font-semibold text-white text-lg">{pos.symbol}</div>
                      <div className="text-xs text-slate-400">{pos.exchange}</div>
                    </div>
                    <div className="flex gap-2">
                      <button 
                        onClick={() => handleBuyMore(pos.symbol)}
                        disabled={buyingSymbol === pos.symbol}
                        className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-sm font-medium text-white transition flex items-center gap-1"
                      >
                        <Plus className="w-4 h-4" />
                        {buyingSymbol === pos.symbol ? 'Buying...' : 'Buy'}
                      </button>
                      <button 
                        onClick={() => handleClosePosition(pos.symbol, parseFloat(pos.qty))}
                        disabled={closingSymbol === pos.symbol}
                        className="px-3 py-1 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed rounded text-sm font-medium text-white transition flex items-center gap-1"
                      >
                        <Minus className="w-4 h-4" />
                        {closingSymbol === pos.symbol ? 'Selling...' : 'Sell'}
                      </button>
                    </div>
                  </div>

                  <div className="space-y-2 text-sm mb-3">
                    <div className="flex justify-between">
                      <span className="text-slate-400">Quantity:</span>
                      <span className="text-white font-semibold">
                        {parseFloat(pos.qty).toFixed(0)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Entry Price:</span>
                      <span className="text-white">${parseFloat(pos.avg_entry_price).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Current Price:</span>
                      <span className="text-white">${parseFloat(pos.current_price).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-400">Market Value:</span>
                      <span className="text-white font-semibold">
                        ${parseFloat(pos.market_value).toFixed(2)}
                      </span>
                    </div>
                  </div>

                  <div className={`p-3 rounded bg-slate-600/30 border ${isPositive ? 'border-green-700/30' : 'border-red-700/30'}`}>
                    <div className="flex justify-between">
                      <span className="text-slate-300">P&L:</span>
                      <span className={`font-semibold flex items-center gap-1 ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
                        {isPositive ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                        {isPositive ? '+' : ''}${Math.abs(pnl).toFixed(2)} ({isPositive ? '+' : ''}
                        {pnlPercent.toFixed(2)}%)
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="bg-slate-800 rounded-lg p-12 border border-slate-700 text-center">
          <p className="text-slate-400">No positions found</p>
        </div>
      )}

      {/* Summary */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Total Positions</p>
          <p className="text-2xl font-bold text-white">{positions.length}</p>
        </div>
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Total Position Value</p>
          <p className="text-2xl font-bold text-white">
            ${positions.reduce((sum, pos) => sum + parseFloat(pos.market_value), 0).toFixed(2)}
          </p>
        </div>
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <p className="text-slate-400 text-sm mb-1">Total Unrealized P&L</p>
          <p className={`text-2xl font-bold ${
            positions.reduce((sum, pos) => sum + parseFloat(pos.unrealized_pl), 0) >= 0
              ? 'text-green-400'
              : 'text-red-400'
          }`}>
            ${positions.reduce((sum, pos) => sum + parseFloat(pos.unrealized_pl), 0).toFixed(2)}
          </p>
        </div>
      </div>
    </div>
  );
}