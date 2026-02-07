import { renderHook, act } from '@testing-library/react';
import { useDebounce, useDebouncedCallback } from '../useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('returns initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('initial', 500));
    expect(result.current).toBe('initial');
  });

  it('debounces value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'first', delay: 500 } }
    );

    expect(result.current).toBe('first');

    // Change value
    rerender({ value: 'second', delay: 500 });
    expect(result.current).toBe('first'); // Still old value

    // Fast-forward time
    act(() => {
      jest.advanceTimersByTime(500);
    });

    expect(result.current).toBe('second'); // Now updated
  });

  it('resets timer on rapid changes', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 500),
      { initialProps: { value: 'first' } }
    );

    rerender({ value: 'second' });
    act(() => {
      jest.advanceTimersByTime(300);
    });

    rerender({ value: 'third' });
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Should still be 'first' because timer was reset
    expect(result.current).toBe('first');

    // Now complete the timer
    act(() => {
      jest.advanceTimersByTime(200);
    });

    expect(result.current).toBe('third');
  });

  it('uses custom delay', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 1000),
      { initialProps: { value: 'first' } }
    );

    rerender({ value: 'second' });

    act(() => {
      jest.advanceTimersByTime(500);
    });
    expect(result.current).toBe('first');

    act(() => {
      jest.advanceTimersByTime(500);
    });
    expect(result.current).toBe('second');
  });
});

describe('useDebouncedCallback', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('calls callback after delay', () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useDebouncedCallback(callback, 500));

    act(() => {
      result.current('test');
    });

    expect(callback).not.toHaveBeenCalled();

    act(() => {
      jest.advanceTimersByTime(500);
    });

    expect(callback).toHaveBeenCalledWith('test');
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('cancels previous timeout on rapid calls', () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useDebouncedCallback(callback, 500));

    act(() => {
      result.current('first');
    });

    act(() => {
      jest.advanceTimersByTime(300);
    });

    act(() => {
      result.current('second');
    });

    act(() => {
      jest.advanceTimersByTime(300);
    });

    act(() => {
      result.current('third');
    });

    act(() => {
      jest.advanceTimersByTime(500);
    });

    // Only the last call should execute
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith('third');
  });

  it('passes multiple arguments correctly', () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useDebouncedCallback(callback, 500));

    act(() => {
      result.current('arg1', 'arg2', 'arg3');
    });

    act(() => {
      jest.advanceTimersByTime(500);
    });

    expect(callback).toHaveBeenCalledWith('arg1', 'arg2', 'arg3');
  });

  it('cleans up timeout on unmount', () => {
    const callback = jest.fn();
    const { result, unmount } = renderHook(() => useDebouncedCallback(callback, 500));

    act(() => {
      result.current('test');
    });

    unmount();

    act(() => {
      jest.advanceTimersByTime(500);
    });

    // Callback should not be called after unmount
    expect(callback).not.toHaveBeenCalled();
  });
});
