import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

export interface Trade {
  id: string;
  symbol: string;
  exchange: string;
  entry_time: string;
  exit_time: string | null;
  entry_price: number;
  exit_price: number | null;
  qty: number;
  side: 'buy' | 'sell';
  status: 'filled' | 'open' | 'closed' | 'pending' | string;
  realized_pl: number;
  realized_plpc: number;
  duration_ms?: number;
  order_type?: string;
  time_in_force?: string;
  submitted_at?: string;
  filled_at?: string;
  expires_at?: string;
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
        trades: (data?.trades || []).map((trade: any) => ({
          id: trade.id,
          symbol: trade.symbol,
          exchange: trade.exchange || 'NASDAQ',
          entry_time: trade.submitted_at || trade.entry_time || new Date().toISOString(),
          exit_time: trade.filled_at || trade.exit_time || null,
          entry_price: parseFloat(trade.entry_price) || 0,
          exit_price: trade.exit_price ? parseFloat(trade.exit_price) : null,
          qty: parseFloat(trade.qty) || 0,
          side: (trade.side?.toLowerCase() as 'buy' | 'sell') || 'buy',
          status: trade.status || 'unknown',
          realized_pl: parseFloat(trade.realized_pl) || 0,
          realized_plpc: parseFloat(trade.realized_plpc) || 0,
          duration_ms: trade.duration_ms,
          order_type: trade.order_type,
          time_in_force: trade.time_in_force,
          submitted_at: trade.submitted_at,
          filled_at: trade.filled_at,
          expires_at: trade.expires_at,
        })),
        risk_status: data?.risk_status || { enabled: true },
        timestamp: data?.timestamp || Date.now(),
      };
    },
    staleTime: 1000 * 10, // 10 seconds - trades data can change frequently
    refetchInterval: 1000 * 10, // 10 seconds - faster updates for recent trades
  });
}
