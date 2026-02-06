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

export interface PendingOrder {
  id: string;
  symbol: string;
  side: 'buy' | 'sell';
  qty: string;
  filled_qty: string;
  type: string;
  status: string;
  submitted_at: string;
  filled_avg_price: string | null;
}

export interface PositionsResponse {
  count: number;
  positions: Position[];
  pending_orders: PendingOrder[];
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
      try {
        const response = await apiClient.get('/positions');
        const data = response?.data || response;
        
        console.log('=== Positions Hook ===');
        console.log('Raw Response:', response);
        console.log('Parsed Data:', data);
        console.log('Positions Array:', data?.positions);
        console.log('Positions Count:', data?.positions?.length || 0);
        
        const result = {
          count: data?.count || data?.positions?.length || 0,
          positions: Array.isArray(data?.positions) ? data.positions : [],
          pending_orders: Array.isArray(data?.pending_orders) ? data.pending_orders : [],
          risk_status: data?.risk_status || { enabled: true },
          timestamp: data?.timestamp || Date.now(),
        };
        
        console.log('Final Result:', result);
        
        return result;
      } catch (error) {
        console.error('Failed to fetch positions:', error);
        throw error;
      }
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchInterval: 1000 * 30, // 30 seconds
  });
}
