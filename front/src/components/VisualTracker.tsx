"use client";

import React from 'react';
import { CheckCircle2, Plane, Package, Building2, CircleDot } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Dictionary, ShipmentData } from '@/types/shipment';
import { motion, Variants, useScroll, useTransform } from 'framer-motion';

interface VisualTrackerProps {
  shipment: ShipmentData;
  dict: Dictionary;
}

const steps = [
  { key: 'PENDING', icon: CircleDot, labelKey: 'pending', fallback: 'Dispatched' },
  { key: 'IN_TRANSIT', icon: Package, labelKey: 'inTransit', fallback: 'In Transit' },
  { key: 'DELIVERED', icon: Building2, labelKey: 'delivered', fallback: 'Arrived' },
];

// Contrail particles behind the plane
const CONTRAIL_COUNT = 4;
const contrailOffsets = Array.from({ length: CONTRAIL_COUNT }, (_, i) => (i + 1) * 0.06);

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

  // Map status to 3-step model: PENDING(0), IN_TRANSIT(1), DELIVERED(2)
  const progressIndex = status === 'CANCELED' ? -1
    : status === 'DELIVERED' ? 2
      : (status === 'IN_TRANSIT' || status === 'OUT_FOR_DELIVERY') ? 1
        : 0; // PENDING = step 0 (Dispatched)

  // Calculate dynamic progress across the full journey [0% - 100%]
  // 0% = First Node (Dispatched) → 50% = Second Node (In Transit) → 100% = Last Node (Arrived)
  const calculateJourneyProgress = () => {
    if (status === 'CANCELED') return 0;
    if (status === 'PENDING') return 0;
    if (status === 'DELIVERED') return 100;

    const now = Date.now();
    const transit = shipment.scheduledTransitTime ? new Date(shipment.scheduledTransitTime).getTime() : null;
    const outForDelivery = shipment.outfordeliveryTime ? new Date(shipment.outfordeliveryTime).getTime() : null;
    const arrival = shipment.expectedDeliveryTime ? new Date(shipment.expectedDeliveryTime).getTime() : null;

    if (status === 'IN_TRANSIT' && transit && outForDelivery) {
      // Scale from 50% to 75% (first half of In Transit → Arrived segment)
      const stepProgress = Math.min(0.95, Math.max(0.05, (now - transit) / (outForDelivery - transit)));
      return 50 + (stepProgress * 25);
    }

    if (status === 'OUT_FOR_DELIVERY' && outForDelivery && arrival) {
      // Scale from 75% to 95% (second half of In Transit → Arrived segment)
      const stepProgress = Math.min(0.95, Math.max(0.05, (now - outForDelivery) / (arrival - outForDelivery)));
      return 75 + (stepProgress * 20);
    }

    // Fallback: use status as rough marker
    if (status === 'IN_TRANSIT') return 60;
    if (status === 'OUT_FOR_DELIVERY') return 80;
    return 0;
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

          {/* Contrail Particles - trailing dots behind the plane */}
          {status !== 'CANCELED' && contrailOffsets.map((offset, i) => {
            const trailProgress = Math.max(0, normalizedP - offset);
            const trailLeft = 10 + trailProgress * 80;
            const trailTop = 50 - 15 * Math.sin(trailProgress * Math.PI * 2);
            return (
              <motion.div
                key={`contrail-${i}`}
                className="absolute z-20 pointer-events-none"
                animate={{
                  left: `${trailLeft}%`,
                  top: `${trailTop}%`,
                  x: '-50%',
                  y: '-50%',
                  opacity: journeyProgress > 3 ? (0.6 - i * 0.15) : 0,
                  scale: 1 - i * 0.15,
                }}
                transition={{ duration: 1.8 + i * 0.1, ease: "easeInOut" }}
              >
                <div
                  className="rounded-full bg-accent"
                  style={{ width: `${Math.max(3, 6 - i * 1.5)}px`, height: `${Math.max(3, 6 - i * 1.5)}px` }}
                />
              </motion.div>
            );
          })}

          {/* Mascot (Cargo Plane) - visible for all statuses except CANCELED */}
          {status !== 'CANCELED' && (
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
              {/* Glow halo behind plane */}
              <motion.div
                className="absolute inset-0 rounded-full bg-accent/20 blur-md"
                animate={{ scale: [1, 1.8, 1], opacity: [0.4, 0.1, 0.4] }}
                transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
                style={{ width: 24, height: 24, top: -4, left: -4 }}
              />
              <motion.div
                animate={{
                  y: [0, -3, 0, -1.5, 0],
                  rotate: [5, 10, 5, -2, 5]
                }}
                transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
              >
                <Plane className="w-3.5 h-3.5 sm:w-4 sm:h-4 md:w-5 md:h-5 fill-accent drop-shadow-[0_0_8px_rgba(var(--color-accent-rgb),0.7)]" />
              </motion.div>
            </motion.div>
          )}

          {/* Journey Waypoints — 3 stops */}
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
                {/* Radar ping on active node */}
                {isActive && (
                  <motion.div
                    className="absolute rounded-full border-2 border-accent/40"
                    style={{ width: '100%', height: '100%' }}
                    animate={{ scale: [1, 2.2], opacity: [0.6, 0] }}
                    transition={{ duration: 1.5, repeat: Infinity, ease: "easeOut" }}
                  />
                )}

                <motion.div
                  initial={false}
                  animate={isActive ? {
                    scale: [1.1, 1.18, 1.1],
                    backgroundColor: 'var(--color-accent)',
                    borderColor: 'var(--color-accent)',
                    boxShadow: [
                      '0 0 20px rgba(var(--color-accent-rgb), 0.3), 0 0 40px rgba(var(--color-accent-rgb), 0.1)',
                      '0 0 35px rgba(var(--color-accent-rgb), 0.5), 0 0 70px rgba(var(--color-accent-rgb), 0.2)',
                      '0 0 20px rgba(var(--color-accent-rgb), 0.3), 0 0 40px rgba(var(--color-accent-rgb), 0.1)',
                    ],
                  } : {
                    scale: 1,
                    backgroundColor: isCompleted ? 'var(--color-accent)' : 'var(--color-surface-muted)',
                    borderColor: isCompleted ? 'var(--color-accent)' : 'var(--color-border)',
                    boxShadow: isCompleted
                      ? '0 0 12px rgba(var(--color-accent-rgb), 0.15)'
                      : 'none',
                  }}
                  className={cn(
                    "w-9 h-9 sm:w-12 sm:h-12 md:w-16 md:h-16 rounded-2xl sm:rounded-[20px] flex items-center justify-center border-2",
                    isCompleted ? "text-white" : "text-text-muted border-dashed bg-surface-muted"
                  )}
                  transition={isActive
                    ? { duration: 2, repeat: Infinity, ease: "easeInOut" }
                    : { duration: 0.6, ease: "easeInOut" }
                  }
                >
                  {isCompleted && !isActive ? (
                    <motion.div
                      initial={{ scale: 0, rotate: -90 }}
                      animate={{ scale: 1, rotate: 0 }}
                      transition={{ type: "spring", stiffness: 400, damping: 15, delay: 0.2 }}
                    >
                      <CheckCircle2 className="w-4 h-4 sm:w-5 sm:h-5 md:w-7 md:h-7" />
                    </motion.div>
                  ) : (
                    <Icon className="w-4 h-4 sm:w-5 sm:h-5 md:w-7 md:h-7" />
                  )}
                </motion.div>

                <div className="absolute top-[100%] mt-2 sm:mt-3 flex flex-col items-center w-16 sm:w-20 md:w-24">
                  <p className={cn(
                    "text-[7px] sm:text-[9px] md:text-[11px] font-black uppercase tracking-wider sm:tracking-widest text-center leading-tight",
                    isCompleted ? "text-text-main" : "text-text-muted opacity-30"
                  )}>
                    {dict.admin?.[step.labelKey] || step.fallback}
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
