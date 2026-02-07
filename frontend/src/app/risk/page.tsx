'use client';

import { useState, useEffect } from 'react';
import { AlertCircle, AlertTriangle, TrendingDown, DollarSign, Activity, RefreshCw } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import StatCard from '@/components/ui/StatCard';
import SkeletonLoader from '@/components/ui/SkeletonLoader';
import StatusAlert from '@/components/ui/StatusAlert';
import { formatCurrency } from '@/lib/formatters';
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
    // Auto-fetch risk data when page loads
    fetchRiskData();
    
    // Auto-refresh every 10 seconds for real-time alerts
    const interval = setInterval(() => {
      fetchRiskData();
    }, 10000);
    
    return () => clearInterval(interval);
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
        <PageHeader title="Risk Dashboard" description="Portfolio risk management and alerts" />
        <SkeletonLoader count={4} withContent />
      </div>
    );
  }

  if (error && !riskData) {
    return (
      <div className="space-y-6">
        <PageHeader title="Risk Dashboard" description="Portfolio risk management and alerts" />
        <StatusAlert message={error} variant="error" />
      </div>
    );
  }

  const metrics = riskData?.metrics;

  return (
    <div className="w-full space-y-8">
      {/* Controls */}
      <div className="flex justify-end">
        <button
          onClick={fetchRiskData}
          disabled={isLoading}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white px-4 py-2 rounded-lg font-semibold transition"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>
      
      {/* Header */}
      <PageHeader title="Risk Dashboard" description="Portfolio risk management and alerts" />

      {/* Key Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Daily Loss Tracking with Progress Bar */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-3">
            <DollarSign className="w-4 h-4 text-red-400" />
            <p className="text-slate-400 text-sm font-semibold">Daily Loss</p>
          </div>
          <p className={`text-2xl font-bold ${(metrics?.dailyLoss || 0) > 0 ? 'text-red-400' : 'text-green-400'}`}>
            {formatCurrency(metrics?.dailyLoss || 0)}
          </p>
          <p className="text-xs text-slate-500 mt-1">Limit: {formatCurrency(metrics?.dailyLossLimit || 0)}</p>
          
          {/* Progress Bar */}
          <div className="mt-4">
            <div className="flex justify-between text-xs text-slate-400 mb-1">
              <span>Usage</span>
              <span>{getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1)}%</span>
            </div>
            <div className="w-full bg-slate-700 rounded-full h-2.5 overflow-hidden">
              <div
                className={`h-full transition-all duration-300 ${getRiskBarColor(getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1))}`}
                style={{ width: `${Math.min(getRiskPercentage(metrics?.dailyLoss || 0, metrics?.dailyLossLimit || 1), 100)}%` }}
              />
            </div>
          </div>
        </div>

        {/* Portfolio Risk */}
        <StatCard
          label="Portfolio Risk"
          value={`${metrics?.portfolioRisk || 0}%`}
          subtext="Risk per position allocation"
          variant={metrics?.portfolioRisk && metrics.portfolioRisk >= 3 ? "warning" : "default"}
          icon={<TrendingDown className="w-4 h-4" />}
        />

        {/* Max Drawdown */}
        <StatCard
          label="Max Drawdown"
          value={`${metrics?.maxDrawdownPercent || 0}%`}
          subtext={`${formatCurrency(metrics?.maxDrawdown || 0)} from peak`}
          variant={metrics?.maxDrawdownPercent && metrics.maxDrawdownPercent <= -10 ? "negative" : "default"}
          icon={<TrendingDown className="w-4 h-4" />}
        />

        {/* Position Limit Status with Progress Bar */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <div className="flex items-center gap-2 mb-3">
            <Activity className="w-4 h-4 text-blue-400" />
            <p className="text-slate-400 text-sm font-semibold">Open Positions</p>
          </div>
          <p className="text-2xl font-bold text-white">
            {metrics?.openPositions || 0}<span className="text-slate-400 text-lg">/{metrics?.positionLimit || 0}</span>
          </p>
          <p className="text-xs text-slate-500 mt-1">
            {getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1)}% capacity used
          </p>
          
          {/* Progress Bar */}
          <div className="mt-4">
            <div className="flex justify-between text-xs text-slate-400 mb-1">
              <span>Capacity</span>
              <span>{getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1)}%</span>
            </div>
            <div className="w-full bg-slate-700 rounded-full h-2.5 overflow-hidden">
              <div
                className={`h-full transition-all duration-300 ${getRiskBarColor(getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1))}`}
                style={{ width: `${Math.min(getRiskPercentage(metrics?.openPositions || 0, metrics?.positionLimit || 1), 100)}%` }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Alerts Section */}
      <div className="content-card page-section">
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
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {/* Largest Position */}
        <div className="content-card">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">Largest Position at Risk</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-white">{metrics?.largestPosition.symbol}</p>
              <p className="text-sm text-slate-400">Risk: {formatCurrency(metrics?.largestPosition.risk || 0)}</p>
            </div>
            <TrendingDown className="w-8 h-8 text-red-400 opacity-50" />
          </div>
        </div>

        {/* Average Risk Per Trade */}
        <div className="content-card">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">Avg Risk Per Trade</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-2xl font-bold text-white">
                {formatCurrency(metrics?.averageRiskPerTrade || 0)}
              </p>
              <p className="text-sm text-slate-400">Per position</p>
            </div>
            <Activity className="w-8 h-8 text-blue-400 opacity-50" />
          </div>
        </div>

        {/* Risk Level Indicator */}
        <div className="content-card">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">Overall Risk Level</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className={`text-2xl font-bold ${
                (metrics?.portfolioRisk || 0) >= 5 ? 'text-red-400' :
                (metrics?.portfolioRisk || 0) >= 3 ? 'text-yellow-400' :
                'text-green-400'
              }`}>
                {(metrics?.portfolioRisk || 0) >= 5 ? 'HIGH' :
                 (metrics?.portfolioRisk || 0) >= 3 ? 'MODERATE' :
                 'LOW'}
              </p>
              <p className="text-sm text-slate-400">Based on portfolio risk %</p>
            </div>
            <AlertTriangle className={`w-8 h-8 opacity-50 ${
              (metrics?.portfolioRisk || 0) >= 5 ? 'text-red-400' :
              (metrics?.portfolioRisk || 0) >= 3 ? 'text-yellow-400' :
              'text-green-400'
            }`} />
          </div>
        </div>
      </div>
    </div>
  );
}
