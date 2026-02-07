import { AlertCircle } from 'lucide-react';

interface PendingOrder {
  id: string;
  symbol: string;
  side: string;
  qty: string;
  status: string;
}

interface PendingOrdersAlertProps {
  orders: PendingOrder[];
}

export default function PendingOrdersAlert({ orders }: PendingOrdersAlertProps) {
  if (orders.length === 0) return null;

  return (
    <div className="bg-yellow-500/20 border-2 border-yellow-500 rounded-lg p-6 shadow-lg">
      <div className="flex items-start gap-4">
        <div className="bg-yellow-500 rounded-full p-2">
          <AlertCircle className="w-6 h-6 text-slate-900 flex-shrink-0" />
        </div>
        <div className="flex-1">
          <p className="text-yellow-400 font-bold text-lg mb-3">
            Pending Orders ({orders.length})
          </p>
          <div className="space-y-2">
            {orders.map((order) => (
              <div key={order.id} className="bg-slate-800/50 rounded px-3 py-2 border border-yellow-500/30">
                <span className="font-bold text-yellow-300 text-base">{order.symbol}</span>
                {' \u2022 '}
                <span className="capitalize text-white font-medium">{order.side}</span>
                {' \u2022 '}
                <span className="text-white">{parseFloat(order.qty)} shares</span>
                {' \u2022 '}
                <span className="text-yellow-400 font-semibold uppercase text-xs">{order.status}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
