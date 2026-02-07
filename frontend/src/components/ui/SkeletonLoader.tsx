interface SkeletonLoaderProps {
  count?: number;
  /** Height in tailwind class, e.g. "h-16", "h-24" */
  height?: string;
  /** Show inner content lines */
  withContent?: boolean;
}

export default function SkeletonLoader({
  count = 3,
  height = 'h-16',
  withContent = false,
}: SkeletonLoaderProps) {
  return (
    <div className="space-y-2">
      {Array.from({ length: count }, (_, i) => (
        <div
          key={i}
          className={`bg-slate-800 rounded-lg ${height} animate-pulse border border-slate-700 ${withContent ? 'p-4' : ''}`}
        >
          {withContent && (
            <>
              <div className="h-5 bg-slate-700 rounded w-1/3 mb-3" />
              <div className="h-4 bg-slate-700 rounded w-full" />
            </>
          )}
        </div>
      ))}
    </div>
  );
}
