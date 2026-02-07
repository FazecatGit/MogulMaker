/**
 * Zustand Global State Store
 * 
 * Centralizes state management for:
 * - User preferences
 * - Real-time notifications
 * - Trading alerts
 * - Cached data for performance
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  timestamp: number;
  read: boolean;
}

interface Alert {
  id: string;
  symbol: string;
  type: 'price' | 'pnl' | 'risk';
  condition: string;
  value: number;
  triggered: boolean;
  createdAt: number;
}

interface UserPreferences {
  theme: 'dark' | 'light';
  defaultChartPeriod: '1D' | '1W' | '1M' | '3M' | '1Y';
  notificationsEnabled: boolean;
  soundEnabled: boolean;
  refreshInterval: number; // seconds
}

interface GlobalState {
  // User Preferences
  preferences: UserPreferences;
  setPreferences: (preferences: Partial<UserPreferences>) => void;

  // Notifications
  notifications: Notification[];
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void;
  markNotificationRead: (id: string) => void;
  clearNotifications: () => void;
  unreadCount: () => number;

  // Alerts
  alerts: Alert[];
  addAlert: (alert: Omit<Alert, 'id' | 'triggered' | 'createdAt'>) => void;
  removeAlert: (id: string) => void;
  triggerAlert: (id: string) => void;
  clearAlerts: () => void;

  // Real-time Data Status
  lastUpdate: number;
  setLastUpdate: (timestamp: number) => void;
  isConnected: boolean;
  setIsConnected: (connected: boolean) => void;
}

export const useGlobalStore = create<GlobalState>()(
  persist(
    (set, get) => ({
      // Default Preferences
      preferences: {
        theme: 'dark',
        defaultChartPeriod: '1M',
        notificationsEnabled: true,
        soundEnabled: false,
        refreshInterval: 5,
      },
      setPreferences: (newPreferences) =>
        set((state) => ({
          preferences: { ...state.preferences, ...newPreferences },
        })),

      // Notifications
      notifications: [],
      addNotification: (notification) =>
        set((state) => ({
          notifications: [
            {
              ...notification,
              id: `notif-${Date.now()}-${Math.random()}`,
              timestamp: Date.now(),
              read: false,
            },
            ...state.notifications,
          ].slice(0, 50), // Keep max 50 notifications
        })),
      markNotificationRead: (id) =>
        set((state) => ({
          notifications: state.notifications.map((n) =>
            n.id === id ? { ...n, read: true } : n
          ),
        })),
      clearNotifications: () => set({ notifications: [] }),
      unreadCount: () => get().notifications.filter((n) => !n.read).length,

      // Alerts
      alerts: [],
      addAlert: (alert) =>
        set((state) => ({
          alerts: [
            ...state.alerts,
            {
              ...alert,
              id: `alert-${Date.now()}-${Math.random()}`,
              triggered: false,
              createdAt: Date.now(),
            },
          ],
        })),
      removeAlert: (id) =>
        set((state) => ({
          alerts: state.alerts.filter((a) => a.id !== id),
        })),
      triggerAlert: (id) =>
        set((state) => ({
          alerts: state.alerts.map((a) =>
            a.id === id ? { ...a, triggered: true } : a
          ),
        })),
      clearAlerts: () => set({ alerts: [] }),

      // Real-time Status
      lastUpdate: Date.now(),
      setLastUpdate: (timestamp) => set({ lastUpdate: timestamp }),
      isConnected: true,
      setIsConnected: (connected) => set({ isConnected: connected }),
    }),
    {
      name: 'mogul-maker-storage', // localStorage key
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        preferences: state.preferences,
        alerts: state.alerts,
        // Don't persist notifications and real-time status
      }),
    }
  )
);

// Selectors for optimized re-renders
export const selectPreferences = (state: GlobalState) => state.preferences;
export const selectNotifications = (state: GlobalState) => state.notifications;
export const selectUnreadCount = (state: GlobalState) => state.unreadCount();
export const selectAlerts = (state: GlobalState) => state.alerts;
export const selectIsConnected = (state: GlobalState) => state.isConnected;
