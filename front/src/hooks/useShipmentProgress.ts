import { useState, useEffect } from 'react';
import { ShipmentData } from '@/types/shipment';

export function useShipmentProgress(shipment: ShipmentData | null): number {
  const [now, setNow] = useState<number | null>(null);

  const status = shipment?.status;
  const needsTimer = status === 'IN_TRANSIT' || status === 'OUT_FOR_DELIVERY';

  useEffect(() => {
    if (!needsTimer) return;

    const update = () => setNow(Date.now());
    const initialSync = setTimeout(update, 0);
    const timer = setInterval(update, 10000);
    return () => {
      clearTimeout(initialSync);
      clearInterval(timer);
    };
  }, [needsTimer]);

  if (!shipment) return 0;

  if (status === 'CANCELED') return 0;
  if (status === 'PENDING') return 0;
  if (status === 'DELIVERED') return 100;
  if (!now) return 0;

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
}
