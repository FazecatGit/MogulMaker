import React from 'react';

interface StatCardProps {
  label: string;
  value: string | number;
  subtext?: string;
  variant?: 'default' | 'positive' | 'negative' | 'neutral' | 'warning';
  icon?: React.ReactNode;
}

export default function StatCard({
  label,
  value,
  subtext,
  variant = 'default',
  icon,
}: StatCardProps) {
  const variantColors = {
    default: 'text-slate-300',
    positive: 'text-green-400',
    negative: 'text-red-400',
    neutral: 'text-blue-400',
    warning: 'text-yellow-400',
  };

  return (
    <div className="bg-slate-700/50 rounded-lg p-4 border border-slate-600">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-slate-400 text-xs mb-1 font-medium">{label}</p>
          <p className={`text-2xl font-bold ${variantColors[variant]}`}>
            {value}
          </p>
          {subtext && <p className="text-slate-400 text-xs mt-1">{subtext}</p>}
        </div>
        {icon && <div className="text-slate-400 flex-shrink-0">{icon}</div>}
      </div>
    </div>
  );
}
