import { cn } from '@/lib/utils';
import { ShipmentData } from '@/types/shipment';

import { Dictionary } from '@/lib/dictionaries';

interface ShipmentDetailsProps {
  shippingData: ShipmentData;
  dict: Dictionary;
}

export function ShipmentDetails({ shippingData, dict }: ShipmentDetailsProps) {
  return (
    <div className="space-y-16 order-2 lg:order-1">
      <div className="space-y-8">
        <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
          {dict.shipment.details}
          <span className="h-px flex-1 bg-accent/20" />
        </h4>
        <div className="space-y-2">
          {[
            { label: dict.shipment.receiver, value: shippingData.receiverName },
            { label: dict.shipment.destination, value: shippingData.receiverCountry, italic: true },
            { label: dict.shipment.from, value: shippingData.senderName },
            { label: dict.shipment.origin, value: shippingData.senderCountry },
            { label: dict.shipment.weight || "Weight", value: `${shippingData.weight || 15} KGS` },
          ].map((detail, idx) => (
            <div key={idx} className="flex justify-between items-center py-2 md:py-6 border-b border-border last:border-0 group/item">
              <span className="text-text-muted font-black text-[7px] md:text-sm uppercase tracking-widest">{detail.label}</span>
              <span className={cn("font-black text-text-main text-[10px] md:text-xl group-hover:text-accent transition-colors", detail.italic && "italic")}>{detail.value}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
