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

  // Plane position: map 0-100% progress → 10%-90% left, with S-curve vertical offset
  const planeLeft = 10 + (journeyProgress / 100) * 80;
  const normalizedP = journeyProgress / 100;
  // sin(2π·p) gives: 0 → +1 → 0 → -1 → 0, matching the S-curve up-center-down-center
  const planeTop = 50 - 15 * Math.sin(normalizedP * Math.PI * 2);
  
  const { scrollY } = useScroll();
  const scale = useTransform(scrollY, [0, 100], [1, 0.98]);

  return (
    <motion.div 
      style={{ scale }}
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="sticky top-0 z-40 w-full pt-4 pb-16 mb-8 bg-surface/90 backdrop-blur-3xl border-b border-border/50 shadow-sm transition-all duration-300 overflow-x-clip"
    >
      <div className="max-w-4xl mx-auto px-4 md:px-6 overflow-visible">
        {/* Card Header with Live Badge */}
        <div className="flex flex-wrap justify-between items-end mb-8 gap-2 px-1">
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

        <div className="relative h-20 sm:h-24 md:h-32 flex items-center justify-between px-2 sm:px-6 md:px-16">
          
          {/* SVG Path - Centered within [0, 100] viewbox, spanning 20% to 80% to avoid edge clipping */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none" preserveAspectRatio="none" viewBox="0 0 100 100">
            <path
              id="journey-path"
              d="M 10 50 Q 30 20, 50 50 T 90 50"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeDasharray="4 6"
              className="text-text-muted/50"
            />
            
            <motion.path
              d="M 10 50 Q 30 20, 50 50 T 90 50"
              fill="none"
              stroke="var(--color-accent)"
              strokeWidth="3"
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
              animate={{ 
                left: `${planeLeft}%`,
                top: `${planeTop}%`,
                x: '-50%',
                y: '-50%',
              }}
              transition={{ duration: 1.5, ease: "easeInOut" }}
            >
              <motion.div
                animate={{ 
                  y: [0, -4, 0, -2, 0],
                  rotate: [5, 12, 5, -2, 5] 
                }}
                transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
              >
                <Plane className="w-5 h-5 sm:w-6 sm:h-6 fill-accent drop-shadow-2xl" />
              </motion.div>
            </motion.div>
          )}

          {/* Journey Steps (20%, 50%, 80%) */}
          {steps.map((step, index) => {
            const Icon = step.icon;
            const isCompleted = index <= progressIndex && status !== 'CANCELED';
            const isActive = index === progressIndex && status !== 'CANCELED';
            const positions = [10, 50, 90]; 
            
            return (
              <motion.div 
                key={step.key} 
                variants={itemVariants}
                className="absolute z-20 flex flex-col items-center"
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
                    "w-9 h-9 sm:w-12 sm:h-12 md:w-16 md:h-16 rounded-2xl sm:rounded-[20px] flex items-center justify-center border-2 transition-all duration-600",
                    isCompleted ? "text-white" : "text-text-muted border-dashed bg-surface-muted"
                  )}
                >
                  {isCompleted && !isActive ? (
                    <CheckCircle2 className="w-4 h-4 sm:w-5 sm:h-5 md:w-7 md:h-7" />
                  ) : (
                    <Icon className={cn("w-4 h-4 sm:w-5 sm:h-5 md:w-7 md:h-7", isActive && "animate-pulse")} />
                  )}
                </motion.div>

                <div className="absolute top-[100%] mt-2 sm:mt-3 flex flex-col items-center w-16 sm:w-20 md:w-24">
                   <p className={cn(
                    "text-[7px] sm:text-[9px] md:text-[11px] font-black uppercase tracking-wider sm:tracking-widest text-center leading-tight",
                    isCompleted ? "text-text-main" : "text-text-muted opacity-30"
                  )}>
                    {dict.admin?.[step.labelKey] || step.key.replace(/_/g, ' ')}
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
