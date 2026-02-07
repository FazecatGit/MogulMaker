interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export default function SearchInput({
  value,
  onChange,
  placeholder = 'Search...',
  className = '',
}: SearchInputProps) {
  return (
    <input
      type="text"
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={`bg-slate-700 text-white placeholder-slate-400 rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none ${className}`}
    />
  );
}
