'use client';

import { useState, useEffect } from 'react';
import { Star, Plus, X, RefreshCw, Trash2, Eye } from 'lucide-react';
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
  const [sortBy, setSortBy] = useState<'score' | 'symbol' | 'date'>('score');
  const [filterSymbol, setFilterSymbol] = useState('');

  useEffect(() => {
    fetchWatchlist();
  }, []);

  const fetchWatchlist = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get('/watchlist') as WatchlistData;
      console.log('[Watchlist] API Response:', response);
      if (response) {
        setWatchlistData(response);
      }
    } catch (err: any) {
      console.error('[Watchlist] Full error:', err);
      const errorMessage = err.message || 'Failed to fetch watchlist';
      setError(errorMessage);
      // Fallback mock data for demonstration
      setWatchlistData({
        watchlist: [
          {
            id: 1,
            symbol: 'TSLA',
            asset_type: 'stock',
            score: 87,
            reason: 'Strong bullish signal on daily chart',
            added_date: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
            last_updated: new Date().toISOString(),
          },
          {
            id: 2,
            symbol: 'NVDA',
            asset_type: 'stock',
            score: 92,
            reason: 'AI sector momentum',
            added_date: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
            last_updated: new Date().toISOString(),
          },
          {
            id: 3,
            symbol: 'AAPL',
            asset_type: 'stock',
            score: 65,
            reason: 'Watching for reversal',
            added_date: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
            last_updated: new Date().toISOString(),
          },
          {
            id: 4,
            symbol: 'MSFT',
            asset_type: 'stock',
            score: 78,
            reason: 'Cloud growth potential',
            added_date: new Date(Date.now() - 21 * 24 * 60 * 60 * 1000).toISOString(),
            last_updated: new Date().toISOString(),
          },
        ],
        count: 4,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const addToWatchlist = async () => {
    if (!newSymbol.trim()) return;

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
      setError(err.message || 'Failed to add symbol to watchlist');
    }
  };

  const removeFromWatchlist = async (symbol: string) => {
    if (!confirm(`Remove ${symbol} from watchlist?`)) return;

    try {
      await apiClient.delete('/watchlist', {
        data: { symbol },
      });
      fetchWatchlist();
    } catch (err: any) {
      setError(err.message || 'Failed to remove symbol from watchlist');
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
    if (score >= 80) return 'bg-green-500/20 border-green-500/50 text-green-400';
    if (score >= 60) return 'bg-yellow-500/20 border-yellow-500/50 text-yellow-400';
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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">Watchlist</h1>
          <p className="text-slate-400">Monitor stocks and market opportunities</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={fetchWatchlist}
            disabled={isLoading}
            className="flex items-center gap-2 bg-slate-700 hover:bg-slate-600 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={() => setAddSymbolModal(true)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold transition"
          >
            <Plus className="w-4 h-4" />
            Add Symbol
          </button>
        </div>
      </div>

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
                  onChange={(e) => setNewSymbol(e.target.value.toUpperCase())}
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

              <div className="flex gap-2 justify-end">
                <button
                  onClick={() => setAddSymbolModal(false)}
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-semibold transition"
                >
                  Cancel
                </button>
                <button
                  onClick={addToWatchlist}
                  disabled={!newSymbol.trim()}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white rounded-lg font-semibold transition"
                >
                  Add
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
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700">
                <th className="px-4 py-3 text-left text-slate-300 font-semibold">Symbol</th>
                <th className="px-4 py-3 text-center text-slate-300 font-semibold">Score</th>
                <th className="px-4 py-3 text-left text-slate-300 font-semibold">Reason</th>
                <th className="px-4 py-3 text-center text-slate-300 font-semibold">Added</th>
                <th className="px-4 py-3 text-center text-slate-300 font-semibold">Actions</th>
              </tr>
            </thead>
            <tbody>
              {getSortedSymbols().map((item) => (
                <tr
                  key={item.symbol}
                  className="border-b border-slate-700 hover:bg-slate-700/50 transition"
                >
                  {/* Symbol */}
                  <td className="px-4 py-4">
                    <span className="font-bold text-blue-400">{item.symbol}</span>
                  </td>

                  {/* Score */}
                  <td className="px-4 py-4 text-center">
                    <span className={`inline-block px-3 py-1 rounded-lg font-bold text-sm border ${getScoreBadgeColor(item.score)}`}>
                      {item.score.toFixed(0)}
                    </span>
                  </td>

                  {/* Reason */}
                  <td className="px-4 py-4">
                    <p className="text-sm text-slate-400">
                      {safeValue(item.reason) || 'No reason provided'}
                    </p>
                  </td>

                  {/* Date Added */}
                  <td className="px-4 py-4 text-center">
                    <p className="text-sm text-slate-400">
                      {formatDate(safeValue(item.added_date))}
                    </p>
                  </td>

                  {/* Actions */}
                  <td className="px-4 py-4 text-center">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        title="View details"
                        className="text-blue-400 hover:text-blue-300 transition"
                      >
                        <Eye className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => removeFromWatchlist(item.symbol)}
                        title="Remove from watchlist"
                        className="text-red-400 hover:text-red-300 transition"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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
              {getSortedSymbols().filter((item) => item.score >= 80).length}/{watchlistData.count}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
