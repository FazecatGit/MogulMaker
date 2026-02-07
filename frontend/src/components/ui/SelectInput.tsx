interface SelectOption {
  value: string;
  label: string;
}

interface SelectInputProps {
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
  className?: string;
}

export default function SelectInput({ value, onChange, options, className = '' }: SelectInputProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={`bg-slate-700 text-white rounded px-4 py-2 border border-slate-600 focus:border-blue-500 focus:outline-none ${className}`}
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}
