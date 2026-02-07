/**
 * Get color class for score/confidence values
 * @param score - Score value (0-10)
 */
export const getScoreColor = (score: number | null | undefined): string => {
  if (score === null || score === undefined) return 'text-slate-400';
  if (score >= 8) return 'text-green-400';
  if (score >= 6) return 'text-yellow-400';
  return 'text-red-400';
};

/**
 * Get background color for score badges
 * @param score - Score value (0-10)
 */
export const getScoreBgColor = (score: number | null | undefined): string => {
  if (score === null || score === undefined) return 'bg-slate-700';
  if (score >= 8) return 'bg-green-900/30 border-green-700/50';
  if (score >= 6) return 'bg-yellow-900/30 border-yellow-700/50';
  return 'bg-red-900/30 border-red-700/50';
};

/**
 * Get color for P&L (profit/loss) values
 * @param value - The P&L value
 */
export const getPnLColor = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return 'text-slate-300';
  return value >= 0 ? 'text-green-400' : 'text-red-400';
};

/**
 * Get background color for P&L values
 * @param value - The P&L value
 */
export const getPnLBgColor = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return 'bg-slate-700/30';
  return value >= 0 ? 'bg-green-900/20 border-green-700/30' : 'bg-red-900/20 border-red-700/30';
};

/**
 * Get trend color for up/down movements
 * @param direction - 'up' or 'down'
 */
export const getTrendColor = (direction: 'up' | 'down'): string => {
  return direction === 'up' ? 'text-green-400' : 'text-red-400';
};

/**
 * Get status color
 * @param status - Status string
 */
export const getStatusColor = (status: string): string => {
  const statusLower = status?.toLowerCase() || '';
  
  if (statusLower.includes('success') || statusLower.includes('filled') || statusLower.includes('active')) {
    return 'text-green-400';
  }
  if (statusLower.includes('pending') || statusLower.includes('partial')) {
    return 'text-yellow-400';
  }
  if (statusLower.includes('error') || statusLower.includes('failed') || statusLower.includes('canceled')) {
    return 'text-red-400';
  }
  
  return 'text-slate-400';
};

/**
 * Get confidence/RSI color
 * @param value - Confidence/RSI value (0-100)
 */
export const getConfidenceColor = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return 'text-slate-400';
  if (value >= 70) return 'text-green-400';
  if (value >= 50) return 'text-yellow-400';
  return 'text-red-400';
};

/**
 * Get variant for StatCard based on numeric value
 */
export const getStatCardVariant = (value: number | null | undefined): 'positive' | 'negative' | 'neutral' | 'default' => {
  if (value === null || value === undefined) return 'default';
  if (value > 0) return 'positive';
  if (value < 0) return 'negative';
  return 'neutral';
};
