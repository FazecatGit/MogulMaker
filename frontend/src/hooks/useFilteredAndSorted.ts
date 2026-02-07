import { useMemo } from 'react';

export type SortDirection = 'asc' | 'desc';

interface FilterConfig {
  searchField?: string;
  statusField?: string;
  statusValue?: string | number;
  customFilter?: (item: any) => boolean;
}

interface SortConfig {
  field: string;
  direction?: SortDirection;
  customSort?: (a: any, b: any) => number;
}

/**
 * Custom hook for filtering and sorting data with search, status, and custom filters
 * @param data - Array of data to filter and sort
 * @param searchTerm - Search term to filter by (if searchField is specified)
 * @param sortConfig - Sort configuration
 * @param filterConfig - Filter configuration
 * @returns Filtered and sorted data
 */
export function useFilteredAndSorted<T extends Record<string, any>>(
  data: T[],
  searchTerm: string = '',
  sortConfig: SortConfig,
  filterConfig: FilterConfig = {}
): T[] {
  return useMemo(() => {
    let result = [...(data || [])];

    // Apply search filter
    if (searchTerm && filterConfig.searchField) {
      const searchLower = searchTerm.toLowerCase();
      result = result.filter((item) => {
        const value = String(item[filterConfig.searchField as keyof T]).toLowerCase();
        return value.includes(searchLower);
      });
    }

    // Apply status filter
    if (filterConfig.statusValue !== undefined && filterConfig.statusField) {
      result = result.filter((item) => item[filterConfig.statusField as keyof T] === filterConfig.statusValue);
    }

    // Apply custom filter
    if (filterConfig.customFilter) {
      result = result.filter(filterConfig.customFilter);
    }

    // Apply sorting
    if (sortConfig.customSort) {
      result.sort(sortConfig.customSort);
    } else if (sortConfig.field) {
      result.sort((a, b) => {
        const aValue = a[sortConfig.field as keyof T] as any;
        const bValue = b[sortConfig.field as keyof T] as any;

        // Handle numeric comparison
        if (typeof aValue === 'number' && typeof bValue === 'number') {
          return sortConfig.direction === 'asc'
            ? aValue - bValue
            : bValue - aValue;
        }

        // Handle string comparison
        if (typeof aValue === 'string' && typeof bValue === 'string') {
          return sortConfig.direction === 'asc'
            ? aValue.localeCompare(bValue)
            : bValue.localeCompare(aValue);
        }

        // Handle date comparison
        if (aValue instanceof Date && bValue instanceof Date) {
          return sortConfig.direction === 'asc'
            ? aValue.getTime() - bValue.getTime()
            : bValue.getTime() - aValue.getTime();
        }

        return 0;
      });
    }

    return result;
  }, [data, searchTerm, sortConfig, filterConfig]);
}
