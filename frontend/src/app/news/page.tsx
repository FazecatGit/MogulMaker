'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, TrendingUp, TrendingDown, Calendar, ExternalLink, RefreshCw } from 'lucide-react';
import apiClient from '@/lib/apiClient';

interface NewsItem {
  id: string;
  symbol: string;
  title: string;
  summary: string;
  source: string;
  publishedAt: string;
  sentiment: 'bullish' | 'bearish' | 'neutral';
  relevanceScore: number;
  url: string;
}

interface NewsData {
  news: NewsItem[];
  lastUpdated: string;
  count: number;
}

export default function NewsPage() {
  const [newsData, setNewsData] = useState<NewsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSentiment, setSelectedSentiment] = useState<'all' | 'bullish' | 'bearish' | 'neutral'>('all');
  const [searchSymbol, setSearchSymbol] = useState('');

  useEffect(() => {
    fetchNews();
  }, []);

  const fetchNews = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get('/analysis/report?symbol=*');
      if (response) {
        // Handle both array and object responses
        const data = response.data;
        const news = Array.isArray(data) ? data : data.news || data;
        setNewsData(Array.isArray(news) ? news : []);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to fetch news');
      // Fallback mock data for demonstration
      setNewsData([
        {
          id: '1',
          symbol: 'TSLA',
          title: 'Tesla Q4 Earnings Beat Expectations',
          summary: 'Tesla reported stronger than expected Q4 earnings with improved margins.',
          source: 'Bloomberg',
          publishedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
          sentiment: 'bullish',
          relevanceScore: 0.92,
          url: '#',
        },
        {
          id: '2',
          symbol: 'AAPL',
          title: 'Apple Faces Supply Chain Challenges',
          summary: 'Supply chain disruptions could impact Q1 2025 iPhone production.',
          source: 'Reuters',
          publishedAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
          sentiment: 'bearish',
          relevanceScore: 0.78,
          url: '#',
        },
        {
          id: '3',
          symbol: 'MSFT',
          title: 'Microsoft Cloud Revenue Grows 30%',
          summary: 'Azure cloud services continue strong growth trajectory.',
          source: 'MarketWatch',
          publishedAt: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(),
          sentiment: 'bullish',
          relevanceScore: 0.85,
          url: '#',
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const filteredNews = newsData.filter(item => {
    const sentimentMatch = selectedSentiment === 'all' || item.sentiment === selectedSentiment;
    const symbolMatch = !searchSymbol || item.symbol.toUpperCase().includes(searchSymbol.toUpperCase());
    return sentimentMatch && symbolMatch;
  });

  const getSentimentColor = (sentiment: string) => {
    switch (sentiment) {
      case 'bullish':
        return 'bg-green-500/20 border-green-500/50 text-green-400';
      case 'bearish':
        return 'bg-red-500/20 border-red-500/50 text-red-400';
      default:
        return 'bg-slate-500/20 border-slate-500/50 text-slate-400';
    }
  };

  const getSentimentIcon = (sentiment: string) => {
    switch (sentiment) {
      case 'bullish':
        return <TrendingUp className="w-4 h-4" />;
      case 'bearish':
        return <TrendingDown className="w-4 h-4" />;
      default:
        return <AlertCircle className="w-4 h-4" />;
    }
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">News & Sentiment</h1>
            <p className="text-slate-400">Latest market news with sentiment analysis</p>
          </div>
          <button
            onClick={fetchNews}
            disabled={isLoading}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 space-y-4">
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
              {(['all', 'bullish', 'bearish', 'neutral'] as const).map((sentiment) => (
                <button
                  key={sentiment}
                  onClick={() => setSelectedSentiment(sentiment)}
                  className={`px-3 py-2 rounded-lg font-semibold transition capitalize ${
                    selectedSentiment === sentiment
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  {sentiment}
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
        {filteredNews.map((item) => (
          <div
            key={item.id}
            className="bg-slate-800 rounded-lg border border-slate-700 p-5 hover:border-slate-600 transition"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                {/* Header Row */}
                <div className="flex items-center gap-3 mb-2">
                  <span className="text-sm font-bold text-blue-400 bg-blue-500/20 px-2 py-1 rounded">
                    {item.symbol}
                  </span>
                  <span
                    className={`flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded border ${getSentimentColor(
                      item.sentiment
                    )}`}
                  >
                    {getSentimentIcon(item.sentiment)}
                    {item.sentiment.charAt(0).toUpperCase() + item.sentiment.slice(1)}
                  </span>
                  <span className="text-xs text-slate-400">
                    {(item.relevanceScore * 100).toFixed(0)}% relevant
                  </span>
                </div>

                {/* Title */}
                <h3 className="text-lg font-bold text-white mb-2">{item.title}</h3>

                {/* Summary */}
                <p className="text-slate-400 text-sm mb-3">{item.summary}</p>

                {/* Footer */}
                <div className="flex items-center justify-between text-xs text-slate-500">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">{item.source}</span>
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3 h-3" />
                      {formatTime(item.publishedAt)}
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
