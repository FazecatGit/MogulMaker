'use client';

import { useState, useEffect } from 'react';
import { TrendingUp, TrendingDown, Clock, Search, ArrowUp, ArrowDown, RefreshCw } from 'lucide-react';
import { useTrades } from '@/hooks/useTrades';
import { useTradeStatistics } from '@/hooks/useTradeStatistics';
import { useGlobalStore } from '@/store/useGlobalStore';
import PageHeader from '@/components/PageHeader';
import Button from '@/components/ui/Button';
import Card from '@/components/ui/Card';
import StatCard from '@/components/ui/StatCard';
import ErrorAlert from '@/components/ui/ErrorAlert';
import SearchInput from '@/components/ui/SearchInput';
import SelectInput from '@/components/ui/SelectInput';
import SkeletonLoader from '@/components/ui/SkeletonLoader';
import PendingOrdersAlert from '@/components/ui/PendingOrdersAlert';
import ResponsiveTable from '@/components/Tables/ResponsiveTable';
import apiClient from '@/lib/apiClient';
import { formatCurrency, formatPercent, formatDuration } from '@/lib/formatters';
import { getPnLColor, getStatCardVariant } from '@/lib/colorHelpers';

interface Position {
  symbol: string;
  unrealized_pl: number;
  unrealized_plpc: number;
  current_price: number;
  [key: string]: any;
}

interface PendingOrder {
  id: string;
  symbol: string;
  side: 'buy' | 'sell';
  qty: string;
  filled_qty: string;
  type: string;
  status: string;
  submitted_at: string;
  filled_avg_price: string | null;
}

export default function TradesPage() {
  const { data, isLoading, error, isError, refetch } = useTrades();
  const { data: statsData, isLoading: statsLoading } = useTradeStatistics();
  const addNotification = useGlobalStore((state) => state.addNotification);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'open' | 'closed'>('all');
  const [sortBy, setSortBy] = useState<'recent' | 'pnl' | 'duration'>('recent');
  const [tradeSymbol, setTradeSymbol] = useState('');
  const [tradeQuantity, setTradeQuantity] = useState(1);
  const [tradingLoading, setTradingLoading] = useState(false);
  const [positions, setPositions] = useState<Position[]>([]);
  const [pendingOrders, setPendingOrders] = useState<PendingOrder[]>([]);
  const [recentTradeExecution, setRecentTradeExecution] = useState(false);

  // Fetch positions for unrealized P&L and pending orders
  useEffect(() => {
    const fetchPositionsAndOrders = async () => {
      try {
        const [riskResponse, positionsResponse] = await Promise.all([
          apiClient.get('/risk'),
          apiClient.get('/positions'),
        ]);
        
        const riskData = riskResponse?.data || riskResponse;
        if (riskData?.positions) {
          setPositions(riskData.positions);
        }
        
        const posData = positionsResponse?.data || positionsResponse;
        if (posData?.pending_orders) {
          setPendingOrders(posData.pending_orders);
        }
      } catch (err) {
        console.error('Failed to fetch positions:', err);
      }
    };

    fetchPositionsAndOrders();
    const interval = setInterval(fetchPositionsAndOrders, 30000); // Refresh every 30s
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
      <div className="w-full space-y-8">
        <PageHeader title="Trade History" description="View and analyze your trading history" />

        <SkeletonLoader count={5} />
      </div>
    );
  }

  // Error state
  if (isError) {
    return (
      <div className="w-full space-y-8">
        <PageHeader title="Trade History" description="View and analyze your trading history" />
        <ErrorAlert 
          title="Failed to load trade history"
          message={error instanceof Error ? error.message : 'Unknown error'}
        />
      </div>
    );
  }

  const trades = sortedTrades;
  const openTrades = trades.filter((t) => t.status === 'open').length;
  const closedTrades = trades.filter((t) => t.status === 'closed').length;
  const totalPnL = trades.reduce((sum, t) => sum + (typeof t.realized_pl === 'number' ? t.realized_pl : parseFloat(t.realized_pl || '0')), 0);

  const handleLongTrade = async () => {
    if (!tradeSymbol.trim()) {
      addNotification({
        type: 'error',
        title: 'Invalid Trade',
        message: 'Please enter a symbol',
      });
      return;
    }
    
    const symbol = tradeSymbol.trim().toUpperCase();
    
    // Check if there are existing positions or pending orders for this symbol
    const existingPosition = positions.find(p => p.symbol === symbol);
    const symbolPendingOrders = pendingOrders.filter(o => o.symbol === symbol && o.side === 'buy') || [];
    
    const existingQty = existingPosition ? parseFloat(existingPosition.qty?.toString() || '0') : 0;
    const pendingQty = symbolPendingOrders.reduce((sum, order) => sum + parseFloat(order.qty), 0);
    
    // Show confirmation if there are existing shares or pending orders
    if (existingQty > 0 || pendingQty > 0) {
      let confirmMessage = `You currently have:\n`;
      if (existingQty > 0) {
        confirmMessage += `• ${existingQty} shares of ${symbol} (filled)\n`;
      }
      if (pendingQty > 0) {
        confirmMessage += `• ${pendingQty} shares of ${symbol} (pending)\n`;
      }
      confirmMessage += `\nAre you sure you want to buy ${tradeQuantity} more shares of ${symbol}?`;
      
      if (!window.confirm(confirmMessage)) {
        return;
      }
    }
    
    setTradingLoading(true);
    setRecentTradeExecution(false);
    try {
      await apiClient.post('/execute-trade', {
        symbol: symbol,
        side: 'buy',
        quantity: tradeQuantity,
      });
      addNotification({
        type: 'success',
        title: 'Trade Executed',
        message: `Long trade executed for ${symbol} (${tradeQuantity} shares)`,
      });
      setRecentTradeExecution(true);
      setTradeSymbol('');
      setTradeQuantity(1);
      // Refresh trades data immediately
      await refetch();
      // Remove animation after 3 seconds
      setTimeout(() => setRecentTradeExecution(false), 3000);
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Trade Failed',
        message: 'Failed to execute long trade. Please try again.',
      });
    } finally {
      setTradingLoading(false);
    }
  };

  const handleShortTrade = async () => {
    if (!tradeSymbol.trim()) {
      addNotification({
        type: 'error',
        title: 'Invalid Trade',
        message: 'Please enter a symbol',
      });
      return;
    }
    setTradingLoading(true);
    setRecentTradeExecution(false);
    const symbol = tradeSymbol.trim().toUpperCase();
    try {
      await apiClient.post('/execute-trade', {
        symbol: symbol,
        side: 'sell',
        quantity: tradeQuantity,
      });
      addNotification({
        type: 'success',
        title: 'Trade Executed',
        message: `Short trade executed for ${symbol} (${tradeQuantity} shares)`,
      });
      setRecentTradeExecution(true);
      setTradeSymbol('');
      setTradeQuantity(1);
      // Refresh trades data immediately
      await refetch();
      // Remove animation after 3 seconds
      setTimeout(() => setRecentTradeExecution(false), 3000);
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Trade Failed',
        message: 'Failed to execute short trade. Please try again.',
      });
    } finally {
      setTradingLoading(false);
    }
  };

  return (
    <div className="w-full space-y-8">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <PageHeader 
            title="Trades" 
            description={`${trades.length} trade${trades.length !== 1 ? 's' : ''} • ${openTrades} open • ${closedTrades} closed${pendingOrders.length > 0 ? ` • ${pendingOrders.length} pending order${pendingOrders.length !== 1 ? 's' : ''}` : ''}`} 
          />
          <Button
            variant="secondary"
            icon={<RefreshCw className="w-4 h-4" />}
            onClick={() => refetch()}
          >
            Refresh
          </Button>
        </div>
      </div>

      <PendingOrdersAlert orders={pendingOrders} />

      {/* Trade Statistics */}
      {statsData && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard label="Win Rate" value={`${statsData.win_rate.toFixed(1)}%`} subtext={`${statsData.winning_trades}W / ${statsData.losing_trades}L`} />
          <StatCard label="Realized P&L" value={`$${statsData.total_pnl.toFixed(2)}`} subtext={`Avg: $${statsData.avg_pnl.toFixed(2)}`} variant={getStatCardVariant(statsData.total_pnl)} />
          <StatCard label="Unrealized P&L" value={`$${statsData.open_pnl.toFixed(2)}`} subtext={`${statsData.open_positions} open position(s)`} variant={getStatCardVariant(statsData.open_pnl)} />
          <StatCard label="Sharpe Ratio" value={statsData.sharpe_ratio.toFixed(2)} subtext={`Win: $${statsData.largest_win.toFixed(2)} / Loss: $${Math.abs(statsData.largest_loss).toFixed(2)}`} />
        </div>
      )}

      {/* Trading Panel */}
      <div className={`bg-slate-800 rounded-lg p-6 space-y-4 ${
        recentTradeExecution 
          ? 'trading-success-pulse' 
          : 'border border-slate-700'
      }`}>
        <h2 className="text-xl font-bold text-white flex items-center gap-2">
          <Search className="w-5 h-5" />
          Execute Trade
        </h2>
        
        <div className="space-y-4">
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
          <SearchInput value={searchTerm} onChange={setSearchTerm} placeholder="Search by symbol..." className="flex-1" />
          <SelectInput
            value={filterStatus}
            onChange={(v) => setFilterStatus(v as any)}
            options={[
              { value: 'all', label: 'All Trades' },
              { value: 'open', label: 'Open' },
              { value: 'closed', label: 'Closed' },
            ]}
          />
          <SelectInput
            value={sortBy}
            onChange={(v) => setSortBy(v as any)}
            options={[
              { value: 'recent', label: 'Sort by Recent' },
              { value: 'pnl', label: 'Sort by P&L' },
              { value: 'duration', label: 'Sort by Duration' },
            ]}
          />
        </div>
      </div>

      {/* Trades Table */}
      <ResponsiveTable
        data={trades}
        keyExtractor={(trade) => trade.id}
        columns={[
          {
            key: 'symbol',
            label: 'Symbol',
            render: (val, trade) => (
              <div>
                <div className="font-semibold text-white">{val}</div>
                <div className="text-xs text-slate-400">{trade.exchange}</div>
              </div>
            ),
          },
          {
            key: 'side',
            label: 'Side',
            align: 'center',
            render: (val) => (
              <span className={`px-3 py-1 rounded text-sm font-semibold ${
                val === 'buy' ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'
              }`}>
                {val.toUpperCase()}
              </span>
            ),
          },
          {
            key: 'entry_price',
            label: 'Entry Price',
            align: 'right',
            render: (val) => formatCurrency(typeof val === 'number' ? val : parseFloat(val)),
          },
          {
            key: 'exit_price',
            label: 'Exit Price',
            align: 'right',
            render: (val, trade) => {
              if (trade.status === 'open') {
                const position = positionMap.get(trade.symbol);
                const currentPrice = position?.current_price;
                return currentPrice ? (
                  <span className="text-blue-400">{formatCurrency(currentPrice)}</span>
                ) : (
                  <span className="text-slate-500">Open</span>
                );
              }
              return val ? formatCurrency(typeof val === 'number' ? val : parseFloat(val)) : <span className="text-yellow-400">Pending</span>;
            },
          },
          {
            key: 'qty',
            label: 'Quantity',
            align: 'right',
            render: (val) => (typeof val === 'number' ? val : parseFloat(val)).toFixed(0),
          },
          {
            key: 'entry_time',
            label: 'Entry Time',
            render: (val) => new Date(val).toLocaleDateString('en-US', {
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            }),
            mobileHidden: true,
          },
          {
            key: 'duration_ms',
            label: 'Duration',
            align: 'center',
            render: (val, trade) => {
              if (trade.status === 'open') {
                const entryTime = new Date(trade.entry_time).getTime();
                const now = Date.now();
                const openDuration = Math.floor((now - entryTime) / 1000);
                return (
                  <div className="flex items-center justify-center gap-1 text-blue-400">
                    <Clock className="w-4 h-4" />
                    {formatDuration(openDuration)} (open)
                  </div>
                );
              }
              if (!val || val === 0) {
                return (
                  <div className="flex items-center justify-center gap-1 text-yellow-400">
                    <Clock className="w-4 h-4" />
                    <span className="text-xs">Calculating...</span>
                  </div>
                );
              }
              return (
                <div className="flex items-center justify-center gap-1">
                  <Clock className="w-4 h-4" />
                  {formatDuration(val / 1000)}
                </div>
              );
            },
          },
          {
            key: 'realized_pl',
            label: 'P&L',
            align: 'right',
            render: (val, trade) => {
              const position = positionMap.get(trade.symbol);
              const pnl = trade.status === 'open' ? (position?.unrealized_pl || 0) : (val || 0);
              return (
                <div className={`flex items-center justify-end gap-1 font-semibold ${getPnLColor(pnl)}`}>
                  {pnl >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                  {pnl >= 0 ? '+' : ''}{formatCurrency(Math.abs(pnl))}
                </div>
              );
            },
          },
          {
            key: 'realized_plpc',
            label: 'Return %',
            align: 'right',
            render: (val, trade) => {
              const position = positionMap.get(trade.symbol);
              const pnlPercent = trade.status === 'open' ? ((position?.unrealized_plpc || 0) * 100) : ((val || 0) * 100);
              return (
                <span className={`font-semibold ${getPnLColor(pnlPercent)}`}>
                  {pnlPercent >= 0 ? '+' : ''}{pnlPercent.toFixed(2)}%
                </span>
              );
            },
          },
          {
            key: 'status',
            label: 'Status',
            align: 'center',
            render: (val) => (
              <span className={`px-3 py-1 rounded text-xs font-semibold ${
                val === 'closed' ? 'bg-slate-700 text-slate-300' : 'bg-blue-900/30 text-blue-400'
              }`}>
                {val.charAt(0).toUpperCase() + val.slice(1)}
              </span>
            ),
          },
        ]}
        renderMobileCard={(trade) => {
          const position = positionMap.get(trade.symbol);
          const pnl = trade.status === 'open' ? (position?.unrealized_pl || 0) : (trade.realized_pl || 0);
          const pnlPercent = trade.status === 'open' ? ((position?.unrealized_plpc || 0) * 100) : ((trade.realized_plpc || 0) * 100);
          return (
            <>
              <div className="space-y-2 text-sm mb-3">
                <div className="flex justify-between"><span className="text-slate-400">Entry:</span><span className="text-white">{formatCurrency(typeof trade.entry_price === 'number' ? trade.entry_price : parseFloat(trade.entry_price))}</span></div>
                {trade.exit_price && <div className="flex justify-between"><span className="text-slate-400">Exit:</span><span className="text-white">{formatCurrency(typeof trade.exit_price === 'number' ? trade.exit_price : parseFloat(trade.exit_price))}</span></div>}
                <div className="flex justify-between"><span className="text-slate-400">Qty:</span><span className="text-white font-semibold">{(typeof trade.qty === 'number' ? trade.qty : parseFloat(trade.qty)).toFixed(0)}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Entry Time:</span><span className="text-white text-xs">{new Date(trade.entry_time).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Duration:</span><span className="text-white flex items-center gap-1"><Clock className="w-3 h-3" />{formatDuration((trade.duration_ms || 0) / 1000)}</span></div>
              </div>
              <div className={`p-3 rounded bg-slate-600/30 border mb-2 ${pnl >= 0 ? 'border-green-700/30' : 'border-red-700/30'}`}>
                <div className="flex justify-between"><span className="text-slate-300">P&L:</span><span className={`font-semibold flex items-center gap-1 ${getPnLColor(pnl)}`}>{pnl >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}{pnl >= 0 ? '+' : ''}{formatCurrency(Math.abs(pnl))} ({pnl >= 0 ? '+' : ''}{pnlPercent.toFixed(2)}%)</span></div>
              </div>
              <div className="flex justify-end"><span className={`px-2 py-1 rounded text-xs font-semibold ${trade.status === 'closed' ? 'bg-slate-700 text-slate-300' : 'bg-blue-900/30 text-blue-400'}`}>{trade.status.charAt(0).toUpperCase() + trade.status.slice(1)}</span></div>
            </>
          );
        }}
        emptyMessage="No trades found"
      />

      {/* Summary */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total Trades" value={trades.length.toString()} />
        <StatCard label="Open Trades" value={openTrades.toString()} variant="neutral" />
        <StatCard label="Closed Trades" value={closedTrades.toString()} />
        <StatCard label="Total P&L" value={`${totalPnL >= 0 ? '+' : ''}$${Math.abs(totalPnL).toFixed(2)}`} variant={getStatCardVariant(totalPnL)} />
      </div>
    </div>
  );
}
