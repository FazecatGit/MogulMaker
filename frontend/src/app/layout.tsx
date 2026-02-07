import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import Providers from '@/components/providers/Providers';
import Header from '@/components/layout/Header';
import Sidebar from '@/components/layout/Sidebar';
import Footer from '@/components/layout/Footer';
import './globals.css';

/**
 * Root Layout Component
 * 
 * This is a SERVER component (no 'use client')
 * Allows metadata export
 * Providers are in a separate CLIENT component
 */

export const metadata: Metadata = {
  title: 'MogulMaker - Trading Dashboard',
  description: 'Portfolio management and trading dashboard',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{
          __html: `
            (function() {
              try {
                localStorage.setItem('theme', 'dark');
                document.documentElement.classList.add('dark');
              } catch (e) {}
            })()
          `
        }} />
      </head>
      <body style={{ backgroundColor: 'var(--background)', color: 'var(--foreground)' }} className="antialiased">
        {/* Providers component wraps everything with TanStack Query */}
        <Providers>
          <div className="flex h-screen">
            {/* Sidebar - full height, fixed width */}
            <Sidebar />

            {/* Main content area with header */}
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Header - top navigation */}
              <Header userName="Trader" />

              {/* Scrollable main content */}
              <main className="flex-1 overflow-auto">
                {/* Page content */}
                <div className="p-6">
                  {children}
                </div>

                {/* Footer - at bottom of scrollable content */}
                <Footer />
              </main>
            </div>
          </div>
        </Providers>
      </body>
    </html>
  );
}
