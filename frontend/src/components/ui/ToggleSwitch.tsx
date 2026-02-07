interface ToggleSwitchProps {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

export default function ToggleSwitch({ label, description, checked, onChange }: ToggleSwitchProps) {
  return (
    <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
      <div>
        <p className="font-semibold text-white">{label}</p>
        {description && <p className="text-xs text-slate-400">{description}</p>}
      </div>
      <button
        onClick={() => onChange(!checked)}
        className={`relative w-14 h-8 rounded-full transition ${
          checked ? 'bg-blue-600' : 'bg-slate-600'
        }`}
      >
        <div
          className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
            checked ? 'translate-x-6' : ''
          }`}
        />
      </button>
    </div>
  );
}
