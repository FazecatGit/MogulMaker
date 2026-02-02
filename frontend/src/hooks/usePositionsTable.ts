/**
 * usePositionsTable Hook
 * 
 * Fetches detailed positions data with sorting/filtering capabilities
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

export interface Position {
  asset_id: string;
  symbol: string;
  exchange: string;
  asset_class: string;
  asset_marginable: boolean;
  qty: string;
  qty_available: string;
  avg_entry_price: string;
  side: 'long' | 'short';
  market_value: string;
  cost_basis: string;
  unrealized_pl: string;
  unrealized_plpc: string;
  unrealized_intraday_pl: string;
  unrealized_intraday_plpc: string;
  current_price: string;
  lastday_price: string;
  change_today: string;
}

export interface PositionsResponse {
  count: number;
  positions: Position[];
  risk_status: {
    enabled: boolean;
  };
  timestamp: number;
}

/**
 * Hook to fetch positions list with full details
 */
export function usePositionsTable() {
  return useQuery<PositionsResponse>({
    queryKey: ['positions-detailed'],
    queryFn: async () => {
      const response = await apiClient.get('/positions');
      const data = response?.data || response;
      
      console.log('Positions Response:', data);
      
      return {
        count: data?.count || 0,
        positions: data?.positions || [],
        risk_status: data?.risk_status || { enabled: true },
        timestamp: data?.timestamp || Date.now(),
      };
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchInterval: 1000 * 30, // 30 seconds
  });
}
