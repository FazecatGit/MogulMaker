import { AlertCircle, Check } from 'lucide-react';

interface StatusAlertProps {
  message: string;
  variant: 'error' | 'success' | 'warning';
}

const config = {
  error: {
    className: 'bg-red-500/20 border border-red-500/50 text-red-400',
    Icon: AlertCircle,
  },
  success: {
    className: 'bg-green-500/20 border border-green-500/50 text-green-400',
    Icon: Check,
  },
  warning: {
    className: 'bg-yellow-500/20 border border-yellow-500/50 text-yellow-400',
    Icon: AlertCircle,
  },
};

export default function StatusAlert({ message, variant }: StatusAlertProps) {
  const { className, Icon } = config[variant];
  return (
    <div className={`rounded-lg p-4 flex items-center gap-2 ${className}`}>
      <Icon className="w-5 h-5 flex-shrink-0" />
      {message}
    </div>
  );
}
