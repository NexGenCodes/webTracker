import { AlertCircle, CheckCircle } from 'lucide-react';
import { Dictionary } from '@/lib/dictionaries';

interface ShipmentTerminalStateProps {
  type: 'canceled' | 'delivered';
  dict: Dictionary;
}

export function ShipmentTerminalState({ type, dict }: ShipmentTerminalStateProps) {
  if (type === 'canceled') {
    return (
      <div className="py-24 flex flex-col items-center text-center">
        <div className="w-32 h-32 bg-error/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner">
          <AlertCircle className="w-16 h-16 text-error" />
        </div>
        <div className="max-w-2xl">
          <h3 className="text-2xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-error">{dict.shipment.canceledTitle || 'Shipment Canceled'}</h3>
          <p className="text-text-muted text-lg md:text-xl leading-relaxed font-bold opacity-80">
            {dict.shipment.canceledDesc || 'This shipment has been canceled by the administrator. Please contact support for more details.'}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="py-24 flex flex-col items-center text-center">
      <div className="w-32 h-32 bg-success/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner rotate-3 transition-transform hover:rotate-0 duration-500">
        <CheckCircle className="w-16 h-16 text-success" />
      </div>
      <div className="max-w-2xl">
        <h3 className="text-2xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-gradient">{dict.shipment.deliveredTitle}</h3>
        <p className="text-text-muted text-lg md:text-xl leading-relaxed font-bold opacity-80">
          {dict.shipment.deliveredDesc}
        </p>
      </div>
    </div>
  );
}
