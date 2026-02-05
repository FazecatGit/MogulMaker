'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, AlertTriangle, TrendingDown, DollarSign, Activity, RefreshCw } from 'lucide-react';
import apiClient from '@/lib/apiClient';

interface RiskAlert {
  id: string;
  level: 'critical' | 'warning' | 'info';
  title: string;
  description: string;
  metric: string;
  currentValue: number;
  threshold: number;
  symbol?: string;
  timestamp: string;
}

interface RiskMetrics {
  dailyLoss: number;
  dailyLossLimit: number;
  portfolioRisk: number;
  maxDrawdown: number;
  maxDrawdownPercent: number;
  openPositions: number;
  positionLimit: number;
  averageRiskPerTrade: number;
  largestPosition: {
    symbol: string;
    risk: number;
  };
}

interface RiskData {
  metrics: RiskMetrics;
  alerts: RiskAlert[];
  lastUpdated: string;
}

export default function RiskPage() {
  const [riskData, setRiskData] = useState<RiskData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchRiskData();
  }, []);

  const fetchRiskData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      // Fetch both risk metrics and alerts
      const [riskResponse, alertsResponse] = await Promise.all([
        apiClient.get('/risk'),
        apiClient.get('/risk/alerts'),
      ]);

      const backendData: any = riskResponse || {};
      const alertsData: any = alertsResponse || {};
      
      if (backendData) {
        // Transform backend response to match RiskData interface
        const transformedData: RiskData = {
          metrics: {
            dailyLoss: Math.abs((backendData.daily_loss_percent || 0) * (backendData.portfolio_value || 0) / 100) || 0,
            dailyLossLimit: (backendData.account_balance || 0) * 0.05, // 5% of account
            portfolioRisk: parseFloat(((backendData.portfolio_risk_pct || 0) * 100).toFixed(2)) || 0,
            maxDrawdown: Math.abs(backendData.total_unrealized_pnl || 0) || 0,
            maxDrawdownPercent: parseFloat((((backendData.total_unrealized_pnl || 0) / (backendData.portfolio_value || 1)) * 100).toFixed(2)) || 0,
            openPositions: backendData.open_positions || 0,
            positionLimit: backendData.position_limit || 10,
            averageRiskPerTrade: backendData.positions ? 
              (backendData.positions.reduce((sum: number, p: any) => sum + Math.abs(p.unrealized_pl || 0), 0) / backendData.positions.length) || 0 : 0,
            largestPosition: {
              symbol: backendData.positions?.[0]?.symbol || 'N/A',
              risk: Math.max(...(backendData.positions?.map((p: any) => Math.abs(p.unrealized_pl || 0)) || [0])) || 0,
            },
          },
          alerts: alertsData.alerts || [],
          lastUpdated: new Date((backendData.timestamp || Date.now() / 1000) * 1000).toISOString(),
        };
        setRiskData(transformedData);
      }
    } catch (err: any) {
      console.error('Risk data fetch error:', err);
      setError(err.message || 'Failed to fetch risk data');
    } finally {
      setIsLoading(false);
    }
  };

  const getAlertIcon = (level: string) => {
    switch (level) {
      case 'critical':
        return <AlertTriangle className="w-5 h-5" />;
      case 'warning':
        return <AlertCircle className="w-5 h-5" />;
      default:
        return <Activity className="w-5 h-5" />;
    }
  };

  const getAlertColors = (level: string) => {
    switch (level) {
      case 'critical':
        return 'bg-red-500/20 border-red-500/50 text-red-400';
      case 'warning':
        return 'bg-yellow-500/20 border-yellow-500/50 text-yellow-400';
      default:
        return 'bg-blue-500/20 border-blue-500/50 text-blue-400';
    }
  };

  const getRiskPercentage = (current: number, limit: number) => {
    return Math.round((current / limit) * 100);
  };

  const getRiskBarColor = (percentage: number) => {
    if (percentage >= 80) return 'bg-red-600';
    if (percentage >= 60) return 'bg-yellow-600';
    return 'bg-green-600';
  };

  if (!riskData && isLoading) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Risk Dashboard</h1>
          <p className="text-slate-400">Portfolio risk management and alerts</p>
        </div>

        {/* Loading Skeleton */}
        <div className="space-y-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="bg-slate-800 rounded-lg border border-slate-700 p-4 animate-pulse">
              <div className="h-5 bg-slate-700 rounded w-1/3 mb-3"></div>
              <div className="h-4 bg-slate-700 rounded w-full"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error && !riskData) {
    return (
      <div className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Risk Dashboard</h1>
          <p className="text-slate-400">Portfolio risk management and alerts</p>
        </div>

        <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-4 text-red-400">
          <p>{error}</p>
        </div>
      </div>
    );
  }

  const metrics = riskData?.metrics;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">Risk Dashboard</h1>
          <p className="text-slate-400">Portfolio risk management and alerts</p>
        </div>
        <button
          onClick={fetchRiskData}
          disabled={isLoading}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Key Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Daily Loss */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-slate-400 text-sm font-semibold">Daily Loss</span>
            <DollarSign className="w-4 h-4 text-slate-500" />
          </div>
          <div className="text-2xl font-bold text-white mb-3">
            ${metrics?.dailyLoss.toLocaleString() || 0}
          </div>
          <div className="space-y-2">
            <div className="flex justify-between text-xs text-slate-400">
              <span>Usage</span>
              <span>{getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1)}%</span>
            </div>
            <div className="w-full bg-slate-700 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all ${getRiskBarColor(
                  getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1)
                )}`}
                style={{
                  width: `${Math.min(
                    100,
                    getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1)
                  )}%`,
                }}
              ></div>
            </div>
            <div className="text-xs text-slate-500">
              Limit: ${metrics?.dailyLossLimit.toLocaleString() || 0}
            </div>
          </div>
        </div>

        {/* Portfolio Risk */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-slate-400 text-sm font-semibold">Portfolio Risk</span>
            <TrendingDown className="w-4 h-4 text-slate-500" />
          </div>
          <div className="text-2xl font-bold text-white mb-3">{metrics?.portfolioRisk}%</div>
          <div className="text-xs text-slate-400">
            <p>Risk per position allocation</p>
            <p className="mt-2 text-yellow-400">⚠️ Approaching threshold (3%)</p>
          </div>
        </div>

        {/* Max Drawdown */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-slate-400 text-sm font-semibold">Max Drawdown</span>
            <TrendingDown className="w-4 h-4 text-slate-500" />
          </div>
          <div className="text-2xl font-bold text-white mb-3">
            {metrics?.maxDrawdownPercent}%
          </div>
          <div className="text-xs text-slate-400">
            ${metrics?.maxDrawdown.toLocaleString() || 0} from peak
          </div>
        </div>

        {/* Open Positions */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-slate-400 text-sm font-semibold">Open Positions</span>
            <Activity className="w-4 h-4 text-slate-500" />
          </div>
          <div className="text-2xl font-bold text-white mb-3">
            {metrics?.openPositions}/{metrics?.positionLimit}
          </div>
          <div className="space-y-2">
            <div className="flex justify-between text-xs text-slate-400">
              <span>Usage</span>
              <span>
                {getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1)}%
              </span>
            </div>
            <div className="w-full bg-slate-700 rounded-full h-2">
              <div
                className="h-2 rounded-full bg-blue-600 transition-all"
                style={{
                  width: `${Math.min(
                    100,
                    getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1)
                  )}%`,
                }}
              ></div>
            </div>
          </div>
        </div>
      </div>

      {/* Alerts Section */}
      <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
        <h2 className="text-xl font-bold text-white mb-4">Active Alerts</h2>

        {riskData?.alerts && riskData.alerts.length > 0 ? (
          <div className="space-y-3">
            {riskData.alerts.map((alert) => (
              <div
                key={alert.id}
                className={`border rounded-lg p-4 flex items-start gap-4 ${getAlertColors(alert.level)}`}
              >
                <div className="flex-shrink-0 mt-0.5">{getAlertIcon(alert.level)}</div>
                <div className="flex-1">
                  <h3 className="font-semibold mb-1">{alert.title}</h3>
                  <p className="text-sm opacity-90">{alert.description}</p>
                  {alert.symbol && (
                    <p className="text-xs mt-2 opacity-75">Symbol: {alert.symbol}</p>
                  )}
                </div>
                <div className="flex-shrink-0 text-right text-xs">
                  <p className="font-semibold">{alert.currentValue}</p>
                  <p className="opacity-75">of {alert.threshold}</p>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-slate-400">No active alerts - all systems normal</p>
          </div>
        )}
      </div>

      {/* Risk Details */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Largest Position */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">Largest Position at Risk</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-white">{metrics?.largestPosition.symbol}</p>
              <p className="text-sm text-slate-400">Risk: ${metrics?.largestPosition.risk}</p>
            </div>
            <TrendingDown className="w-8 h-8 text-red-400 opacity-50" />
          </div>
        </div>

        {/* Average Risk Per Trade */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">Avg Risk Per Trade</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-white">
                ${metrics?.averageRiskPerTrade.toLocaleString()}
              </p>
              <p className="text-sm text-slate-400">Per position</p>
            </div>
            <Activity className="w-8 h-8 text-blue-400 opacity-50" />
          </div>
        </div>
      </div>
    </div>
  );
}
