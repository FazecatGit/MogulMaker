/**
 * Notification Toast Component
 * 
 * Displays notifications from Zustand global store
 * Auto-dismisses after 5 seconds
 */

'use client';

import { useEffect } from 'react';
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react';
import { useGlobalStore } from '@/store/useGlobalStore';

export default function NotificationToast() {
  const notifications = useGlobalStore((state) => state.notifications);
  const markNotificationRead = useGlobalStore((state) => state.markNotificationRead);

  // Auto-dismiss notifications after 5 seconds
  useEffect(() => {
    const unreadNotifications = notifications.filter((n) => !n.read);
    
    unreadNotifications.forEach((notification) => {
      const age = Date.now() - notification.timestamp;
      const remainingTime = 5000 - age;
      
      if (remainingTime > 0) {
        setTimeout(() => {
          markNotificationRead(notification.id);
        }, remainingTime);
      } else {
        markNotificationRead(notification.id);
      }
    });
  }, [notifications, markNotificationRead]);

  const visibleNotifications = notifications.filter((n) => !n.read).slice(0, 3);

  if (visibleNotifications.length === 0) return null;

  const getIcon = (type: string) => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-400" />;
      case 'error':
        return <AlertCircle className="w-5 h-5 text-red-400" />;
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-400" />;
      case 'info':
      default:
        return <Info className="w-5 h-5 text-blue-400" />;
    }
  };

  const getStyles = (type: string) => {
    switch (type) {
      case 'success':
        return 'bg-green-900/90 border-green-600';
      case 'error':
        return 'bg-red-900/90 border-red-600';
      case 'warning':
        return 'bg-yellow-900/90 border-yellow-600';
      case 'info':
      default:
        return 'bg-blue-900/90 border-blue-600';
    }
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-md">
      {visibleNotifications.map((notification) => (
        <div
          key={notification.id}
          className={`flex items-start gap-3 p-4 rounded-lg border-2 shadow-lg backdrop-blur-sm animate-slide-in ${getStyles(notification.type)}`}
        >
          {getIcon(notification.type)}
          <div className="flex-1">
            <p className="font-semibold text-white">{notification.title}</p>
            <p className="text-sm text-slate-200 mt-1">{notification.message}</p>
          </div>
          <button
            onClick={() => markNotificationRead(notification.id)}
            className="text-slate-300 hover:text-white transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  );
}
