'use client';

import { BarChart3 } from 'lucide-react';
import Link from 'next/link';

/**
 * Header Component
 * 
 * CONCEPTS EXPLAINED:
 * - 'use client': Tells Next.js this runs in the browser (client-side)
 * - React Functional Component: A function that returns JSX (HTML-like syntax)
 * - Props: Component receives data from parent
 * - Tailwind CSS: Utility classes for styling (no CSS files needed)
 */

interface HeaderProps {
  userName?: string;
}

export default function Header({ userName = 'Trader' }: HeaderProps) {
  return (
    // Header container - sticky at top, dark background
    <header className="sticky top-0 z-50 bg-gradient-to-r from-slate-900 to-slate-800 border-b border-slate-700 shadow-lg">
      <div className="flex items-center justify-between px-6 py-4 max-w-screen-2xl mx-auto">
        
        {/* Logo Section */}
        <Link href="/dashboard" className="flex items-center gap-2 hover:opacity-80 transition">
          <BarChart3 className="w-8 h-8 text-blue-400" />
          <span className="text-xl font-bold text-white">MogulMaker</span>
        </Link>

        {/* User Section */}
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm text-slate-400">Hello,</p>
            <p className="text-white font-semibold">{userName}</p>
          </div>
        </div>
      </div>
    </header>
  );
}
