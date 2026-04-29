"use client";

import dynamic from 'next/dynamic';
import { Loader2, Map as MapIcon } from 'lucide-react';

import { ShipmentData } from '@/types/shipment';
import { Dictionary } from '@/lib/dictionaries';

export interface DynamicMapProps {
  origin: [number, number];
  destination: [number, number];
  progress?: number;
  shipment?: ShipmentData | null;
  dict?: Dictionary;
}

const DynamicMap = dynamic<DynamicMapProps>(() => import('./MapComponent'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[400px] sm:h-[500px] rounded-[2rem] overflow-hidden glass-panel border border-border/50 flex items-center justify-center relative shadow-2xl">
      <div className="absolute inset-0 bg-accent/5 blur-3xl rounded-full opacity-50 flex-shrink-0" />
      <div className="flex flex-col items-center gap-6 text-accent relative z-10">
        <div className="relative">
          <MapIcon className="w-12 h-12 opacity-80" />
          <Loader2 className="absolute -bottom-2 -right-2 w-6 h-6 animate-spin text-text-main" />
        </div>
        <span className="text-xs font-black uppercase tracking-widest text-text-main">Initializing Live Map...</span>
      </div>
    </div>
  )
});

export default DynamicMap;
