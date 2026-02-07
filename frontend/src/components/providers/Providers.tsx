'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode, useEffect } from 'react';

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

function DarkModeInitializer({ children }: { children: ReactNode }) {
  useEffect(() => {
    // Always force dark mode
    document.documentElement.classList.add('dark');
    localStorage.setItem('theme', 'dark');
  }, []);

  return children;
}

function AuthInitializer({ children }: { children: ReactNode }) {
  useEffect(() => {
    const initializeAuth = async () => {
      // Check if token already exists and is still valid
      const existingToken = localStorage.getItem('authToken');
      if (existingToken) {
        return; // Token already exists, no need to generate a new one
      }

      try {
        // Generate a new JWT token for this session
        const response = await fetch('http://localhost:3000/api/token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            user_id: 'default-trader',
            email: 'trader@mogulmaker.local',
          }),
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('authToken', data.token);
        } else {
          console.error('Failed to generate auth token');
        }
      } catch (error) {
        console.error('Error initializing auth:', error);
      }
    };

    initializeAuth();
  }, []);

  return children;
}

export default function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <DarkModeInitializer>
        <AuthInitializer>{children}</AuthInitializer>
      </DarkModeInitializer>
    </QueryClientProvider>
  );
}
