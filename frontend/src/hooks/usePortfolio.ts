/**
 * usePortfolio Hook
 * 
 * CONCEPTS EXPLAINED:
 * - useQuery: TanStack Query hook for fetching data
 * - Automatic caching, retry, loading states
 * - Returns: { data, isLoading, error, isError }
 * 
 * FLOW:
 * 1. Component renders and calls usePortfolio()
 * 2. Hook makes API request to /api/positions (via api-gateway)
 * 3. Data is cached for 5 minutes (from Providers.tsx config)
 * 4. Component re-renders with data
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

interface PortfolioData {
  totalPnL: number;
  portfolioValue: number;
  winRate: number;
  totalTrades: number;
  openPositions: number;
  dailyPnLPercent: number;
}

/**
 * Hook to fetch portfolio metrics
 * 
 * Usage in component:
 * const { data, isLoading, error } = usePortfolio();
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * return <div>{data.portfolioValue}</div>;
 */
export function usePortfolio() {
  return useQuery<PortfolioData>({
    queryKey: ['portfolio'], // Unique identifier for this data
    queryFn: async () => {
      // Fetch portfolio summary, risk adjustments (for account balance), and trades
      const summaryResponse = await apiClient.get('/portfolio-summary');
      const riskResponse = await apiClient.get('/risk/adjustments');
      const tradesResponse = await apiClient.get('/trades/statistics');

      // Extract account balance (total equity) from risk adjustments
      const risk: any = riskResponse || {};
      const accountBalance = parseFloat(risk?.account_balance || 0);

      // Extract position data from portfolio summary
      const summary: any = summaryResponse || {};
      const totalPnL = parseFloat(summary?.total_gain || 0);
      const openPositions = summary?.total_positions || 0;

      const stats: any = tradesResponse || {};

      console.log('Account Balance:', accountBalance);
      console.log('Portfolio Summary:', summary);
      console.log('Trade Statistics:', stats);

      const winRate = stats.win_rate || 0;
      const totalTrades = stats.total_trades || 0;

      console.log('Calculated portfolioValue:', accountBalance, 'totalPnL:', totalPnL, 'winRate:', winRate);

      return {
        totalPnL,
        portfolioValue: accountBalance, // Use total account equity, not just positions
        winRate: winRate * 100, // Convert to percentage
        totalTrades,
        openPositions,
        dailyPnLPercent: accountBalance > 0 ? (totalPnL / accountBalance) * 100 : 0,
      };
    },
    staleTime: 1000 * 60 * 5, // Revalidate every 5 minutes
    refetchInterval: 1000 * 30, // Auto-refresh every 30 seconds
  });
}

/**
 * Hook to fetch positions list
 */
export function usePositions() {
  return useQuery({
    queryKey: ['positions'],
    queryFn: async () => {
      return await apiClient.get('/positions');
    },
    staleTime: 1000 * 60 * 5,
    refetchInterval: 1000 * 30,
  });
}

/**
 * Hook to fetch trades history
 */
export function useTrades(limit = 50) {
  return useQuery({
    queryKey: ['trades', limit],
    queryFn: async () => {
      return await apiClient.get(`/trades?limit=${limit}`);
    },
    staleTime: 1000 * 60 * 5,
  });
}
