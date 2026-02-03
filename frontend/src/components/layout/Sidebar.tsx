'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  TrendingUp,
  Wallet,
  BarChart3,
  Clock,
  AlertCircle,
  Settings,
  ChevronDown,
  Search,
  Newspaper,
  Microscope,
} from 'lucide-react';

/**
 * Sidebar Component
 * 
 * CONCEPTS EXPLAINED:
 * - useState: React hook for state management (track what page user is on)
 * - usePathname: Next.js hook to get current URL path
 * - Conditional Rendering: Show/hide elements based on conditions
 * - Array Mapping: Loop through navigation items and create buttons
 */

interface NavItem {
  label: string;
  href: string;
  icon: React.ReactNode;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: <LayoutDashboard className="w-5 h-5" /> },
  { label: 'Positions', href: '/positions', icon: <TrendingUp className="w-5 h-5" /> },
  { label: 'Trades', href: '/trades', icon: <Wallet className="w-5 h-5" /> },
  { label: 'Scouter', href: '/scouter', icon: <Search className="w-5 h-5" /> },
  { label: 'News', href: '/news', icon: <Newspaper className="w-5 h-5" /> },
  { label: 'Analyzer', href: '/analyzer', icon: <Microscope className="w-5 h-5" /> },
  { label: 'Backtest', href: '/backtest', icon: <BarChart3 className="w-5 h-5" /> },
  { label: 'Watchlist', href: '/watchlist', icon: <Clock className="w-5 h-5" /> },
  { label: 'Risk', href: '/risk', icon: <AlertCircle className="w-5 h-5" /> },
  { label: 'Settings', href: '/settings', icon: <Settings className="w-5 h-5" /> },
];

export default function Sidebar() {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();

  // Check if a nav item is active (current page)
  const isActive = (href: string) => pathname === href || pathname.startsWith(href + '/');

  return (
    <>
      {/* Mobile Toggle Button - only visible on small screens */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="fixed top-20 left-4 z-40 md:hidden bg-blue-600 p-2 rounded-lg text-white"
      >
        <ChevronDown className="w-5 h-5" />
      </button>

      {/* Sidebar Container */}
      <aside
        className={`
          fixed left-0 top-16 h-[calc(100vh-4rem)] w-64 bg-slate-900 border-r border-slate-700
          transform transition-transform duration-300 ease-in-out
          ${isOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0
          overflow-y-auto z-30
        `}
      >
        <nav className="p-4 space-y-2">
          {/* Map through navigation items and create a button for each */}
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setIsOpen(false)} // Close sidebar on mobile after click
              className={`
                flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200
                ${
                  isActive(item.href)
                    ? 'bg-blue-600 text-white shadow-lg' // Active state (highlighted)
                    : 'text-slate-300 hover:bg-slate-800 hover:text-white' // Inactive state
                }
              `}
            >
              {item.icon}
              <span className="font-medium">{item.label}</span>
            </Link>
          ))}
        </nav>
      </aside>
    </>
  );
}
