'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';

/**
 * Providers Component (Client Component)
 * 
 * This is a separate client component that handles all providers
 * Keeps 'use client' logic separate from server layout
 * 
 * WHY SEPARATE?
 * - layout.tsx needs metadata (server-only)
 * - Providers need 'use client' for hooks
 * - Solution: Move providers to separate file
 */

// Create QueryClient instance (reused for all requests)
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // Data stays fresh for 5 minutes
      retry: 1, // Retry failed requests once
      refetchOnWindowFocus: false, // Don't refetch when user switches tabs
    },
  },
});

export default function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
