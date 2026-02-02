'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, TrendingUp, RefreshCw, Target, Zap, Star, Plus, X, Play } from 'lucide-react';
import { useScout } from '@/hooks/useScout';
import apiClient from '@/lib/apiClient';

export default function ScouterPage() {
  const [minScore, setMinScore] = useState(50);
  const [limit, setLimit] = useState(15);
  const [scanTriggered, setScanTriggered] = useState(false);
  const [offset, setOffset] = useState(0);
  const [allOpportunities, setAllOpportunities] = useState<any[]>([]);
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null);
  
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
    console.log('[ScouterPage] Starting fresh scan...');
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
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Stock Scouter</h1>
          <p className="text-slate-400">AI-powered stock screening and opportunity detection</p>
        </div>

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
                max="100"
                value={minScore}
                onChange={(e) => setMinScore(parseInt(e.target.value))}
                className="flex-1 h-2 bg-slate-700 rounded cursor-pointer"
              />
              <span className="text-white font-bold text-lg w-12">{minScore}</span>
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

  // Loading state
  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Stock Scouter</h1>
          <p className="text-slate-400">Scanning stocks...</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-800 rounded-lg p-4 border border-slate-700 animate-pulse h-24" />
          ))}
        </div>

        <div className="space-y-2">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="bg-slate-800 rounded-lg h-20 animate-pulse border border-slate-700"
            />
          ))}
        </div>
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
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Stock Scouter</h1>
          <p className="text-slate-400">AI-powered stock screening and opportunity detection</p>
        </div>

        <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 space-y-3">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
            <div>
              <p className="text-red-400 font-semibold">Failed to scan stocks</p>
              <p className="text-red-300 text-sm">
                {errorMessage}{errorStatus}{errorCode}
              </p>
            </div>
          </div>
          
          <div className="text-xs text-red-400 mt-2 p-3 bg-red-950/30 rounded font-mono">
            Check browser console for more details
          </div>
          
          <button
            onClick={() => setScanTriggered(false)}
            className="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm font-semibold"
          >
            Try Again
          </button>
        </div>
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
                    max="100"
                    value={minScore}
                    onChange={(e) => setMinScore(parseInt(e.target.value))}
                    className="flex-1 h-2 bg-slate-700 rounded cursor-pointer"
                  />
                  <span className="text-white font-bold w-8 text-center">{minScore}</span>
                </div>
              </div>

              <div className="border-t border-slate-700 pt-4">
                <label className="block text-slate-300 text-sm font-semibold mb-3">Limit</label>
                <select
                  value={limit}
                  onChange={(e) => setLimit(parseInt(e.target.value))}
                  className="w-full bg-slate-700 text-white rounded px-3 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none text-sm"
                >
                  <option value="10">10 stocks</option>
                  <option value="15">15 stocks</option>
                  <option value="25">25 stocks</option>
                  <option value="50">50 stocks</option>
                  <option value="100">100 stocks</option>
                </select>
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
        <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
          {/* Desktop View - Table */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50 border-b border-slate-700">
                <tr>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Rank
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-slate-300">
                    Symbol
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    Score
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    RSI
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-semibold text-slate-300">
                    ATR
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-slate-300">
                    Analysis
                  </th>
                  <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {opportunities.map((opp) => (
                  <>
                    <tr key={opp.symbol}
                      className="hover:bg-slate-700/30 transition cursor-pointer"
                      onClick={() => setExpandedSymbol(expandedSymbol === opp.symbol ? null : opp.symbol)}
                    >
                      <td className="px-6 py-4 text-center">
                        <div className="flex items-center justify-center gap-1">
                          <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
                          <span className="text-white font-semibold">{opp.rank}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="font-semibold text-white text-lg">{opp.symbol}</div>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className={`px-3 py-1 rounded text-sm font-bold ${
                          opp.score >= 80 ? 'bg-green-900/30 text-green-400' : 
                          opp.score >= 60 ? 'bg-yellow-900/30 text-yellow-400' : 
                          'bg-orange-900/30 text-orange-400'
                        }`}>
                          {opp.score.toFixed(1)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        {opp.rsi ? opp.rsi.toFixed(2) : '-'}
                      </td>
                      <td className="px-6 py-4 text-right text-slate-300">
                        {opp.atr ? opp.atr.toFixed(4) : '-'}
                      </td>
                      <td className="px-6 py-4 text-slate-300 text-sm truncate">
                        {opp.analysis}
                      </td>
                      <td className="px-6 py-4 text-center">
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
                      </td>
                    </tr>
                    
                    {/* Expanded Details */}
                    {expandedSymbol === opp.symbol && (
                      <tr className="bg-slate-700/20">
                        <td colSpan={7} className="px-6 py-6">
                          <div className="space-y-4">
                            <div>
                              <h4 className="text-sm font-semibold text-slate-300 mb-2">Analysis</h4>
                              <p className="text-white text-sm leading-relaxed">{opp.analysis}</p>
                            </div>
                            
                            <div className="grid grid-cols-3 gap-4">
                              <div className="bg-slate-800/50 rounded p-3 border border-slate-700">
                                <p className="text-xs text-slate-400 mb-1">Score</p>
                                <p className={`text-lg font-bold ${
                                  opp.score >= 80 ? 'text-green-400' : 
                                  opp.score >= 60 ? 'text-yellow-400' : 
                                  'text-orange-400'
                                }`}>
                                  {opp.score.toFixed(2)}
                                </p>
                              </div>
                              
                              <div className="bg-slate-800/50 rounded p-3 border border-slate-700">
                                <p className="text-xs text-slate-400 mb-1">RSI</p>
                                <p className="text-lg font-bold text-blue-400">
                                  {opp.rsi ? opp.rsi.toFixed(2) : 'N/A'}
                                </p>
                                <p className="text-xs text-slate-500 mt-1">
                                  {opp.rsi && opp.rsi < 35 ? '🔴 Oversold' : opp.rsi && opp.rsi > 75 ? '🟢 Overbought' : '⚪ Neutral'}
                                </p>
                              </div>
                              
                              <div className="bg-slate-800/50 rounded p-3 border border-slate-700">
                                <p className="text-xs text-slate-400 mb-1">ATR</p>
                                <p className="text-lg font-bold text-purple-400">
                                  {opp.atr ? opp.atr.toFixed(4) : 'N/A'}
                                </p>
                                <p className="text-xs text-slate-500 mt-1">Volatility</p>
                              </div>
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile View - Cards */}
          <div className="md:hidden space-y-3 p-4">
            {opportunities.map((opp) => (
              <div
                key={opp.symbol}
                className="bg-slate-700/50 rounded-lg p-4 border border-slate-600 cursor-pointer hover:border-blue-500 transition"
                onClick={() => setSelectedSymbol(selectedSymbol === opp.symbol ? null : opp.symbol)}
              >
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
                      <span className="text-white font-bold">#{opp.rank}</span>
                    </div>
                    <div className="font-semibold text-white text-lg">{opp.symbol}</div>
                  </div>
                  <span className={`px-3 py-1 rounded text-sm font-bold ${
                    opp.score >= 80 ? 'bg-green-900/30 text-green-400' : 
                    opp.score >= 60 ? 'bg-yellow-900/30 text-yellow-400' : 
                    'bg-orange-900/30 text-orange-400'
                  }`}>
                    {opp.score.toFixed(1)}
                  </span>
                </div>

                <div className="space-y-2 text-sm mb-3 border-t border-slate-600 pt-3">
                  <div className="flex justify-between">
                    <span className="text-slate-400">RSI:</span>
                    <span className="text-white font-semibold">{opp.rsi ? opp.rsi.toFixed(2) : '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">ATR:</span>
                    <span className="text-white font-semibold">{opp.atr ? opp.atr.toFixed(4) : '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Analysis:</span>
                    <span className="text-white text-xs text-right truncate ml-2">{opp.analysis}</span>
                  </div>
                </div>

                <button 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleAddToWatchlist(opp.symbol);
                  }}
                  className={`w-full px-3 py-2 rounded text-sm font-semibold transition ${
                    addedToWatchlist.has(opp.symbol)
                      ? 'bg-green-600 text-white'
                      : 'bg-blue-600 hover:bg-blue-700 text-white'
                  }`}
                >
                  {addedToWatchlist.has(opp.symbol) ? '✓ Added to Watchlist' : '+ Add to Watchlist'}
                </button>
              </div>
            ))}
          </div>
        </div>
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

      {/* Info Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-blue-900/10 border border-blue-700/30 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-white mb-3 flex items-center gap-2">
            <Target className="w-5 h-5 text-blue-400" />
            Scoring System
          </h3>
          <ul className="space-y-2 text-slate-300 text-sm">
            <li className="flex gap-2">
              <span className="text-blue-400 font-bold">80+</span>
              <span>Excellent - Strong buy signals</span>
            </li>
            <li className="flex gap-2">
              <span className="text-yellow-400 font-bold">60-79</span>
              <span>Good - Positive indicators</span>
            </li>
            <li className="flex gap-2">
              <span className="text-orange-400 font-bold">50-59</span>
              <span>Fair - Moderate potential</span>
            </li>
            <li className="flex gap-2">
              <span className="text-slate-400 font-bold">&lt;50</span>
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


