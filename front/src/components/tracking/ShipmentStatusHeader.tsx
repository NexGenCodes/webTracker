import { Package, Copy } from 'lucide-react';
import { toast } from 'react-hot-toast';
import { ShipmentData } from '@/types/shipment';

import { Dictionary } from '@/lib/dictionaries';

interface ShipmentStatusHeaderProps {
  shippingData: ShipmentData;
  dict: Dictionary;
}

export function ShipmentStatusHeader({ shippingData, dict }: ShipmentStatusHeaderProps) {
  return (
    <div className="flex flex-row justify-between items-center mb-4 md:mb-16 pb-3 md:pb-12 border-b border-border relative z-10 gap-x-2">
      <div className="flex items-center gap-2 md:gap-6">
        <div className="relative">
          <div className="absolute inset-0 bg-accent blur-3xl opacity-20 animate-pulse" />
          <div className="relative w-10 h-10 md:w-24 md:h-24 bg-accent rounded-xl md:rounded-[2.5rem] flex items-center justify-center text-white shadow-2xl shadow-accent/40">
            <Package size={18} strokeWidth={2.5} className="md:hidden" />
            <Package size={36} strokeWidth={2.5} className="hidden md:block" />
          </div>
        </div>
        <div className="flex-1">
          <span className="text-accent text-[8px] md:text-[11px] font-black uppercase tracking-[0.2em] md:tracking-[0.4em] mb-0.5 md:mb-2 block">{dict.shipment.status}</span>
          <h2 className="text-xs md:text-3xl lg:text-4xl font-black text-text-main tracking-tighter uppercase leading-none">
            {shippingData.isArchived ? dict.shipment.finalized : (dict.statuses?.[shippingData.status] || shippingData.status.replace(/_/g, ' '))}
          </h2>
        </div>
      </div>

      <div className="flex flex-col items-end md:items-end w-auto">
        <span className="text-text-muted text-[7px] md:text-[11px] font-black uppercase tracking-[0.2em] md:tracking-[0.4em] mb-0.5 md:mb-3">{dict.shipment.trackingId}</span>
        <div className="flex items-center gap-1 md:gap-4 bg-surface-muted px-1.5 md:px-6 py-1 md:py-4 rounded-lg md:rounded-3xl border border-border group/copy transition-all hover:border-accent/30 shadow-inner w-auto max-w-[120px] md:max-w-none">
          <span className="font-mono text-[9px] md:text-lg lg:text-2xl font-black tracking-tighter md:tracking-widest text-text-main group-hover:text-accent transition-colors truncate">{shippingData.trackingNumber}</span>
          <div className="flex items-center gap-1 ml-auto">
            <button
              onClick={() => {
                navigator.clipboard.writeText(shippingData.trackingNumber);
                toast.success(dict.admin?.copied || "Copied!");
              }}
              className="p-1 md:p-2 rounded-xl md:rounded-2xl transition-all bg-surface hover:bg-surface-muted text-text-muted hover:text-accent border border-border flex items-center justify-center active:scale-90"
              title={dict.admin?.copy || "Copy"}
            >
              <Copy size={14} className="md:w-[18px] md:h-[18px]" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
