'use client';

import React, { useState } from 'react';
import { RefreshCw, Search, TrendingUp, TrendingDown, ZoomIn, ZoomOut } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import apiClient from '@/lib/apiClient';

export default function AnalyzerPage() {
  const [symbol, setSymbol] = useState('');
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [analysisData, setAnalysisData] = useState<any | null>(null);
  const [buyingSymbol, setBuyingSymbol] = useState<string | null>(null);
  const [sellingSymbol, setSellingSymbol] = useState<string | null>(null);

  const analyzeStock = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!inputValue.trim()) {
      setError('Please enter a stock symbol');
      return;
    }

    const stockSymbol = inputValue.toUpperCase().trim();
    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.get(`/watchlist/analyze?symbol=${stockSymbol}`);
      setAnalysisData(response);
      setSymbol(stockSymbol);
    } catch (err: any) {
      setError(err.message || `Failed to analyze ${stockSymbol}`);
      setAnalysisData(null);
    } finally {
      setIsLoading(false);
    }
  };

  const clearAnalysis = () => {
    setSymbol('');
    setInputValue('');
    setAnalysisData(null);
    setError(null);
  };

  const handleBuyStock = async () => {
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
      // Refresh analysis to show any changes
      await analyzeStock(new Event('submit') as any);
    } catch (err) {
      console.error('Failed to buy:', err);
      alert(`Failed to buy ${symbol}`);
    } finally {
      setBuyingSymbol(null);
    }
  };

  const handleSellStock = async () => {
    const quantityStr = prompt(`How many shares of ${symbol} do you want to sell?`, '1');
    if (!quantityStr || isNaN(parseFloat(quantityStr)) || parseFloat(quantityStr) <= 0) {
      return;
    }

    setSellingSymbol(symbol);
    try {
      await apiClient.post('/trades', {
        symbol,
        side: 'sell',
        quantity: parseFloat(quantityStr),
      });
      // Refresh analysis to show any changes
      await analyzeStock(new Event('submit') as any);
    } catch (err) {
      console.error('Failed to sell:', err);
      alert(`Failed to sell ${symbol}`);
    } finally {
      setSellingSymbol(null);
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 text-white p-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
      <PageHeader 
        title="Stock Analyzer" 
        description="Deep dive analysis of any stock symbol"
      />
        {/* Search Section */}
        <form onSubmit={analyzeStock} className="mb-8">
          <div className="flex gap-2">
            <div className="flex-1 relative">
              <input
                type="text"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                placeholder="Enter stock symbol (e.g., AAPL, NVDA, TSLA)"
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              />
              <Search className="absolute right-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-slate-500" />
            </div>
            <button
              type="submit"
              disabled={isLoading}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 disabled:cursor-not-allowed rounded-lg px-6 py-3 font-semibold flex items-center gap-2 transition"
            >
              {isLoading && <RefreshCw className="w-4 h-4 animate-spin" />}
              {isLoading ? 'Analyzing...' : 'Analyze'}
            </button>
          </div>
        </form>

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/30 border border-red-700 rounded-lg p-4 mb-6 text-red-300">
            {error}
          </div>
        )}

        {/* Analysis Results */}
        {analysisData && (
          <div className="space-y-6">
            {/* Header with Symbol and Clear Button */}
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-3xl font-bold">{symbol} Analysis</h2>
              <div className="flex gap-2">
                <button
                  onClick={handleBuyStock}
                  disabled={buyingSymbol === symbol}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed rounded font-semibold text-white transition"
                >
                  {buyingSymbol === symbol ? 'Buying...' : 'Buy'}
                </button>
                <button
                  onClick={handleSellStock}
                  disabled={sellingSymbol === symbol}
                  className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed rounded font-semibold text-white transition"
                >
                  {sellingSymbol === symbol ? 'Selling...' : 'Sell'}
                </button>
                <button
                  onClick={clearAnalysis}
                  className="text-slate-400 hover:text-white transition px-2"
                >
                  Clear
                </button>
              </div>
            </div>

            {/* Core Metrics Grid */}
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              {/* Current Price */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">Price</p>
                <p className="text-white font-bold text-2xl">
                  ${analysisData.current_price?.toFixed(2) || 'N/A'}
                </p>
              </div>

              {/* RSI */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">RSI</p>
                <p className="text-white font-bold text-2xl">
                  {analysisData.rsi?.toFixed(1) || 'N/A'}
                </p>
                <p className="text-xs text-slate-400 mt-2">
                  {analysisData.rsi ? (
                    analysisData.rsi > 70
                      ? '🔴 Overbought'
                      : analysisData.rsi < 30
                        ? '🟢 Oversold'
                        : '🟡 Neutral'
                  ) : (
                    'N/A'
                  )}
                </p>
              </div>

              {/* ATR (Raw Value) */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">ATR</p>
                <p className="text-white font-bold text-2xl">
                  ${analysisData.atr?.toFixed(2) || 'N/A'}
                </p>
                <p className="text-xs text-slate-400 mt-2">
                  14-period avg range
                </p>
              </div>

              {/* Volatility % */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">Volatility %</p>
                <p className="text-white font-bold text-2xl">
                  {analysisData.atr && analysisData.current_price ? (
                    `${((analysisData.atr / analysisData.current_price) * 100).toFixed(2)}%`
                  ) : (
                    'N/A'
                  )}
                </p>
                <p className="text-xs text-slate-400 mt-2">
                  {analysisData.atr && analysisData.current_price ? (
                    (() => {
                      const atrPercent = (analysisData.atr / analysisData.current_price) * 100;
                      if (atrPercent > 3) return '🔴 High';
                      if (atrPercent > 1.5) return '🟡 Medium';
                      return '🟢 Low';
                    })()
                  ) : (
                    'N/A'
                  )}
                </p>
              </div>

              {/* SMA 20 */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">SMA 20</p>
                <p className="text-white font-bold text-2xl">
                  ${analysisData.sma_20?.toFixed(2) || 'N/A'}
                </p>
              </div>

              {/* Trend */}
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
                <p className="text-slate-400 text-xs font-semibold mb-2 uppercase">Trend</p>
                <div className="flex items-center gap-2">
                  {analysisData.trend === 'bullish' ? (
                    <TrendingUp className="w-5 h-5 text-green-400" />
                  ) : analysisData.trend === 'bearish' ? (
                    <TrendingDown className="w-5 h-5 text-red-400" />
                  ) : (
                    <div className="w-5 h-5 text-yellow-400">→</div>
                  )}
                  <p className={`font-bold text-lg ${
                    analysisData.trend === 'bullish'
                      ? 'text-green-400'
                      : analysisData.trend === 'bearish'
                        ? 'text-red-400'
                        : 'text-yellow-400'
                  }`}>
                    {analysisData.trend?.toUpperCase() || 'N/A'}
                  </p>
                </div>
              </div>
            </div>

            {/* Historical Price Chart */}
            {analysisData.historical_bars && analysisData.historical_bars.length > 0 && (
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
                <div className="flex justify-between items-center mb-2">
                  <h3 className="text-xl font-bold">Price History ({analysisData.historical_bars.length} days)</h3>
                  <div className="text-sm text-slate-400">
                    {analysisData.historical_bars.length} trading days
                  </div>
                </div>
                
                {/* Chart Container */}
                <div className="bg-slate-900/50 rounded p-4 overflow-x-auto mb-6">
                  <ResponsiveContainer width="100%" height={400}>
                    <LineChart
                      data={analysisData.historical_bars.map((bar: any) => ({
                        ...bar,
                        date: new Date(bar.timestamp * 1000).toLocaleDateString('en-US', { 
                          month: 'short', 
                          day: 'numeric' 
                        }),
                        fullDate: new Date(bar.timestamp * 1000).toLocaleDateString(),
                      }))}
                      margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
                    >
                      <defs>
                        <linearGradient id="closeGradient" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#22c55e" stopOpacity={0.8}/>
                          <stop offset="95%" stopColor="#22c55e" stopOpacity={0}/>
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                      <XAxis 
                        dataKey="date" 
                        stroke="#64748b"
                        tick={{ fontSize: 12 }}
                        interval={Math.floor(analysisData.historical_bars.length / 10)}
                      />
                      <YAxis 
                        stroke="#64748b"
                        tick={{ fontSize: 12 }}
                        domain={['dataMin - 5', 'dataMax + 5']}
                        label={{ value: 'Price ($)', angle: -90, position: 'insideLeft' }}
                      />
                      <Tooltip 
                        contentStyle={{
                          backgroundColor: '#1e293b',
                          border: '1px solid #475569',
                          borderRadius: '6px',
                        }}
                        labelStyle={{ color: '#e2e8f0' }}
                        formatter={(value: any, name: string | undefined) => {
                          if (!name) return [value, name];
                          if (['open', 'high', 'low', 'close'].includes(name)) {
                            return [`$${value.toFixed(2)}`, name.toUpperCase()];
                          }
                          if (name === 'volume') {
                            return [`${(value / 1000000).toFixed(1)}M`, 'Volume'];
                          }
                          if (['rsi', 'atr'].includes(name)) {
                            return [value.toFixed(2), name.toUpperCase()];
                          }
                          return [value, name];
                        }}
                        cursor={{ stroke: '#64748b', strokeDasharray: '5 5' }}
                      />
                      <Legend 
                        wrapperStyle={{ paddingTop: '20px' }}
                        iconType="line"
                      />
                      <Line 
                        type="monotone" 
                        dataKey="close" 
                        stroke="#22c55e" 
                        dot={false}
                        strokeWidth={2}
                        name="Close"
                        isAnimationActive={false}
                      />
                      <Line 
                        type="monotone" 
                        dataKey="high" 
                        stroke="#6b7280" 
                        dot={false}
                        strokeWidth={1}
                        name="High"
                        strokeDasharray="5 5"
                        isAnimationActive={false}
                      />
                      <Line 
                        type="monotone" 
                        dataKey="low" 
                        stroke="#6b7280" 
                        dot={false}
                        strokeWidth={1}
                        name="Low"
                        strokeDasharray="5 5"
                        isAnimationActive={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                {/* Price Statistics */}
                <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mt-6 text-sm">
                  <div className="bg-slate-700/30 rounded p-3">
                    <p className="text-slate-400 text-xs mb-1">Highest</p>
                    <p className="font-bold text-green-400 text-lg">
                      ${Math.max(...analysisData.historical_bars.map((b: any) => b.high)).toFixed(2)}
                    </p>
                  </div>
                  <div className="bg-slate-700/30 rounded p-3">
                    <p className="text-slate-400 text-xs mb-1">Lowest</p>
                    <p className="font-bold text-red-400 text-lg">
                      ${Math.min(...analysisData.historical_bars.map((b: any) => b.low)).toFixed(2)}
                    </p>
                  </div>
                  <div className="bg-slate-700/30 rounded p-3">
                    <p className="text-slate-400 text-xs mb-1">Avg Volume</p>
                    <p className="font-bold text-white text-lg">
                      {(analysisData.historical_bars.reduce((sum: number, b: any) => sum + b.volume, 0) / analysisData.historical_bars.length / 1000000).toFixed(1)}M
                    </p>
                  </div>
                  <div className="bg-slate-700/30 rounded p-3">
                    <p className="text-slate-400 text-xs mb-1">Range</p>
                    <p className="font-bold text-white text-lg">
                      ${(Math.max(...analysisData.historical_bars.map((b: any) => b.high)) - Math.min(...analysisData.historical_bars.map((b: any) => b.low))).toFixed(2)}
                    </p>
                  </div>
                  <div className="bg-slate-700/30 rounded p-3">
                    <p className="text-slate-400 text-xs mb-1">% Change</p>
                    <p className={`font-bold text-lg ${
                      analysisData.historical_bars[analysisData.historical_bars.length - 1].close >= analysisData.historical_bars[0].open
                        ? 'text-green-400'
                        : 'text-red-400'
                    }`}>
                      {(((analysisData.historical_bars[analysisData.historical_bars.length - 1].close - analysisData.historical_bars[0].open) / analysisData.historical_bars[0].open) * 100).toFixed(2)}%
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Support & Resistance Section */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
              <h3 className="text-xl font-bold mb-4">Support & Resistance Levels</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <p className="text-slate-400 text-sm font-semibold mb-2">Support Level</p>
                  <div className="bg-green-900/20 border border-green-700 rounded p-4">
                    <p className="text-green-400 font-bold text-2xl">
                      ${analysisData.support_level?.toFixed(2) || 'N/A'}
                    </p>
                    {analysisData.distance_to_support !== undefined && (
                      <p className="text-xs text-slate-400 mt-2">
                        {analysisData.distance_to_support?.toFixed(2)}% below current price
                      </p>
                    )}
                  </div>
                </div>
                <div>
                  <p className="text-slate-400 text-sm font-semibold mb-2">Resistance Level</p>
                  <div className="bg-red-900/20 border border-red-700 rounded p-4">
                    <p className="text-red-400 font-bold text-2xl">
                      ${analysisData.resistance_level?.toFixed(2) || 'N/A'}
                    </p>
                    {analysisData.distance_to_resistance !== undefined && (
                      <p className="text-xs text-slate-400 mt-2">
                        {analysisData.distance_to_resistance?.toFixed(2)}% above current price
                      </p>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* Chart Pattern Analysis */}
            {analysisData.chart_pattern && (
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
                <h3 className="text-xl font-bold mb-4">Chart Pattern Analysis</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <div className="space-y-3">
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Pattern</p>
                        <p className="text-white font-bold text-lg">
                          {analysisData.chart_pattern.pattern}
                        </p>
                      </div>
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Direction</p>
                        <p className={`font-bold text-lg ${
                          analysisData.chart_pattern.direction === 'LONG'
                            ? 'text-green-400'
                            : analysisData.chart_pattern.direction === 'SHORT'
                              ? 'text-red-400'
                              : 'text-yellow-400'
                        }`}>
                          {analysisData.chart_pattern.direction}
                        </p>
                      </div>
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Confidence</p>
                        <div className="flex items-center gap-2">
                          <div className="flex-1 bg-slate-700 rounded-full h-2">
                            <div
                              className={`h-2 rounded-full ${
                                analysisData.chart_pattern.confidence >= 70
                                  ? 'bg-green-500'
                                  : analysisData.chart_pattern.confidence >= 50
                                    ? 'bg-yellow-500'
                                    : 'bg-red-500'
                              }`}
                              style={{ width: `${analysisData.chart_pattern.confidence}%` }}
                            />
                          </div>
                          <p className="text-white font-semibold w-12 text-right">
                            {analysisData.chart_pattern.confidence?.toFixed(1)}%
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div>
                    <div className="space-y-3">
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Target Price</p>
                        <p className="text-white font-bold text-lg">
                          ${(analysisData.chart_pattern.target_up || analysisData.chart_pattern.target_down)?.toFixed(2) || 'N/A'}
                        </p>
                      </div>
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Stop Loss</p>
                        <p className="text-white font-bold text-lg">
                          ${analysisData.chart_pattern.stop_loss?.toFixed(2) || 'N/A'}
                        </p>
                      </div>
                      <div>
                        <p className="text-slate-400 text-sm font-semibold">Risk:Reward Ratio</p>
                        <p className="text-white font-bold text-lg">
                          {analysisData.chart_pattern.risk_reward?.toFixed(2) || 'N/A'}:1
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
                {analysisData.chart_pattern.reasoning && (
                  <div className="mt-4 p-3 bg-slate-700/50 rounded">
                    <p className="text-slate-300 text-sm italic">
                      <span className="font-semibold">Reasoning: </span>
                      {analysisData.chart_pattern.reasoning}
                    </p>
                  </div>
                )}
              </div>
            )}

            {/* Trading Recommendation */}
            {analysisData.trading_recommendation && (
              <div className={`rounded-lg p-6 border ${
                analysisData.trading_recommendation.action === 'BUY'
                  ? 'bg-green-900/20 border-green-700'
                  : analysisData.trading_recommendation.action === 'SELL'
                    ? 'bg-red-900/20 border-red-700'
                    : 'bg-yellow-900/20 border-yellow-700'
              }`}>
                <div className="flex justify-between items-start mb-4">
                  <h3 className={`text-2xl font-bold ${
                    analysisData.trading_recommendation.action === 'BUY'
                      ? 'text-green-400'
                      : analysisData.trading_recommendation.action === 'SELL'
                        ? 'text-red-400'
                        : 'text-yellow-400'
                  }`}>
                    {analysisData.trading_recommendation.action}
                  </h3>
                  <div className="text-right">
                    <p className="text-slate-400 text-xs font-semibold">Confidence</p>
                    <p className={`text-2xl font-bold ${
                      analysisData.trading_recommendation.confidence >= 70
                        ? 'text-green-400'
                        : analysisData.trading_recommendation.confidence >= 50
                          ? 'text-yellow-400'
                          : 'text-red-400'
                    }`}>
                      {analysisData.trading_recommendation.confidence?.toFixed(1)}%
                    </p>
                  </div>
                </div>
                {analysisData.trading_recommendation.reasoning && (
                  <p className="text-slate-300 mb-4">
                    {analysisData.trading_recommendation.reasoning}
                  </p>
                )}
              </div>
            )}

            {/* Data Summary */}
            <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
              <p className="text-xs text-slate-400">
                Analysis based on {analysisData.bars_analyzed} bars | Updated: {new Date(analysisData.timestamp * 1000).toLocaleString()}
              </p>
            </div>
          </div>
        )}

        {/* Empty State */}
        {!analysisData && !isLoading && (
          <div className="text-center py-12">
            <Search className="w-16 h-16 text-slate-600 mx-auto mb-4" />
            <p className="text-slate-400 text-lg">Enter a stock symbol to begin analysis</p>
          </div>
        )}
      </div>
    </div>
  );
}
