"use client";

import React from 'react';
import { CheckCircle2, Plane, Box, Building2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ShipmentStatus, Dictionary, ShipmentData } from '@/types/shipment';
import { motion, Variants, useScroll, useTransform } from 'framer-motion';

interface VisualTrackerProps {
  shipment: ShipmentData;
  dict: Dictionary;
}

const steps = [
  { key: 'IN_TRANSIT', icon: Box, labelKey: 'inTransit' },
  { key: 'OUT_FOR_DELIVERY', icon: Box, labelKey: 'outForDelivery' },
  { key: 'DELIVERED', icon: Building2, labelKey: 'delivered' },
];

const containerVariants: Variants = {
  hidden: { opacity: 0, y: -20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.2,
    },
  },
};

const itemVariants: Variants = {
  hidden: { scale: 0.8, opacity: 0 },
  visible: {
    scale: 1,
    opacity: 1,
    transition: { type: "spring", stiffness: 300, damping: 20 },
  },
};

export const VisualTracker: React.FC<VisualTrackerProps> = ({ shipment, dict }) => {
  const status = shipment.status;
  
  const currentStepIndex = steps.findIndex(step => step.key === status);
  const progressIndex = status === 'CANCELED' ? -1 : currentStepIndex;

  // Calculate dynamic progress across the journey [0% - 100%]
  // 0% = First Node (In Transit)
  // 50% = Second Node (Out for Delivery)
  // 100% = Third Node (Arrived)
  const calculateJourneyProgress = () => {
    if (status === 'CANCELED' || status === 'PENDING') return 0;
    if (status === 'DELIVERED') return 100;
    
    const now = Date.now();
    const transit = shipment.scheduledTransitTime ? new Date(shipment.scheduledTransitTime).getTime() : null;
    const outForDelivery = shipment.outfordeliveryTime ? new Date(shipment.outfordeliveryTime).getTime() : null;
    const arrival = shipment.expectedDeliveryTime ? new Date(shipment.expectedDeliveryTime).getTime() : null;

    if (status === 'IN_TRANSIT' && transit && outForDelivery) {
      // Scale from 0% to 50%
      const stepProgress = Math.min(0.95, Math.max(0.05, (now - transit) / (outForDelivery - transit)));
      return stepProgress * 50;
    }

    if (status === 'OUT_FOR_DELIVERY' && outForDelivery && arrival) {
      // Scale from 50% to 100%
      const stepProgress = Math.min(0.95, Math.max(0.05, (now - outForDelivery) / (arrival - outForDelivery)));
      return 50 + (stepProgress * 50);
    }

    // Index-based fallback
    return Math.max(0, (progressIndex / (steps.length - 1)) * 100);
  };

  const journeyProgress = calculateJourneyProgress();
  
  const { scrollY } = useScroll();
  const scale = useTransform(scrollY, [0, 100], [1, 0.98]);

  return (
    <motion.div 
      style={{ scale }}
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="sticky top-0 z-40 w-full pt-4 pb-12 mb-8 bg-surface/90 backdrop-blur-3xl border-b border-border/50 shadow-sm transition-all duration-300 overflow-hidden"
    >
      <div className="max-w-4xl mx-auto px-6 overflow-visible">
        {/* Card Header with Live Badge */}
        <div className="flex justify-between items-end mb-8 px-1">
           <div className="flex flex-col">
              <span className="text-[10px] font-black uppercase tracking-[0.3em] text-accent/60 mb-1">
                Live Shipment Radar
              </span>
              <div className="flex items-baseline gap-2">
                <h3 className="text-xl md:text-2xl font-black text-text-main capitalize">
                  {dict.admin?.[status.toLowerCase()] || status.replace(/_/g, ' ')}
                </h3>
              </div>
           </div>
           
           {(status !== 'DELIVERED' && status !== 'CANCELED') && (
             <div className="flex items-center gap-3 px-4 py-2 bg-accent/5 border border-accent/10 rounded-2xl">
               <div className="relative flex h-3 w-3">
                 <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
                 <span className="relative inline-flex rounded-full h-3 w-3 bg-accent shadow-[0_0_8px_rgba(var(--color-accent-rgb),1)]"></span>
               </div>
               <span className="text-[11px] font-black text-accent uppercase tracking-[0.1em]">
                 Live Data
               </span>
             </div>
           )}
        </div>

        <div className="relative h-24 md:h-32 flex items-center justify-between px-10 md:px-16">
          
          {/* SVG Path - Centered within [0, 100] viewbox, spanning 20% to 80% to avoid edge clipping */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none" preserveAspectRatio="none" viewBox="0 0 100 100">
            <path
              id="journey-path"
              d="M 20 50 Q 35 20, 50 50 T 80 50"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeDasharray="4 6"
              className="text-border/40"
            />
            
            <motion.path
              d="M 20 50 Q 35 20, 50 50 T 80 50"
              fill="none"
              stroke="var(--color-accent)"
              strokeWidth="2.5"
              initial={{ pathLength: 0 }}
              animate={{ pathLength: journeyProgress / 100 }}
              transition={{ duration: 1.5, ease: "easeInOut" }}
              className="drop-shadow-[0_0_12px_rgba(var(--color-accent-rgb),0.5)]"
            />
          </svg>

          {/* Mascot (Cargo Plane) */}
          {status !== 'CANCELED' && status !== 'PENDING' && status !== 'DELIVERED' && (
            <motion.div
              className="absolute z-30 text-accent"
              animate={{ offsetDistance: `${journeyProgress}%` }}
              transition={{ duration: 1.5, ease: "easeInOut" }}
              style={{ 
                offsetPath: "path('M 20 50 Q 35 20, 50 50 T 80 50')",
                transform: 'translate(-50%, -50%)' 
              }}
            >
              <motion.div
                animate={{ 
                  y: [0, -4, 0, -2, 0],
                  rotate: [5, 12, 5, -2, 5] 
                }}
                transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
              >
                <Plane size={26} className="fill-accent drop-shadow-2xl" />
              </motion.div>
            </motion.div>
          )}

          {/* Journey Steps (20%, 50%, 80%) */}
          {steps.map((step, index) => {
            const Icon = step.icon;
            const isCompleted = index <= progressIndex && status !== 'CANCELED';
            const isActive = index === progressIndex && status !== 'CANCELED';
            const positions = [20, 50, 80]; 
            
            return (
              <motion.div 
                key={step.key} 
                variants={itemVariants}
                className="relative z-20 flex flex-col items-center"
                style={{ left: `${positions[index]}%`, transform: 'translateX(-50%)' }}
              >
                <motion.div
                  initial={false}
                  animate={{
                    scale: isActive ? 1.25 : 1,
                    backgroundColor: isCompleted ? 'var(--color-accent)' : 'var(--color-surface-muted)',
                    borderColor: isCompleted ? 'var(--color-accent)' : 'var(--color-border)',
                    boxShadow: isActive ? '0 0 35px rgba(var(--color-accent-rgb), 0.45)' : 'none',
                  }}
                  className={cn(
                    "w-12 h-12 md:w-16 md:h-16 rounded-[20px] flex items-center justify-center border-2 transition-all duration-600",
                    isCompleted ? "text-white" : "text-text-muted border-dashed bg-surface-muted"
                  )}
                >
                  {isCompleted && !isActive ? (
                    <CheckCircle2 size={24} className="md:w-7 md:h-7" />
                  ) : (
                    <Icon size={24} className={cn("md:w-7 md:h-7", isActive && "animate-pulse")} />
                  )}
                </motion.div>

                <div className="absolute top-[100%] mt-3 flex flex-col items-center">
                   <p className={cn(
                    "text-[9px] md:text-[11px] font-black uppercase tracking-widest whitespace-nowrap text-center",
                    isCompleted ? "text-text-main" : "text-text-muted opacity-30"
                  )}>
                    {dict.admin?.[step.labelKey] || step.key}
                  </p>
                </div>
              </motion.div>
            );
          })}
        </div>
      </div>
    </motion.div>
  );
};
