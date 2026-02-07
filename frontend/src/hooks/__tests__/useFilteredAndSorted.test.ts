import { renderHook } from '@testing-library/react';
import { useFilteredAndSorted } from '../useFilteredAndSorted';

describe('useFilteredAndSorted', () => {
  const mockData = [
    { id: 1, name: 'Alice', status: 'active', score: 95 },
    { id: 2, name: 'Bob', status: 'inactive', score: 85 },
    { id: 3, name: 'Charlie', status: 'active', score: 90 },
    { id: 4, name: 'David', status: 'active', score: 88 },
  ];

  it('returns data as-is when no filters or sort', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: '' })
    );
    expect(result.current).toEqual(mockData);
  });

  it('filters by search term', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(
        mockData,
        'ali',
        { field: '' },
        { searchField: 'name' }
      )
    );
    expect(result.current).toHaveLength(1);
    expect(result.current[0].name).toBe('Alice');
  });

  it('filters by status', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(
        mockData,
        '',
        { field: '' },
        { statusField: 'status', statusValue: 'active' }
      )
    );
    expect(result.current).toHaveLength(3);
    expect(result.current.every((item) => item.status === 'active')).toBe(true);
  });

  it('sorts numerically ascending', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: 'score', direction: 'asc' })
    );
    expect(result.current[0].score).toBe(85);
    expect(result.current[3].score).toBe(95);
  });

  it('sorts numerically descending', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: 'score', direction: 'desc' })
    );
    expect(result.current[0].score).toBe(95);
    expect(result.current[3].score).toBe(85);
  });

  it('sorts alphabetically ascending', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: 'name', direction: 'asc' })
    );
    expect(result.current[0].name).toBe('Alice');
    expect(result.current[3].name).toBe('David');
  });

  it('sorts alphabetically descending', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: 'name', direction: 'desc' })
    );
    expect(result.current[0].name).toBe('David');
    expect(result.current[3].name).toBe('Alice');
  });

  it('applies custom filter', () => {
    const customFilter = (item: any) => item.score > 85;
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: '' }, { customFilter })
    );
    expect(result.current).toHaveLength(3);
    expect(result.current.every((item) => item.score > 85)).toBe(true);
  });

  it('applies custom sort', () => {
    const customSort = (a: any, b: any) => b.id - a.id; // Reverse ID order
    const { result } = renderHook(() =>
      useFilteredAndSorted(mockData, '', { field: '', customSort })
    );
    expect(result.current[0].id).toBe(4);
    expect(result.current[3].id).toBe(1);
  });

  it('combines search and status filter', () => {
    const data = [
      { id: 1, name: 'Test Active', status: 'active' },
      { id: 2, name: 'Test Inactive', status: 'inactive' },
      { id: 3, name: 'Another Active', status: 'active' },
    ];

    const { result } = renderHook(() =>
      useFilteredAndSorted(
        data,
        'test',
        { field: '' },
        { searchField: 'name', statusField: 'status', statusValue: 'active' }
      )
    );

    expect(result.current).toHaveLength(1);
    expect(result.current[0].name).toBe('Test Active');
  });

  it('handles empty data array', () => {
    const { result } = renderHook(() =>
      useFilteredAndSorted([], '', { field: 'name', direction: 'asc' })
    );
    expect(result.current).toEqual([]);
  });

  it('handles date sorting', () => {
    const dateData = [
      { id: 1, date: new Date('2024-01-15') },
      { id: 2, date: new Date('2024-01-10') },
      { id: 3, date: new Date('2024-01-20') },
    ];

    const { result } = renderHook(() =>
      useFilteredAndSorted(dateData, '', { field: 'date', direction: 'asc' })
    );

    expect(result.current[0].id).toBe(2);
    expect(result.current[2].id).toBe(3);
  });

  it('memoizes result correctly', () => {
    const { result, rerender } = renderHook(
      ({ data, search, filterConfig }) => useFilteredAndSorted(data, search, { field: 'id' }, filterConfig),
      { 
        initialProps: { 
          data: mockData, 
          search: '',
          filterConfig: { searchField: 'name' }
        } 
      }
    );

    const firstResult = result.current;
    expect(firstResult.length).toBe(4);

    // Re-render with same search - should still be filtered
    rerender({ data: mockData, search: '', filterConfig: { searchField: 'name' } });
    expect(result.current.length).toBe(4);

    // Re-render with different search
    rerender({ data: mockData, search: 'alice', filterConfig: { searchField: 'name' } });
    expect(result.current.length).toBe(1);
    expect(result.current[0].name).toBe('Alice');
  });
});
