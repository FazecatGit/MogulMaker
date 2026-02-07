'use client';

import React, { useState } from 'react';
import { RefreshCw, Search, TrendingUp, TrendingDown } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import Button from '@/components/ui/Button';
import StatCard from '@/components/ui/StatCard';
import ErrorAlert from '@/components/ui/ErrorAlert';
import PriceChart from '@/components/Charts/PriceChart';
import apiClient from '@/lib/apiClient';
import { formatCurrency, formatPercent, formatNumber } from '@/lib/formatters';
import { getPnLColor, getStatCardVariant, getTrendColor, getScoreColor } from '@/lib/colorHelpers';

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
            <Button
              variant="primary"
              icon={isLoading && <RefreshCw className="w-4 h-4 animate-spin" />}
              loading={isLoading}
              onClick={() => analyzeStock({ preventDefault: () => {} } as React.FormEvent)}
            >
              {isLoading ? 'Analyzing...' : 'Analyze'}
            </Button>
          </div>
        </form>

        {/* Error Message */}
        {error && (
          <div className="mb-6">
            <ErrorAlert message={error} />
          </div>
        )}

        {/* Analysis Results */}
        {analysisData && (
          <div className="space-y-6">
            {/* Header with Symbol and Clear Button */}
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-3xl font-bold">{symbol} Analysis</h2>
              <div className="flex gap-2">
                <Button
                  variant="success"
                  loading={buyingSymbol === symbol}
                  onClick={handleBuyStock}
                >
                  Buy
                </Button>
                <Button
                  variant="danger"
                  loading={sellingSymbol === symbol}
                  onClick={handleSellStock}
                >
                  Sell
                </Button>
                <Button
                  variant="secondary"
                  onClick={clearAnalysis}
                >
                  Clear
                </Button>
              </div>
            </div>

            {/* Core Metrics Grid */}
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              {/* Current Price */}
              <StatCard
                label="Price"
                value={formatCurrency(analysisData.current_price || 0)}
              />

              {/* RSI */}
              <StatCard
                label="RSI"
                value={analysisData.rsi?.toFixed(1) || 'N/A'}
                subtext={
                  analysisData.rsi ? (
                    analysisData.rsi > 70
                      ? '🔴 Overbought'
                      : analysisData.rsi < 30
                        ? '🟢 Oversold'
                        : '🟡 Neutral'
                  ) : (
                    'N/A'
                  )
                }
              />

              {/* ATR (Raw Value) */}
              <StatCard
                label="ATR"
                value={formatCurrency(analysisData.atr || 0)}
                subtext="14-period avg range"
              />

              {/* Volatility % */}
              <StatCard
                label="Volatility %"
                value={
                  analysisData.atr && analysisData.current_price
                    ? formatPercent((analysisData.atr / analysisData.current_price) * 0.01)
                    : 'N/A'
                }
                subtext={
                  analysisData.atr && analysisData.current_price ? (
                    (() => {
                      const atrPercent = (analysisData.atr / analysisData.current_price) * 100;
                      if (atrPercent > 3) return '🔴 High';
                      if (atrPercent > 1.5) return '🟡 Medium';
                      return '🟢 Low';
                    })()
                  ) : (
                    'N/A'
                  )
                }
              />

              {/* SMA 20 */}
              <StatCard
                label="SMA 20"
                value={formatCurrency(analysisData.sma_20 || 0)}
              />

              {/* Trend */}
              <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
                <p className="text-slate-400 text-xs mb-1 font-medium">TREND</p>
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
              <>
                <PriceChart
                  data={analysisData.historical_bars}
                  title={`Price History (${analysisData.historical_bars.length} days)`}
                  daysLabel={`${analysisData.historical_bars.length} trading days`}
                  height={400}
                />

                {/* Price Statistics */}
                <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
                  <StatCard
                    label="Highest"
                    value={formatCurrency(Math.max(...analysisData.historical_bars.map((b: any) => b.high)))}
                    variant="positive"
                  />
                  <StatCard
                    label="Lowest"
                    value={formatCurrency(Math.min(...analysisData.historical_bars.map((b: any) => b.low)))}
                    variant="negative"
                  />
                  <StatCard
                    label="Avg Volume"
                    value={formatNumber(analysisData.historical_bars.reduce((sum: number, b: any) => sum + b.volume, 0) / analysisData.historical_bars.length / 1000000, 1) + 'M'}
                  />
                  <StatCard
                    label="Range"
                    value={formatCurrency(Math.max(...analysisData.historical_bars.map((b: any) => b.high)) - Math.min(...analysisData.historical_bars.map((b: any) => b.low)))}
                  />
                  <StatCard
                    label="% Change"
                    value={formatPercent((analysisData.historical_bars[analysisData.historical_bars.length - 1].close - analysisData.historical_bars[0].open) / analysisData.historical_bars[0].open)}
                    variant={getStatCardVariant(analysisData.historical_bars[analysisData.historical_bars.length - 1].close - analysisData.historical_bars[0].open)}
                  />
                </div>
              </>
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
