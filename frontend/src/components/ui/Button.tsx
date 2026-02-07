import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'success';
  icon?: React.ReactNode;
  loading?: boolean;
  children?: React.ReactNode;
}

export default function Button({
  variant = 'primary',
  icon,
  loading = false,
  disabled,
  className,
  children,
  ...props
}: ButtonProps) {
  const baseClasses = 'flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-semibold transition';

  const variantClasses = {
    primary: 'bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 text-white',
    secondary: 'bg-slate-700 hover:bg-slate-600 text-white',
    danger: 'bg-red-600 hover:bg-red-700 disabled:bg-slate-600 text-white',
    success: 'bg-green-600 hover:bg-green-700 disabled:bg-slate-600 text-white',
  };

  return (
    <button
      disabled={disabled || loading}
      className={`${baseClasses} ${variantClasses[variant]} ${className || ''}`}
      {...props}
    >
      {loading && <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />}
      {!loading && icon}
      {children}
    </button>
  );
}
