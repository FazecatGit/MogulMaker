/**
 * Format a number as a percentage
 * @param value - The value to format (0.5 = 50%)
 * @param decimals - Number of decimal places (default: 2)
 */
export const formatPercent = (value: number, decimals = 2): string => {
  return `${(value * 100).toFixed(decimals)}%`;
};

/**
 * Format a number as currency (USD)
 * @param value - The value to format
 * @param decimals - Number of decimal places (default: 2)
 */
export const formatCurrency = (value: number, decimals = 2): string => {
  return `$${value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',')}`;
};

/**
 * Format a date string to readable format
 * @param dateString - Date string in format: YYYY-MM-DD or timestamp
 */
export const formatDate = (dateString: string): string => {
  if (!dateString) return 'N/A';
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  } catch {
    return dateString;
  }
};

/**
 * Format a duration in seconds to human-readable format
 * @param seconds - Duration in seconds
 */
export const formatDuration = (seconds: number): string => {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
  return `${(seconds / 3600).toFixed(1)}h`;
};

/**
 * Format a number with commas for thousands separator
 * @param value - The value to format
 * @param decimals - Number of decimal places (default: 0)
 */
export const formatNumber = (value: number, decimals = 0): string => {
  return value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

/**
 * Format time as HH:MM:SS
 * @param timestamp - Unix timestamp or date string
 */
export const formatTime = (timestamp: string | number): string => {
  try {
    const date = new Date(typeof timestamp === 'string' ? timestamp : timestamp * 1000);
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return 'N/A';
  }
};

/**
 * Parse date from multiple formats
 * @param dateString - Date string in formats: YYYY-MM-DD, DD/MM/YYYY, DD.MM.YYYY, MM-DD-YYYY
 */
export const parseDate = (dateString: string): Date | null => {
  if (!dateString) return null;

  // YYYY-MM-DD
  if (/^\d{4}-\d{2}-\d{2}$/.test(dateString)) {
    return new Date(dateString);
  }

  // DD/MM/YYYY or DD.MM.YYYY
  const dmyRegex = /^(\d{2})[/.]\d{2}[/.]\d{4}$/;
  if (dmyRegex.test(dateString)) {
    const [d, m, y] = dateString.split(/[\/.]/);
    return new Date(`${y}-${m}-${d}`);
  }

  // MM-DD-YYYY
  if (/^\d{2}-\d{2}-\d{4}$/.test(dateString)) {
    const [m, d, y] = dateString.split('-');
    return new Date(`${y}-${m}-${d}`);
  }

  return null;
};

/**
 * Get safe value with fallback
 * @param value - The value
 * @param fallback - Fallback value if undefined/null (default: 'N/A')
 */
export const safeValue = <T,>(value: T | null | undefined, fallback: string = 'N/A'): T | string => {
  return value ?? fallback;
};
