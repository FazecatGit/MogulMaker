import React from 'react';

interface Column<T> {
  key: keyof T;
  label: string;
  width?: string;
  align?: 'left' | 'center' | 'right';
  render?: (value: any, item: T) => React.ReactNode;
  mobileHidden?: boolean;
}

interface ResponsiveTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyExtractor: (item: T, index: number) => string | number;
  renderActions?: (item: T) => React.ReactNode;
  renderMobileCard?: (item: T) => React.ReactNode;
  emptyMessage?: string;
}

export default function ResponsiveTable<T extends Record<string, any>>({
  columns,
  data,
  keyExtractor,
  renderActions,
  renderMobileCard,
  emptyMessage = 'No data found',
}: ResponsiveTableProps<T>) {
  if (data.length === 0) {
    return (
      <div className="bg-slate-800 rounded-lg p-12 border border-slate-700 text-center">
        <p className="text-slate-400">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
      {/* Desktop View - Table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="w-full">
          <thead className="bg-slate-700/50 border-b border-slate-700">
            <tr>
              {columns.map((col) => (
                <th
                  key={String(col.key)}
                  className={`px-6 py-3 text-sm font-semibold text-slate-300 text-${
                    col.align || 'left'
                  }`}
                  style={{ width: col.width }}
                >
                  {col.label}
                </th>
              ))}
              {renderActions && (
                <th className="px-6 py-3 text-center text-sm font-semibold text-slate-300">
                  Actions
                </th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700">
            {data.map((item, idx) => (
              <tr key={keyExtractor(item, idx)} className="hover:bg-slate-700/30 transition">
                {columns.map((col) => (
                  <td
                    key={String(col.key)}
                    className={`px-6 py-4 text-${col.align || 'left'}`}
                  >
                    {col.render ? col.render(item[col.key], item) : item[col.key]}
                  </td>
                ))}
                {renderActions && (
                  <td className="px-6 py-4 text-center flex gap-2 justify-center">
                    {renderActions(item)}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile View - Cards */}
      <div className="md:hidden space-y-3 p-4">
        {data.map((item, idx) => (
          <div
            key={keyExtractor(item, idx)}
            className="bg-slate-700/50 rounded-lg p-4 border border-slate-600"
          >
            {renderMobileCard ? (
              renderMobileCard(item)
            ) : (
              <div className="space-y-2 text-sm">
                {columns
                  .filter((col) => !col.mobileHidden)
                  .map((col) => (
                    <div key={String(col.key)} className="flex justify-between">
                      <span className="text-slate-400">{col.label}:</span>
                      <span className="text-white font-semibold">
                        {col.render ? col.render(item[col.key], item) : item[col.key]}
                      </span>
                    </div>
                  ))}
              </div>
            )}
            {renderActions && (
              <div className="mt-3 flex gap-2">
                {renderActions(item)}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
