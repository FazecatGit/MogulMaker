'use client';

import { useState, useEffect, useRef } from 'react';
import { AlertCircle, TrendingUp, RefreshCw, Target, Zap, Star, Plus, X, Play } from 'lucide-react';
import { useScout } from '@/hooks/useScout';
import PageHeader from '@/components/PageHeader';
import Card from '@/components/ui/Card';
import SelectInput from '@/components/ui/SelectInput';
import SkeletonLoader from '@/components/ui/SkeletonLoader';
import StatusAlert from '@/components/ui/StatusAlert';
import StatCard from '@/components/ui/StatCard';
import ResponsiveTable from '@/components/Tables/ResponsiveTable';
import { getScoreBadgeColor } from '@/lib/colorHelpers';
import apiClient from '@/lib/apiClient';

export default function ScouterPage() {
  const [minScore, setMinScore] = useState(10);
  const [minScoreSlider, setMinScoreSlider] = useState(10); // Setup screen slider
  const [minScoreSidebarSlider, setMinScoreSidebarSlider] = useState(10); // Sidebar slider (doesn't trigger scans)
  const [limit, setLimit] = useState(15);
  const [scanTriggered, setScanTriggered] = useState(false);
  const [offset, setOffset] = useState(0);
  const [allOpportunities, setAllOpportunities] = useState<any[]>([]);
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null);
  const debounceTimer = useRef<NodeJS.Timeout | null>(null);
  
  console.log('[ScouterPage] Rendering with:', { minScore, limit, scanTriggered, offset });
  
  const { data, isLoading, error, isError, refetch } = useScout(minScore, limit, offset, scanTriggered);
  
  console.log('[ScouterPage] useScout result:', { data, isLoading, isError, error: error?.message });
  
  const [selectedSymbol, setSelectedSymbol] = useState<string | null>(null);
  const [addedToWatchlist, setAddedToWatchlist] = useState<Set<string>>(new Set());

  // Accumulate opportunities when we get new data
  useEffect(() => {
    if (data?.opportunities && offset === 0) {
      // First batch - replace
      setAllOpportunities(data.opportunities);
    } else if (data?.opportunities && offset > 0) {
      // Subsequent batches - append (avoiding duplicates)
      const existingSymbols = new Set(allOpportunities.map(o => o.symbol));
      const newOpportunities = data.opportunities.filter(o => !existingSymbols.has(o.symbol));
      setAllOpportunities(prev => [...prev, ...newOpportunities]);
    }
  }, [data?.opportunities, offset]);



  const handleStartScan = () => {
    console.log('[ScouterPage] Starting fresh scan with score:', minScoreSlider);
    setMinScore(minScoreSlider); // Use setup screen slider value for scan
    setOffset(0);
    setAllOpportunities([]);
    setScanTriggered(true);
  };

  const handleLoadNextBatch = () => {
    console.log('[ScouterPage] Loading next batch...');
    setOffset(prev => prev + limit);
  };

  const handleAddToWatchlist = async (symbol: string) => {
    try {
      await apiClient.post('/watchlist', { symbol });
      setAddedToWatchlist(prev => new Set([...prev, symbol]));
      setTimeout(() => {
        setAddedToWatchlist(prev => {
          const newSet = new Set(prev);
          newSet.delete(symbol);
          return newSet;
        });
      }, 2000);
    } catch (error) {
      console.error('Failed to add to watchlist:', error);
    }
  };

  // Initial state - show setup before scan
  if (!scanTriggered) {
    return (
      <div className="space-y-6">
        <PageHeader 
          title="Stock Scouter" 
          description="AI-powered stock screening and opportunity detection"
        />

        <div className="bg-slate-800 rounded-lg border border-slate-700 p-8 space-y-6">
          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-3">Number of Results</label>
            <div className="flex items-center gap-4">
              <div className="flex gap-2">
                <button
                  onClick={() => setLimit(10)}
                  className={`px-4 py-2 rounded font-semibold transition ${
                    limit === 10
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  10
                </button>
                <button
                  onClick={() => setLimit(15)}
                  className={`px-4 py-2 rounded font-semibold transition ${
                    limit === 15
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  15
                </button>
                <button
                  onClick={() => setLimit(25)}
                  className={`px-4 py-2 rounded font-semibold transition ${
                    limit === 25
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  25
                </button>
              </div>
              <input
                type="number"
                min="1"
                max="100"
                value={limit}
                onChange={(e) => setLimit(Math.min(100, Math.max(1, parseInt(e.target.value))))}
                className="w-20 px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
                placeholder="Custom"
                suppressHydrationWarning
              />
            </div>
            <p className="text-slate-400 text-xs mt-2">Selected: {limit} stocks</p>
          </div>

          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-3">Minimum Score</label>
            <div className="flex items-center gap-4">
              <input
                type="range"
                min="0"
                max="10"
                value={minScoreSlider}
                onChange={(e) => setMinScoreSlider(parseInt(e.target.value))}
                className="flex-1 h-2 bg-slate-700 rounded cursor-pointer"
              />
              <span className="text-white font-bold text-lg w-12">{minScoreSlider}</span>
            </div>
          </div>

          <button
            onClick={handleStartScan}
            className="w-full flex items-center justify-center gap-2 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 text-white px-6 py-3 rounded-lg font-semibold text-lg transition"
          >
            <Play className="w-5 h-5" />
            Start Scan
          </button>
        </div>
      </div>
    );
  }

  // Loading state - only show full skeleton on initial load (offset=0)
  if (isLoading && allOpportunities.length === 0) {
    return (
      <div className="space-y-6">
        <PageHeader title="Stock Scouter" description="Scanning stocks..." />
        <SkeletonLoader count={3} height="h-24" />
        <SkeletonLoader count={5} height="h-20" />
      </div>
    );
  }

  // Error state
  if (isError) {
    console.error('[ScouterPage] Error state:', { error, isError });
    const errorMessage = error instanceof Error ? error.message : JSON.stringify(error);
    const errorStatus = (error as any)?.status ? ` (${(error as any).status})` : '';
    const errorCode = (error as any)?.code ? ` [${(error as any).code}]` : '';
    
    return (
      <div className="space-y-6">
        <PageHeader title="Stock Scouter" description="AI-powered stock screening and opportunity detection" />
        <StatusAlert message={`Failed to scan stocks: ${errorMessage}${errorStatus}${errorCode}`} variant="error" />
        <button
          onClick={() => setScanTriggered(false)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm font-semibold"
        >
          Try Again
        </button>
      </div>
    );
  }

  const opportunities = allOpportunities;
  const totalScanned = data?.total_symbols || 0;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
      {/* Sidebar */}
      <div className="lg:col-span-1">
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6 sticky top-4 space-y-6">
          <div>
            <h3 className="text-lg font-bold text-white mb-3">Filters</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-slate-300 text-sm font-semibold mb-2">
                  Results Shown
                </label>
                <p className="text-2xl font-bold text-blue-400">{opportunities.length}</p>
                <p className="text-xs text-slate-400 mt-1">of {totalScanned} total</p>
                {offset > 0 && (
                  <p className="text-xs text-slate-500 mt-2">
                    Showing batches 1-{Math.ceil((offset + limit) / limit)}
                  </p>
                )}
              </div>

              <div className="border-t border-slate-700 pt-4">
                <label className="block text-slate-300 text-sm font-semibold mb-3">Min Score</label>
                <div className="flex items-center gap-2">
                  <input
                    type="range"
                    min="0"
                    max="10"
                    value={minScoreSidebarSlider}
                    onChange={(e) => setMinScoreSidebarSlider(parseInt(e.target.value))}
                    className="flex-1 h-2 bg-slate-700 rounded cursor-pointer"
                  />
                  <span className="text-white font-bold w-8 text-center">{minScoreSidebarSlider}</span>
                </div>
              </div>

              <div className="border-t border-slate-700 pt-4">
                <label className="block text-slate-300 text-sm font-semibold mb-3">Limit</label>
                <SelectInput
                  value={limit.toString()}
                  onChange={(v) => setLimit(parseInt(v))}
                  options={[
                    { value: '10', label: '10 stocks' },
                    { value: '15', label: '15 stocks' },
                    { value: '25', label: '25 stocks' },
                    { value: '50', label: '50 stocks' },
                    { value: '100', label: '100 stocks' },
                  ]}
                  className="w-full"
                />
              </div>

              <button
                onClick={handleStartScan}
                className="w-full flex items-center justify-center gap-2 bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg transition text-sm font-semibold border-t border-slate-700 pt-4 mt-4"
              >
                <RefreshCw className="w-4 h-4" />
                New Scan
              </button>
              
              {scanTriggered && opportunities.length > 0 && (
                <button
                  onClick={handleLoadNextBatch}
                  disabled={isLoading}
                  className="w-full flex items-center justify-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 text-white px-4 py-2 rounded-lg transition text-sm font-semibold"
                >
                  <TrendingUp className="w-4 h-4" />
                  {isLoading ? 'Loading...' : 'Load Next Batch'}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="lg:col-span-3 space-y-6">
        <div className="flex justify-between items-start mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">Stock Scouter</h1>
            <p className="text-slate-400">
              {opportunities.length} opportunities found • {totalScanned} stocks analyzed
            </p>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gradient-to-br from-blue-900/30 to-blue-800/10 rounded-lg p-6 border border-blue-700/30">
            <div className="flex items-center gap-3 mb-3">
              <Target className="w-6 h-6 text-blue-400" />
              <p className="text-slate-400 text-sm">Opportunities Found</p>
            </div>
            <p className="text-3xl font-bold text-white">{opportunities.length}</p>
            <p className="text-xs text-blue-400 mt-2">Top scoring stocks</p>
          </div>

          <div className="bg-gradient-to-br from-green-900/30 to-green-800/10 rounded-lg p-6 border border-green-700/30">
            <div className="flex items-center gap-3 mb-3">
              <TrendingUp className="w-6 h-6 text-green-400" />
              <p className="text-slate-400 text-sm">Avg Top Score</p>
            </div>
            <p className="text-3xl font-bold text-green-400">
              {opportunities.length > 0 
                ? (opportunities.reduce((sum, opp) => sum + opp.score, 0) / opportunities.length).toFixed(1)
                : '-'}
            </p>
            <p className="text-xs text-green-400 mt-2">Average score</p>
          </div>
        </div>

      {/* Opportunities List */}
      {opportunities.length > 0 ? (
        <ResponsiveTable
          data={opportunities}
          keyExtractor={(opp) => opp.symbol}
          columns={[
            {
              key: 'rank',
              label: 'Rank',
              align: 'center',
              width: '80px',
              render: (val) => (
                <div className="flex items-center justify-center gap-1">
                  <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
                  <span className="text-white font-semibold">{val}</span>
                </div>
              ),
            },
            { key: 'symbol', label: 'Symbol', render: (val) => <div className="font-semibold text-white text-lg">{val}</div> },
            {
              key: 'score',
              label: 'Score',
              align: 'right',
              render: (val) => (
                <span className={`inline-block px-3 py-1 rounded-lg font-bold text-sm ${getScoreBadgeColor(val)}`}>
                  {val.toFixed(1)}
                </span>
              ),
            },
            { key: 'rsi', label: 'RSI', align: 'right', render: (val) => val ? val.toFixed(2) : '-' },
            { key: 'atr', label: 'ATR', align: 'right', render: (val) => val ? val.toFixed(4) : '-' },
            { key: 'analysis', label: 'Analysis', render: (val) => <span className="text-slate-300 text-sm truncate">{val}</span> },
          ]}
          renderActions={(opp) => (
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleAddToWatchlist(opp.symbol);
              }}
              className={`px-3 py-1 rounded text-sm font-semibold transition ${
                addedToWatchlist.has(opp.symbol)
                  ? 'bg-green-600 text-white'
                  : 'bg-blue-600 hover:bg-blue-700 text-white'
              }`}
            >
              {addedToWatchlist.has(opp.symbol) ? '✓ Added' : '+ Watchlist'}
            </button>
          )}
          renderMobileCard={(opp) => (
            <>
              <div className="space-y-2 text-sm mb-3 border-t border-slate-600 pt-3">
                <div className="flex justify-between"><span className="text-slate-400">RSI:</span><span className="text-white font-semibold">{opp.rsi ? opp.rsi.toFixed(2) : '-'}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">ATR:</span><span className="text-white font-semibold">{opp.atr ? opp.atr.toFixed(4) : '-'}</span></div>
                <div className="flex justify-between"><span className="text-slate-400">Analysis:</span><span className="text-white text-xs text-right truncate ml-2">{opp.analysis}</span></div>
              </div>
            </>
          )}
          emptyMessage="No opportunities found"
        />
      ) : (
        <div className="bg-slate-800 rounded-lg p-12 border border-slate-700 text-center">
          <AlertCircle className="w-12 h-12 text-slate-400 mx-auto mb-4" />
          <p className="text-slate-400 mb-4">No opportunities found</p>
          <p className="text-slate-500 text-sm mb-4">Try lowering the minimum score or running another scan</p>
          <button
            onClick={() => refetch()}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm font-semibold transition"
          >
            Scan Again
          </button>
        </div>
      )}

      {/* Loading indicator when fetching next batch */}
      {isLoading && allOpportunities.length > 0 && (
        <div className="border-t border-slate-700 pt-6 flex items-center justify-center gap-3">
          <RefreshCw className="w-5 h-5 text-blue-400 animate-spin" />
          <span className="text-slate-400">Loading next batch...</span>
        </div>
      )}

      {/* Info Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-blue-900/10 border border-blue-700/30 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-3 flex items-center gap-2">
            <Target className="w-5 h-5 text-blue-400" />
            Scoring System
          </h3>
          <ul className="space-y-2 text-slate-300 text-sm">
            <li className="flex gap-2">
              <span className="text-blue-400 font-bold">8-10</span>
              <span>Excellent - Strong buy signals</span>
            </li>
            <li className="flex gap-2">
              <span className="text-yellow-400 font-bold">6-7.9</span>
              <span>Good - Positive indicators</span>
            </li>
            <li className="flex gap-2">
              <span className="text-orange-400 font-bold">5-5.9</span>
              <span>Fair - Moderate potential</span>
            </li>
            <li className="flex gap-2">
              <span className="text-slate-400 font-bold">&lt;5</span>
              <span>Poor - Not recommended</span>
            </li>
          </ul>
        </div>

        <div className="bg-green-900/10 border border-green-700/30 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-3 flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-green-400" />
            Key Metrics
          </h3>
          <ul className="space-y-2 text-slate-300 text-sm">
            <li className="flex justify-between">
              <span>RSI (Oversold)</span>
              <span className="text-green-400">&lt; 35 ideal</span>
            </li>
            <li className="flex justify-between">
              <span>RSI (Overbought)</span>
              <span className="text-green-400">&lt; 75 ideal</span>
            </li>
            <li className="flex justify-between">
              <span>ATR (Volatility)</span>
              <span className="text-green-400">&gt; 0.1 needed</span>
            </li>
            <li className="flex justify-between">
              <span>Volume Ratio</span>
              <span className="text-green-400">&gt; 1.0x needed</span>
            </li>
          </ul>
        </div>
      </div>
      </div>
    </div>
  );
}
