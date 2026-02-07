'use client';

import { useState } from 'react';
import { TrendingUp, TrendingDown, Plus, Minus, Radio } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import Button from '@/components/ui/Button';
import StatCard from '@/components/ui/StatCard';
import ErrorAlert from '@/components/ui/ErrorAlert';
import SearchInput from '@/components/ui/SearchInput';
import SelectInput from '@/components/ui/SelectInput';
import SkeletonLoader from '@/components/ui/SkeletonLoader';
import PendingOrdersAlert from '@/components/ui/PendingOrdersAlert';
import ResponsiveTable from '@/components/Tables/ResponsiveTable';
import { usePositionsTable, type Position } from '@/hooks/usePositionsTable';
import apiClient from '@/lib/apiClient';
import { formatCurrency } from '@/lib/formatters';
import { getPnLColor, getStatCardVariant } from '@/lib/colorHelpers';

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
        <PageHeader 
          title="Positions" 
          description="View and manage your open positions"
        />

        <SkeletonLoader count={5} />
      </div>
    );
  }

  // Error state
  if (isError) {
    return (
      <div className="space-y-6">
        <PageHeader 
          title="Positions" 
          description="View and manage your open positions"
        />
        <ErrorAlert 
          title="Failed to load positions"
          message={error instanceof Error ? error.message : 'Unknown error'}
        />
      </div>
    );
  }

  const positions = sortedPositions;
  const pendingOrders = data?.pending_orders || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <PageHeader 
            title="Positions" 
            description={`${positions.length} open position${positions.length !== 1 ? 's' : ''}${pendingOrders.length > 0 ? ` • ${pendingOrders.length} pending order${pendingOrders.length !== 1 ? 's' : ''}` : ''}`}
          />
          <div className="flex items-center gap-2 bg-green-900/20 border border-green-700 rounded-full px-3 py-1.5 shrink-0">
            <Radio className="w-3 h-3 text-green-400 animate-pulse" />
            <span className="text-green-400 text-sm font-medium">Live Updates</span>
            <span className="text-green-300/60 text-xs">(5s)</span>
          </div>
        </div>
      </div>

      <PendingOrdersAlert orders={pendingOrders} />

      {/* Controls */}
      <div className="bg-slate-800 rounded-lg p-4 border border-slate-700 space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <SearchInput
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search by symbol..."
            className="flex-1"
          />
          <SelectInput
            value={sortBy}
            onChange={(v) => setSortBy(v as any)}
            options={[
              { value: 'symbol', label: 'Sort by Symbol' },
              { value: 'pnl', label: 'Sort by P&L' },
              { value: 'value', label: 'Sort by Value' },
            ]}
          />
        </div>
      </div>

      {/* Positions Table */}
      <ResponsiveTable
        data={positions}
        keyExtractor={(pos) => pos.asset_id}
        columns={[
          {
            key: 'symbol',
            label: 'Symbol',
            render: (val, pos) => (
              <div>
                <div className="font-semibold text-white">{val}</div>
                <div className="text-xs text-slate-400">{pos.exchange}</div>
              </div>
            ),
          },
          { key: 'qty', label: 'Quantity', align: 'right', render: (val) => parseFloat(val).toFixed(0) },
          { key: 'avg_entry_price', label: 'Entry Price', align: 'right', render: (val) => formatCurrency(parseFloat(val)) },
          { key: 'current_price', label: 'Current Price', align: 'right', render: (val) => formatCurrency(parseFloat(val)) },
          { key: 'market_value', label: 'Market Value', align: 'right', render: (val) => formatCurrency(parseFloat(val)) },
          {
            key: 'unrealized_pl',
            label: 'P&L',
            align: 'right',
            render: (val) => {
              const pnl = parseFloat(val);
              return (
                <div className={`flex items-center justify-end gap-1 font-semibold ${getPnLColor(pnl)}`}>
                  {pnl >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                  {pnl >= 0 ? '+' : ''}{formatCurrency(Math.abs(pnl))}
                </div>
              );
            },
          },
          {
            key: 'unrealized_plpc',
            label: 'Return %',
            align: 'right',
            render: (val) => {
              const pnlPercent = parseFloat(val) * 100;
              return (
                <span className={`font-semibold ${getPnLColor(pnlPercent)}`}>
                  {pnlPercent >= 0 ? '+' : ''}{pnlPercent.toFixed(2)}%
                </span>
              );
            },
          },
        ]}
        renderActions={(pos) => (
          <div className="flex gap-2">
            <Button
              variant="success"
              icon={<Plus className="w-4 h-4" />}
              onClick={() => handleBuyMore(pos.symbol)}
              loading={buyingSymbol === pos.symbol}
              className="text-sm px-3 py-1"
            >
              Buy
            </Button>
            <Button
              variant="danger"
              icon={<Minus className="w-4 h-4" />}
              onClick={() => handleClosePosition(pos.symbol, parseFloat(pos.qty))}
              loading={closingSymbol === pos.symbol}
              className="text-sm px-3 py-1"
            >
              Sell
            </Button>
          </div>
        )}
        renderMobileCard={(pos) => {
          const pnl = parseFloat(pos.unrealized_pl);
          const pnlPercent = parseFloat(pos.unrealized_plpc) * 100;
          return (
            <>
              <div className="space-y-2 text-sm mb-3">
                <div className="flex justify-between"><span className="text-slate-400">Quantity:</span><span className="text-white font-semibold">{parseFloat(pos.qty).toFixed(0)}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Entry Price:</span><span className="text-white">{formatCurrency(parseFloat(pos.avg_entry_price))}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Current Price:</span><span className="text-white">{formatCurrency(parseFloat(pos.current_price))}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Market Value:</span><span className="text-white font-semibold">{formatCurrency(parseFloat(pos.market_value))}</span></div>
              </div>
              <div className={`p-3 rounded bg-slate-600/30 border ${pnl >= 0 ? 'border-green-700/30' : 'border-red-700/30'}`}>
                <div className="flex justify-between">
                  <span className="text-slate-300">P&L:</span>
                  <span className={`font-semibold flex items-center gap-1 ${getPnLColor(pnl)}`}>
                    {pnl >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                    {pnl >= 0 ? '+' : ''}{formatCurrency(Math.abs(pnl))} ({pnl >= 0 ? '+' : ''}{pnlPercent.toFixed(2)}%)
                  </span>
                </div>
              </div>
            </>
          );
        }}
        emptyMessage="No positions found"
      />

      {/* Summary */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          label="Total Positions"
          value={positions.length.toString()}
        />
        <StatCard
          label="Total Position Value"
          value={formatCurrency(positions.reduce((sum, pos) => sum + parseFloat(pos.market_value), 0))}
        />
        <StatCard
          label="Total Unrealized P&L"
          value={formatCurrency(positions.reduce((sum, pos) => sum + parseFloat(pos.unrealized_pl), 0))}
          variant={getStatCardVariant(positions.reduce((sum, pos) => sum + parseFloat(pos.unrealized_pl), 0))}
        />
      </div>
    </div>
  );
}