import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

export interface Trade {
  id: string;
  symbol: string;
  exchange: string;
  entry_time: string;
  exit_time: string | null;
  entry_price: string;
  exit_price: string | null;
  qty: string;
  side: 'buy' | 'sell';
  status: 'open' | 'closed';
  realized_pl: string;
  realized_plpc: string;
  duration_ms?: number;
}

export interface TradesResponse {
  count: number;
  trades: Trade[];
  risk_status: {
    enabled: boolean;
  };
  timestamp: number;
}

export function useTrades() {
  return useQuery<TradesResponse>({
    queryKey: ['trades-history'],
    queryFn: async () => {
      const response = await apiClient.get('/trades');
      const data = response?.data || response;
      return {
        count: data?.count || 0,
        trades: data?.trades || [],
        risk_status: data?.risk_status || { enabled: true },
        timestamp: data?.timestamp || Date.now(),
      };
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchInterval: 1000 * 30, // 30 seconds
  });
}
