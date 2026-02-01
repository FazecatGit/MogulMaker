'use client';

import { usePortfolio } from '@/hooks/usePortfolio';
import { Loader2, AlertCircle } from 'lucide-react';

/**
 * Dashboard Page - Real Data Version
 * 
 * FLOW:
 * 1. Component calls usePortfolio() hook
 * 2. Hook fetches data from API gateway
 * 3. While loading, show skeleton
 * 4. If error, show error message
 * 5. If success, display real data
 */

export default function DashboardPage() {
  const { data, isLoading, error, isError } = usePortfolio();

  // Loading state - show spinners
  if (isLoading) {
    return (
      <div>
        <h1 className="text-3xl font-bold mb-8">Portfolio Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="bg-slate-800 rounded-lg p-6 border border-slate-700 animate-pulse"
            >
              <div className="h-4 bg-slate-700 rounded mb-4 w-20"></div>
              <div className="h-8 bg-slate-700 rounded mb-2 w-32"></div>
              <div className="h-4 bg-slate-700 rounded w-24"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (isError) {
    return (
      <div>
        <h1 className="text-3xl font-bold mb-8">Portfolio Dashboard</h1>
        <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-400" />
          <div>
            <p className="text-red-400 font-semibold">Failed to load portfolio</p>
            <p className="text-red-300 text-sm">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Success state - display real data
  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">Portfolio Dashboard</h1>

      {/* Real Data Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Total P&L Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Total P&L</p>
          <p
            className={`text-3xl font-bold ${
              data && data.totalPnL >= 0 ? 'text-green-400' : 'text-red-400'
            }`}
          >
            {data ? (
              <>
                {data.totalPnL >= 0 ? '+' : ''}
                ${data.totalPnL.toFixed(2)}
              </>
            ) : (
              'Loading...'
            )}
          </p>
          <p
            className={`text-sm mt-2 ${
              data && data.dailyPnLPercent >= 0
                ? 'text-green-400'
                : 'text-red-400'
            }`}
          >
            {data ? `${data.dailyPnLPercent >= 0 ? '+' : ''}${data.dailyPnLPercent.toFixed(1)}%` : 'Loading...'}
          </p>
        </div>

        {/* Portfolio Value Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Portfolio Value</p>
          <p className="text-3xl font-bold text-white">
            {data ? `$${data.portfolioValue.toFixed(2)}` : 'Loading...'}
          </p>
          <p className="text-sm text-slate-400 mt-2">
            {data ? `${data.openPositions} open positions` : 'Loading...'}
          </p>
        </div>

        {/* Win Rate Card */}
        <div className="bg-slate-800 rounded-lg p-6 border border-slate-700">
          <p className="text-slate-400 text-sm mb-2">Win Rate</p>
          <p className="text-3xl font-bold text-blue-400">
            {data ? `${data.winRate.toFixed(0)}%` : 'Loading...'}
          </p>
          <p className="text-sm text-slate-400 mt-2">
            {data ? `${data.totalTrades} total trades` : 'Loading...'}
          </p>
        </div>
      </div>

      {/* Coming Soon */}
      <div className="mt-12 bg-slate-800 rounded-lg p-8 border border-slate-700 text-center">
        <p className="text-slate-400">More features coming soon...</p>
      </div>
    </div>
  );
}
