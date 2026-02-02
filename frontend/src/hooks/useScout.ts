import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/apiClient';

export interface Opportunity {
  symbol: string;
  score: number;
  analysis: string;
  rsi: number;
  atr: number;
  timestamp: number;
  rank: number;
}

export interface ScoutResponse {
  scanned_count: number;
  total_symbols: number;
  min_score: number;
  limit: number;
  opportunities: Opportunity[];
  scan_timestamp: number;
  message: string;
}

export function useScout(minScore: number = 50, limit: number = 15, offset: number = 0, enabled: boolean = false) {
  return useQuery<ScoutResponse>({
    queryKey: ['scout-opportunities', minScore, limit, offset, enabled],
    queryFn: async () => {
      try {
        console.log('[useScout] Starting query with:', { minScore, limit, offset, enabled });
        
        const response = await apiClient.get(`/scout?limit=${limit}&min_score=${minScore}&offset=${offset}`);
        
        console.log('[useScout] Raw response:', response);
        
        // apiClient already returns response.data from the interceptor
        const data = response as any;
        
        // Ensure opportunities is always an array
        const opportunities = Array.isArray(data?.opportunities) ? data.opportunities : [];
        
        console.log('[useScout] Parsed opportunities:', opportunities.length, 'items');
        
        const result = {
          scanned_count: Number(data?.scanned_count || 0),
          total_symbols: Number(data?.total_symbols || 0),
          min_score: Number(data?.min_score || minScore),
          limit: Number(data?.limit || limit),
          opportunities,
          scan_timestamp: Number(data?.scan_timestamp || Date.now()),
          message: String(data?.message || 'Scanning stocks...'),
        };
        
        console.log('[useScout] Final result:', result);
        return result;
      } catch (error: any) {
        console.error('[useScout] Error caught:', {
          message: error?.message,
          response: error?.response?.data,
          status: error?.response?.status,
          fullError: error,
        });
        throw error;
      }
    },
    enabled: enabled,
    staleTime: Infinity,
    refetchInterval: false,
    retry: false,
  });
}
