'use client';

import React, { useState, useEffect } from 'react';
import { Star, Plus, X, RefreshCw, Trash2, Eye } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import ResponsiveTable from '@/components/Tables/ResponsiveTable';
import { formatDate, safeValue } from '@/lib/formatters';
import apiClient from '@/lib/apiClient';

interface WatchlistItem {
  id?: number;
  symbol: string;
  asset_type?: string;
  score: number;
  reason?: string | null;
  added_date?: string | null;
  last_updated?: string | null;
}

interface WatchlistData {
  watchlist: WatchlistItem[];
  count: number;
  lastUpdated?: string;
}

export default function WatchlistPage() {
  const [watchlistData, setWatchlistData] = useState<WatchlistData>({
    watchlist: [],
    count: 0,
    lastUpdated: new Date().toISOString(),
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [addSymbolModal, setAddSymbolModal] = useState(false);
  const [newSymbol, setNewSymbol] = useState('');
  const [newReason, setNewReason] = useState('');
  const [modalError, setModalError] = useState<string | null>(null);
  const [isAddingSymbol, setIsAddingSymbol] = useState(false);
  const [sortBy, setSortBy] = useState<'score' | 'symbol' | 'date'>('score');
  const [filterSymbol, setFilterSymbol] = useState('');
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null);
  const [analysisLoading, setAnalysisLoading] = useState<string | null>(null);
  const [analysisData, setAnalysisData] = useState<{ [key: string]: any }>({});
  const [isScanning, setIsScanning] = useState(false);

  useEffect(() => {
    // Auto-fetch watchlist when page loads
    fetchWatchlist();
  }, []);

  const fetchWatchlist = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get('/watchlist');
      console.log('[Watchlist] API Response:', response);
      
      // Handle the response - apiClient already extracts .data
      if (response && typeof response === 'object') {
        const data = response as any;
        if (data.watchlist && Array.isArray(data.watchlist)) {
          setWatchlistData({
            watchlist: data.watchlist,
            count: data.count || data.watchlist.length,
            lastUpdated: new Date().toISOString(),
          });
        } else {
          // Empty watchlist
          setWatchlistData({
            watchlist: [],
            count: 0,
            lastUpdated: new Date().toISOString(),
          });
        }
      }
    } catch (err: any) {
      console.error('[Watchlist] Fetch error:', {
        message: err.message,
        status: err.status,
        details: err.details,
        fullError: err,
      });
      const errorMessage = err.message || 'Failed to fetch watchlist';
      setError(errorMessage);
      
      // Only set empty data on error - don't show mock data
      setWatchlistData({
        watchlist: [],
        count: 0,
        lastUpdated: new Date().toISOString(),
      });
    } finally {
      setIsLoading(false);
    }
  };

  const addToWatchlist = async () => {
    if (!newSymbol.trim()) {
      setModalError('Please enter a stock symbol');
      return;
    }

    setIsAddingSymbol(true);
    setModalError(null);

    try {
      await apiClient.post('/watchlist', {
        symbol: newSymbol.toUpperCase(),
        reason: newReason || 'Added from watchlist page',
        score: 50,
      });

      setNewSymbol('');
      setNewReason('');
      setAddSymbolModal(false);
      fetchWatchlist();
    } catch (err: any) {
      // Extract error message from the API response or use the error message
      const errorMessage = err.message || 'Failed to add symbol to watchlist';
      console.error('Add to watchlist error:', {
        message: err.message,
        status: err.status,
        details: err.details,
        fullError: err
      });
      setModalError(errorMessage);
    } finally {
      setIsAddingSymbol(false);
    }
  };

  const removeFromWatchlist = async (symbol: string) => {
    if (!confirm(`Remove ${symbol} from watchlist?`)) return;

    try {
      await apiClient.delete(`/watchlist?symbol=${encodeURIComponent(symbol)}`);
      await fetchWatchlist();
    } catch (err: any) {
      setError(err.message || 'Failed to remove symbol from watchlist');
      console.error('Remove error:', err);
    }
  };

  const analyzeSymbol = async (symbol: string) => {
    if (analysisData[symbol]) {
      setExpandedSymbol(expandedSymbol === symbol ? null : symbol);
      return;
    }

    setAnalysisLoading(symbol);
    try {
      const response = await apiClient.get(`/watchlist/analyze?symbol=${symbol}`);
      setAnalysisData(prev => ({
        ...prev,
        [symbol]: response
      }));
      setExpandedSymbol(symbol);
    } catch (err: any) {
      setError(`Failed to analyze ${symbol}: ${err.message}`);
    } finally {
      setAnalysisLoading(null);
    }
  };

  const scanAllWatchlist = async () => {
    setIsScanning(true);
    setError(null);
    try {
      console.log('[Watchlist] Scanning and updating all scores...');
      const response = await apiClient.put('/watchlist/refresh-scores');
      console.log('[Watchlist] Scores updated:', response);
      // Refetch watchlist to show updated scores
      await fetchWatchlist();
    } catch (err: any) {
      console.error('[Watchlist] Scan error:', err);
      setError(err.message || 'Failed to scan and update scores');
    } finally {
      setIsScanning(false);
    }
  };

  const getSortedSymbols = () => {
    // Ensure watchlist is an array
    const watchlist = Array.isArray(watchlistData?.watchlist) ? watchlistData.watchlist : [];
    let sorted = [...watchlist];

    if (filterSymbol) {
      sorted = sorted.filter((item) =>
        item.symbol.toUpperCase().includes(filterSymbol.toUpperCase())
      );
    }

    switch (sortBy) {
      case 'score':
        sorted.sort((a, b) => b.score - a.score);
        break;
      case 'symbol':
        sorted.sort((a, b) => a.symbol.localeCompare(b.symbol));
        break;
      case 'date':
        sorted.sort((a, b) => {
          const dateA = new Date(a.added_date || 0).getTime();
          const dateB = new Date(b.added_date || 0).getTime();
          return dateB - dateA;
        });
        break;
    }

    return sorted;
  };

  const getScoreBadgeColor = (score: number) => {
    if (score >= 8) return 'bg-green-500/20 border-green-500/50 text-green-400';
    if (score >= 6) return 'bg-yellow-500/20 border-yellow-500/50 text-yellow-400';
    return 'bg-red-500/20 border-red-500/50 text-red-400';
  };

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return 'Unknown';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays}d ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`;
    return date.toLocaleDateString();
  };

  // Safe extraction from database null types
  const safeValue = (value: any): string => {
    if (value === null || value === undefined) return '';
    if (typeof value === 'string') return value;
    if (typeof value === 'object' && value.String !== undefined) return value.String;
    return String(value);
  };

  const getScoreColor = (score: number) => {
    if (score >= 80) return 'text-green-400 bg-green-500/20';
    if (score >= 60) return 'text-yellow-400 bg-yellow-500/20';
    return 'text-red-400 bg-red-500/20';
  };



  return (
    <div className="w-full space-y-8">
      {/* Controls */}
      <div className="flex gap-2 justify-end">
        <button
          onClick={fetchWatchlist}
          disabled={isLoading}
          className="flex items-center gap-2 bg-slate-700 hover:bg-slate-600 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
        <button
          onClick={scanAllWatchlist}
          disabled={isScanning}
          className="flex items-center gap-2 bg-purple-600 hover:bg-purple-700 disabled:bg-slate-600 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <RefreshCw className={`w-4 h-4 ${isScanning ? 'animate-spin' : ''}`} />
          Scan All
        </button>
        <button
          onClick={() => setAddSymbolModal(true)}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <Plus className="w-4 h-4" />
          Add Symbol
        </button>
      </div>

      {/* Header */}
      <PageHeader title="Watchlist" description="Monitor stocks and market opportunities" />

      {/* Add Symbol Modal */}
      {addSymbolModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-white">Add to Watchlist</h2>
              <button
                onClick={() => setAddSymbolModal(false)}
                className="text-slate-400 hover:text-white"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-slate-300 text-sm font-semibold mb-2">
                  Stock Symbol
                </label>
                <input
                  type="text"
                  placeholder="e.g., TSLA"
                  value={newSymbol}
                  onChange={(e) => {
                    setNewSymbol(e.target.value.toUpperCase());
                    setModalError(null);
                  }}
                  onKeyPress={(e) => e.key === 'Enter' && addToWatchlist()}
                  className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                  autoFocus
                />
              </div>

              <div>
                <label className="block text-slate-300 text-sm font-semibold mb-2">
                  Reason (Optional)
                </label>
                <textarea
                  placeholder="Why are you watching this stock?"
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500 resize-none"
                  rows={3}
                />
              </div>

              {modalError && (
                <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-3 text-red-400 text-sm">
                  {modalError}
                </div>
              )}

              <div className="flex gap-2 justify-end">
                <button
                  onClick={() => {
                    setAddSymbolModal(false);
                    setModalError(null);
                  }}
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-semibold transition"
                >
                  Cancel
                </button>
                <button
                  onClick={addToWatchlist}
                  disabled={!newSymbol.trim() || isAddingSymbol}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 text-white rounded-lg font-semibold transition"
                >
                  {isAddingSymbol ? 'Adding...' : 'Add'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-4 text-red-400">
          {error}
        </div>
      )}

      {/* Filters & Sort */}
      <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-2">
              Filter by Symbol
            </label>
            <input
              type="text"
              placeholder="Search symbol..."
              value={filterSymbol}
              onChange={(e) => setFilterSymbol(e.target.value)}
              className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-2">Sort By</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as typeof sortBy)}
              className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white focus:outline-none focus:border-blue-500"
            >
              <option value="score">Score (High to Low)</option>
              <option value="symbol">Symbol (A-Z)</option>
              <option value="date">Date Added (Newest)</option>
            </select>
          </div>
        </div>
      </div>

      {/* Loading State */}
      {isLoading && (!Array.isArray(watchlistData?.watchlist) || watchlistData.watchlist.length === 0) && (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-800 rounded-lg border border-slate-700 p-4 animate-pulse">
              <div className="h-6 bg-slate-700 rounded w-1/4 mb-3"></div>
              <div className="h-4 bg-slate-700 rounded w-full"></div>
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !getSortedSymbols().length && (
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-12 text-center">
          <Star className="w-16 h-16 text-slate-500 mx-auto mb-4 opacity-50" />
          <p className="text-slate-400 mb-4">No symbols in your watchlist yet</p>
          <button
            onClick={() => setAddSymbolModal(true)}
            className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-semibold transition"
          >
            <Plus className="w-4 h-4" />
            Add Your First Symbol
          </button>
        </div>
      )}

      {/* Watchlist Table */}
      {!isLoading && getSortedSymbols().length > 0 && (
        <ResponsiveTable
          data={getSortedSymbols()}
          keyExtractor={(item) => item.symbol}
          columns={[
            { key: 'symbol', label: 'Symbol', render: (val) => <span className="font-bold text-blue-400">{val}</span> },
            {
              key: 'score',
              label: 'Score',
              align: 'center',
              render: (val) => (
                <span className={`inline-block px-3 py-1 rounded-lg font-bold text-sm border ${getScoreBadgeColor(val)}`}>
                  {val.toFixed(1)}
                </span>
              ),
            },
            {
              key: 'reason',
              label: 'Reason',
              render: (val) => <p className="text-sm text-slate-400">{safeValue(val) || 'No reason provided'}</p>,
            },
            {
              key: 'added_date',
              label: 'Added',
              align: 'center',
              render: (val) => <p className="text-sm text-slate-400">{formatDate(safeValue(val))}</p>,
              mobileHidden: true,
            },
          ]}
          renderActions={(item) => (
            <div className="flex items-center justify-center gap-2">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  analyzeSymbol(item.symbol);
                }}
                title={expandedSymbol === item.symbol ? 'Hide details' : 'View details'}
                className={`transition ${
                  expandedSymbol === item.symbol ? 'text-blue-300' : 'text-blue-400 hover:text-blue-300'
                }`}
              >
                <Eye className="w-4 h-4" />
              </button>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  removeFromWatchlist(item.symbol);
                }}
                title="Remove from watchlist"
                className="text-red-400 hover:text-red-300 transition"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          )}
          renderMobileCard={(item) => (
            <div className="space-y-2 text-sm">
              <div className="flex justify-between"><span className="text-slate-400">Score:</span><span className={`font-bold px-2 py-1 rounded ${getScoreBadgeColor(item.score)}`}>{item.score.toFixed(1)}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Reason:</span><span className="text-white text-xs text-right ml-2">{safeValue(item.reason) || 'N/A'}</span></div>
            </div>
          )}
          emptyMessage="No symbols in watchlist"
        />
      )}

      {/* Expanded Analysis Section - Outside Table */}
      {!isLoading && expandedSymbol && analysisData[expandedSymbol] && (
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <div className="mb-4 flex justify-between items-center">
            <h3 className="text-xl font-bold text-white">{expandedSymbol} Analysis</h3>
            <button
              onClick={() => setExpandedSymbol(null)}
              className="text-sm text-blue-400 hover:text-blue-300"
            >
              Hide Details ✕
            </button>
          </div>
          
          {analysisLoading ? (
            <div className="flex items-center justify-center gap-2 text-slate-400 py-8">
              <RefreshCw className="w-4 h-4 animate-spin" />
              Analyzing {expandedSymbol}...
            </div>
          ) : (
            <div className="space-y-4">
              {/* Core Metrics Grid */}
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <div className="bg-slate-700 rounded p-3">
                  <p className="text-slate-400 text-xs font-semibold mb-1">Price</p>
                  <p className="text-white font-bold text-lg">
                    ${analysisData[expandedSymbol].current_price?.toFixed(2) || 'N/A'}
                  </p>
                </div>
                <div className="bg-slate-700 rounded p-3">
                  <p className="text-slate-400 text-xs font-semibold mb-1">RSI</p>
                  <p className="text-white font-bold text-lg">
                    {analysisData[expandedSymbol].rsi?.toFixed(1) || 'N/A'}
                  </p>
                  <p className="text-xs text-slate-400 mt-1">
                    {analysisData[expandedSymbol].rsi ? (
                      analysisData[expandedSymbol].rsi! > 70 ? '🔴 Overbought' :
                      analysisData[expandedSymbol].rsi! < 30 ? '🟢 Oversold' : '🟡 Neutral'
                    ) : 'N/A'}
                  </p>
                </div>
                <div className="bg-slate-700 rounded p-3">
                  <p className="text-slate-400 text-xs font-semibold mb-1">ATR</p>
                  <p className="text-white font-bold text-lg">
                    {analysisData[expandedSymbol].atr?.toFixed(2) || 'N/A'}
                  </p>
                </div>
                <div className="bg-slate-700 rounded p-3">
                  <p className="text-slate-400 text-xs font-semibold mb-1">SMA 20</p>
                  <p className="text-white font-bold text-lg">
                    ${analysisData[expandedSymbol].sma_20?.toFixed(2) || 'N/A'}
                  </p>
                </div>
                <div className="bg-slate-700 rounded p-3">
                  <p className="text-slate-400 text-xs font-semibold mb-1">Trend</p>
                  <p className={`font-bold text-lg ${
                    analysisData[expandedSymbol].trend === 'bullish' ? 'text-green-400' :
                    analysisData[expandedSymbol].trend === 'bearish' ? 'text-red-400' : 'text-yellow-400'
                  }`}>
                    {analysisData[expandedSymbol].trend?.toUpperCase() || 'N/A'}
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Summary Stats */}
      {!isLoading && getSortedSymbols().length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
            <p className="text-slate-400 text-sm font-semibold mb-2">Total Symbols</p>
            <p className="text-2xl font-bold text-white">{watchlistData.count}</p>
          </div>

          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
            <p className="text-slate-400 text-sm font-semibold mb-2">Avg Score</p>
            <p className="text-2xl font-bold text-white">
              {watchlistData.count > 0 ? (
                (
                  getSortedSymbols().reduce((sum, item) => sum + item.score, 0) /
                  getSortedSymbols().length
                ).toFixed(1)
              ) : (
                '0'
              )}
            </p>
          </div>

          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
            <p className="text-slate-400 text-sm font-semibold mb-2">High Scoring</p>
            <p className="text-2xl font-bold text-green-400">
              {getSortedSymbols().filter((item) => item.score >= 8).length}/{watchlistData.count}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
