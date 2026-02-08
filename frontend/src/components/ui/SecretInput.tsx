import { useState } from 'react';
import { Eye, EyeOff, Copy, Check } from 'lucide-react';

interface SecretInputProps {
  label: string;
  value: string;
  maskedValue?: string;
  onChange: (value: string) => void;
  placeholder?: string;
  subtext?: React.ReactNode;
}

export default function SecretInput({
  label,
  value,
  maskedValue = '****...****',
  onChange,
  placeholder,
  subtext,
}: SecretInputProps) {
  const [show, setShow] = useState(false);
  const [copied, setCopied] = useState(false);
  const [isFocused, setIsFocused] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleFocus = () => {
    setIsFocused(true);
    if (!value) {
      // Clear the masked value when user focuses to type
      onChange('');
    }
  };

  const handleBlur = () => {
    setIsFocused(false);
  };

  const displayValue = (isFocused || value) ? value : maskedValue;

  return (
    <div>
      <label className="block text-slate-300 text-sm font-semibold mb-2">{label}</label>
      <div className="flex gap-2">
        <div className="flex-1 relative">
          <input
            type={show ? 'text' : 'password'}
            value={displayValue}
            onChange={(e) => onChange(e.target.value)}
            onFocus={handleFocus}
            onBlur={handleBlur}
            placeholder={placeholder}
            className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
          />
          <button
            onClick={() => setShow(!show)}
            type="button"
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400 hover:text-slate-300"
          >
            {show ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        <button
          onClick={handleCopy}
          type="button"
          disabled={!value}
          className="px-4 py-2 bg-slate-700 hover:bg-slate-600 disabled:bg-slate-800 disabled:cursor-not-allowed text-white rounded transition flex items-center gap-2"
        >
          {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
        </button>
      </div>
      {subtext && <p className="text-xs text-slate-400 mt-1">{subtext}</p>}
    </div>
  );
}
