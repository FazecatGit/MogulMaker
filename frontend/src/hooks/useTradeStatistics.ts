import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

export interface TradeStatistics {
  total_trades: number;
  winning_trades: number;
  losing_trades: number;
  win_rate: number;
  total_pnl: number;
  avg_pnl: number;
  largest_win: number;
  largest_loss: number;
  avg_trade_duration: string;
  sharpe_ratio: number;
  sortino_ratio: number;
  open_positions: number;
  open_pnl: number;
  timestamp: number;
}

export function useTradeStatistics() {
  return useQuery<TradeStatistics>({
    queryKey: ['trade-statistics'],
    queryFn: async () => {
      const response = await apiClient.get('/trades/statistics');
      const data = response?.data || response;
      return {
        total_trades: data?.total_trades || 0,
        winning_trades: data?.winning_trades || 0,
        losing_trades: data?.losing_trades || 0,
        win_rate: data?.win_rate || 0,
        total_pnl: data?.total_pnl || 0,
        avg_pnl: data?.avg_pnl || 0,
        largest_win: data?.largest_win || 0,
        largest_loss: data?.largest_loss || 0,
        avg_trade_duration: data?.avg_trade_duration || '0s',
        sharpe_ratio: data?.sharpe_ratio || 0,
        sortino_ratio: data?.sortino_ratio || 0,
        open_positions: data?.open_positions || 0,
        open_pnl: data?.open_pnl || 0,
        timestamp: data?.timestamp || Date.now(),
      };
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchInterval: 1000 * 30, // 30 seconds
  });
}
