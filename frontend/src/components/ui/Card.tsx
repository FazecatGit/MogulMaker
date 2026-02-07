import React from 'react';

interface CardProps {
  children: React.ReactNode;
  padding?: 'sm' | 'md' | 'lg';
  className?: string;
}

const paddingMap = {
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6',
};

export default function Card({ children, padding = 'lg', className = '' }: CardProps) {
  return (
    <div className={`bg-slate-800 rounded-lg border border-slate-700 ${paddingMap[padding]} ${className}`}>
      {children}
    </div>
  );
}
