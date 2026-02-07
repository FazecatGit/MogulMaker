'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, TrendingUp, TrendingDown, Calendar, ExternalLink, RefreshCw, Zap } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import { formatTime } from '@/lib/formatters';
import apiClient from '@/lib/apiClient';

interface NewsItem {
  id: number;
  symbol: string;
  headline: string;
  url: string;
  published_at: string;
  source: string;
  sentiment: string;
  catalyst: string;
  impact: number;
}

interface NewsData {
  news: NewsItem[];
  count: number;
  symbols_tracked: number;
  message?: string;
}

export default function NewsPage() {
  const [newsData, setNewsData] = useState<NewsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSentiment, setSelectedSentiment] = useState<string>('all');
  const [searchSymbol, setSearchSymbol] = useState('');

  // Auto-fetch news when page loads
  useEffect(() => {
    fetchNews();
  }, []);

  const fetchNews = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await apiClient.get('/news') as NewsData;
      console.log('News data received:', data);
      if (data) {
        setNewsData(data.news || []);
      }
    } catch (err: any) {
      console.error('News fetch error:', err);
      setError(err.message || 'Failed to fetch news');
      setNewsData([]);
    } finally {
      setIsLoading(false);
    }
  };

  const filteredNews = newsData.filter(item => {
    const sentimentMatch = selectedSentiment === 'all' || item.sentiment.toUpperCase() === selectedSentiment.toUpperCase();
    const symbolMatch = !searchSymbol || item.symbol.toUpperCase().includes(searchSymbol.toUpperCase());
    return sentimentMatch && symbolMatch;
  });

  const getSentimentColor = (sentiment: string) => {
    const sentimentUpper = sentiment.toUpperCase();
    switch (sentimentUpper) {
      case 'POSITIVE':
        return 'bg-green-500/20 border-green-500/50 text-green-400';
      case 'NEGATIVE':
        return 'bg-red-500/20 border-red-500/50 text-red-400';
      default:
        return 'bg-slate-500/20 border-slate-500/50 text-slate-400';
    }
  };

  const getSentimentIcon = (sentiment: string) => {
    const sentimentUpper = sentiment.toUpperCase();
    switch (sentimentUpper) {
      case 'POSITIVE':
        return <TrendingUp className="w-4 h-4" />;
      case 'NEGATIVE':
        return <TrendingDown className="w-4 h-4" />;
      default:
        return <AlertCircle className="w-4 h-4" />;
    }
  };

  return (
    <div className="w-full space-y-8">
      {/* Controls */}
      <div className="flex justify-end">
        <button
          onClick={fetchNews}
          disabled={isLoading}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>
      
      {/* Header */}
      <PageHeader title="News & Sentiment" description="Latest market news with sentiment analysis" />

      {/* Filters */}
      <div className="control-panel">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Symbol Filter */}
          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-2">Filter by Symbol</label>
            <input
              type="text"
              placeholder="e.g., TSLA, AAPL"
              value={searchSymbol}
              onChange={(e) => setSearchSymbol(e.target.value)}
              className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
            />
          </div>

          {/* Sentiment Filter */}
          <div>
            <label className="block text-slate-300 text-sm font-semibold mb-2">Filter by Sentiment</label>
            <div className="flex gap-2">
              {(['all', 'POSITIVE', 'NEGATIVE', 'NEUTRAL'] as const).map((sentiment) => (
                <button
                  key={sentiment}
                  onClick={() => setSelectedSentiment(sentiment === 'all' ? 'all' : sentiment)}
                  className={`px-3 py-2 rounded-lg font-semibold transition capitalize ${
                    selectedSentiment === sentiment
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  {sentiment === 'all' ? 'All' : sentiment.charAt(0) + sentiment.slice(1).toLowerCase()}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Error State */}
      {error && !newsData.length && (
        <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-4 text-red-400">
          <p>{error}</p>
        </div>
      )}

      {/* Loading State */}
      {isLoading && !newsData.length && (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-slate-800 rounded-lg border border-slate-700 p-4 animate-pulse">
              <div className="h-5 bg-slate-700 rounded w-3/4 mb-3"></div>
              <div className="h-4 bg-slate-700 rounded w-full mb-2"></div>
              <div className="h-4 bg-slate-700 rounded w-5/6"></div>
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !filteredNews.length && (
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-8 text-center">
          <AlertCircle className="w-12 h-12 text-slate-500 mx-auto mb-4" />
          <p className="text-slate-400">No news articles found</p>
        </div>
      )}

      {/* News List */}
      <div className="space-y-4">
        {filteredNews.map((item, index) => (
          <div
            key={`${item.symbol}-${item.url}-${index}`}
            className="bg-slate-800 rounded-lg border border-slate-700 p-5 hover:border-slate-600 transition"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                {/* Header Row */}
                <div className="flex items-center gap-3 mb-2 flex-wrap">
                  <span className="text-sm font-bold text-blue-400 bg-blue-500/20 px-2 py-1 rounded">
                    {item.symbol}
                  </span>
                  <span
                    className={`flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded border ${getSentimentColor(
                      item.sentiment
                    )}`}
                  >
                    {getSentimentIcon(item.sentiment)}
                    {item.sentiment.charAt(0).toUpperCase() + item.sentiment.slice(1).toLowerCase()}
                  </span>
                  {item.catalyst && item.catalyst !== 'NO_CATALYST' && (
                    <span className="text-xs font-semibold px-2 py-1 rounded border bg-purple-500/20 border-purple-500/50 text-purple-400 flex items-center gap-1">
                      <Zap className="w-3 h-3" />
                      {item.catalyst}
                    </span>
                  )}
                  {item.impact > 0 && (
                    <span className="text-xs text-slate-400">
                      Impact: {(item.impact * 100).toFixed(0)}%
                    </span>
                  )}
                </div>

                {/* Title */}
                <h3 className="text-lg font-bold text-white mb-2">{item.headline}</h3>

                {/* Footer */}
                <div className="flex items-center justify-between text-xs text-slate-500">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">{item.source}</span>
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3 h-3" />
                      {formatTime(item.published_at)}
                    </span>
                  </div>
                </div>
              </div>

              {/* Read More Button */}
              <a
                href={item.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 hover:text-blue-300 transition flex-shrink-0"
              >
                <ExternalLink className="w-5 h-5" />
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
