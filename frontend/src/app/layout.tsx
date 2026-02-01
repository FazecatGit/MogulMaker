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
    <html lang="en">
      <body className="bg-slate-950 text-white antialiased">
        {/* Providers component wraps everything with TanStack Query */}
        <Providers>
          {/* Header - top navigation */}
          <Header userName="Trader" />

          <div className="flex">
            {/* Sidebar - left navigation */}
            <Sidebar />

            {/* Main Content Area */}
            <main className="flex-1 md:ml-64 min-h-screen flex flex-col">
              {/* Page content goes here */}
              <div className="flex-1 p-6">
                {children}
              </div>

              {/* Footer - always at bottom */}
              <Footer />
            </main>
          </div>
        </Providers>
      </body>
    </html>
  );
}
