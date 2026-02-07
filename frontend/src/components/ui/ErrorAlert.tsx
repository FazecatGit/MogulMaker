import React from 'react';
import { AlertCircle } from 'lucide-react';

interface ErrorAlertProps {
  message: string;
  title?: string;
  onClose?: () => void;
  variant?: 'error' | 'warning';
}

export default function ErrorAlert({
  message,
  title,
  onClose,
  variant = 'error',
}: ErrorAlertProps) {
  const colors = {
    error: 'bg-red-500/20 border border-red-500/50 text-red-400',
    warning: 'bg-yellow-500/20 border border-yellow-500/50 text-yellow-400',
  };

  return (
    <div className={`rounded-lg p-3 flex gap-2 text-sm ${colors[variant]}`}>
      <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
      <div className="flex-1">
        {title && <p className="font-semibold">{title}</p>}
        <p>{message}</p>
      </div>
      {onClose && (
        <button
          onClick={onClose}
          className="text-current hover:opacity-70 transition flex-shrink-0"
        >
          ✕
        </button>
      )}
    </div>
  );
}
