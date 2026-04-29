import { MapPin, Package } from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { ShipmentData } from '@/types/shipment';
import { Dictionary } from '@/lib/dictionaries';

interface ShipmentTimelineProps {
  shippingData: ShipmentData;
  dict: Dictionary;
}

export function ShipmentTimeline({ shippingData, dict }: ShipmentTimelineProps) {
  return (
    <div className="space-y-8 order-1 lg:order-2">
      <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
        {dict.shipment.latestUpdates}
        <span className="h-px flex-1 bg-accent/20" />
      </h4>
      <div className="space-y-4 md:space-y-12 relative border-l-2 border-border ml-2 md:ml-3 pl-6 md:pl-10 py-2">
        {shippingData.timeline ? (
          shippingData.timeline.map((event, i) => (
            <motion.div
              key={i}
              className="relative group/event"
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: i * 0.15, ease: "easeOut" }}
            >
              <div className={cn(
                "absolute left-[-32px] md:left-[-52px] top-1.5 w-4 h-4 md:w-8 md:h-8 rounded-full border-[3px] md:border-4 border-bg transition-all duration-500",
                event.is_completed ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border opacity-50"
              )}>
                {event.is_completed && !shippingData.timeline?.[i + 1]?.is_completed && (
                  <motion.div
                    className="absolute inset-0 rounded-full border-2 border-accent"
                    animate={{ scale: [1, 2], opacity: [0.6, 0] }}
                    transition={{ duration: 1.5, repeat: Infinity, ease: "easeOut" }}
                  />
                )}
              </div>
              <div className="flex flex-wrap items-center gap-2 md:gap-4 mb-2">
                <p className={cn("font-black text-sm md:text-lg lg:text-xl tracking-tight leading-none uppercase", event.is_completed ? "text-text-main" : "text-text-muted opacity-50")}>
                  {dict.statuses?.[event.status as keyof typeof dict.statuses] || event.status}
                </p>
                {event.is_completed && !shippingData.timeline?.[i + 1]?.is_completed && (
                  <span className="bg-accent text-white text-[8px] md:text-[9px] font-black uppercase px-2 md:px-2 py-0.5 md:py-1 rounded-lg animate-pulse tracking-wider md:tracking-widest">{dict.shipment.live}</span>
                )}
              </div>
              <motion.span
                className="text-[7px] md:text-[10px] font-black text-accent/60 uppercase tracking-[0.15em] md:tracking-[0.2em] block mb-4"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.4, delay: i * 0.15 + 0.2 }}
              >
                {new Date(event.timestamp).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
              </motion.span>
              <motion.div
                className="mt-4 md:mt-4 flex items-start gap-2 md:gap-4 text-xs md:text-sm font-bold text-text-muted bg-surface-muted/50 p-4 md:p-4 rounded-xl md:rounded-2xl border border-border group-hover/event:border-accent/20 transition-colors"
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: i * 0.15 + 0.3 }}
              >
                <Package size={14} className="text-accent/60 mt-0.5 md:w-4 md:h-4 shrink-0" />
                <span className="leading-relaxed">{event.description}</span>
              </motion.div>
              {event.is_completed && (
                <>
                  <motion.div
                    className="mt-2 md:mt-3 flex items-center gap-2 text-[10px] md:text-xs font-bold text-text-muted/70 bg-surface/30 px-3 py-2 rounded-lg border border-border/50"
                    initial={{ opacity: 0, y: 6 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.4, delay: i * 0.15 + 0.4 }}
                  >
                    <MapPin size={12} className="text-accent/50 shrink-0 md:w-3.5 md:h-3.5" />
                    <span className="truncate">
                      {event.status.toLowerCase().includes('delivered') || event.status.toLowerCase().includes('arrival')
                        ? `${shippingData.receiverCountry} ${dict.shipment.destinationLabel || '(Destination)'}`
                        : event.status.toLowerCase().includes('depart') || event.status.toLowerCase().includes('origin')
                          ? `${shippingData.senderCountry} ${dict.shipment.originLabel || '(Origin)'}`
                          : (dict.shipment.inTransitLabel || 'In Transit')}
                    </span>
                    <span className="text-accent/30">•</span>
                    <span className="text-[9px] md:text-[10px] opacity-60 truncate">{new Date(event.timestamp).toLocaleDateString()}</span>
                  </motion.div>

                  <motion.div
                    className="mt-2 flex items-center justify-between gap-2 text-[9px] md:text-[10px] text-text-muted/60 bg-surface/20 px-3 py-1.5 rounded-md border border-border/30"
                    initial={{ opacity: 0, y: 4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.4, delay: i * 0.15 + 0.5 }}
                  >
                    <div className="flex items-center gap-1.5">
                      <span className="w-1.5 h-1.5 rounded-full bg-accent/60"></span>
                      <span className="font-bold uppercase tracking-wider">{dict.shipment.airFreight || 'Air Freight'}</span>
                    </div>
                    <span className="opacity-50">|</span>
                    <div className="flex items-center gap-1">
                      <span className="font-mono">ID: #{shippingData.trackingNumber.slice(-6)}</span>
                    </div>
                  </motion.div>
                </>
              )}
            </motion.div>
          ))
        ) : (
          shippingData.events?.map((event, i: number) => (
            <div key={event.id} className="relative group/event">
              <div className={cn(
                "absolute left-[-33px] md:left-[-51px] top-1.5 w-6 h-6 rounded-full border-4 border-bg transition-all duration-500",
                i === 0 ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border group-hover/event:bg-accent/40"
              )} />
              <div className="flex items-center gap-4 mb-2">
                <p className={cn("font-black text-lg md:text-xl tracking-tight leading-none uppercase", i === 0 ? "text-text-main" : "text-text-muted")}>
                  {event.status.replace(/_/g, ' ')}
                </p>
                {i === 0 && (
                  <span className="bg-accent text-white text-[9px] font-black uppercase px-2.5 py-1 rounded-lg animate-pulse tracking-widest">{dict.shipment.live}</span>
                )}
              </div>
              <span className="text-[10px] font-black text-accent/60 uppercase tracking-[0.2em]">
                {new Date(event.timestamp).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
              </span>
              <div className="mt-4 flex items-center gap-3 text-sm font-bold text-text-muted bg-surface-muted/50 p-4 rounded-2xl border border-border group-hover/event:border-accent/20 transition-colors">
                <MapPin size={16} className="text-accent/60" />
                {event.location}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
