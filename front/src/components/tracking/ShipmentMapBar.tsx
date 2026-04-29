import { MapPin } from 'lucide-react';
import { ShipmentData } from '@/types/shipment';
import dynamic from 'next/dynamic';

const DynamicMap = dynamic(() => import('@/components/map/DynamicMap'), {
  ssr: false,
});

import { Dictionary } from '@/lib/dictionaries';

interface ShipmentMapBarProps {
  shippingData: ShipmentData;
  originCoords: [number, number];
  destCoords: [number, number];
  liveProgress: number;
  dict: Dictionary;
}

export function ShipmentMapBar({ shippingData, originCoords, destCoords, liveProgress, dict }: ShipmentMapBarProps) {
  return (
    <>
      <div className="grid grid-cols-3 gap-1.5 md:gap-4 mb-4 md:mb-6 w-full max-w-6xl mx-auto animate-fade-in md:px-0">
        <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-surface/50 text-center md:text-left">
          <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-text-muted/10 flex items-center justify-center shrink-0">
            <MapPin size={12} className="text-text-muted md:w-5 md:h-5" />
          </div>
          <div className="min-w-0">
            <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-text-muted block mb-0.5 truncate">{dict.shipment.from || 'Origin'}</span>
            <span className="text-[8px] md:text-sm font-black text-text-main truncate block">{shippingData.senderCountry}</span>
          </div>
        </div>

        <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-accent/5 border-accent/20 text-center md:text-left">
          <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-accent/10 flex items-center justify-center shrink-0">
            <MapPin size={12} className="text-accent md:w-5 md:h-5" />
          </div>
          <div className="min-w-0">
            <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-accent block mb-0.5 truncate">{dict.shipment.destination || 'Destination'}</span>
            <span className="text-[8px] md:text-sm font-black text-text-main truncate block">{shippingData.receiverCountry}</span>
          </div>
        </div>

        <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-surface/50 text-center md:text-left">
          <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-surface/10 flex items-center justify-center shrink-0">
            <div className="relative flex h-1.5 w-1.5 md:h-3 md:w-3">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
              <span className="relative inline-flex rounded-full h-1.5 w-1.5 md:h-3 md:w-3 bg-accent"></span>
            </div>
          </div>
          <div className="min-w-0">
            <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-accent block mb-0.5 truncate">{dict.shipment.live || 'Live'} </span>
            <span className="text-[8px] md:text-sm font-black text-text-main uppercase tracking-tighter truncate block">
              {dict.statuses?.[shippingData.status] || shippingData.status.replace(/_/g, ' ')}
            </span>
          </div>
        </div>
      </div>

      <div className="mb-16 relative z-10 animate-fade-in delay-300 w-full max-w-6xl mx-auto shadow-2xl rounded-[2rem] overflow-hidden">
        <DynamicMap 
          origin={originCoords} 
          destination={destCoords} 
          progress={liveProgress}
          shipment={shippingData}
          dict={dict}
        />
      </div>
    </>
  );
}
