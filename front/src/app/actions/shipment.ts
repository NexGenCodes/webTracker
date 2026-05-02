'use server';

import { ShipmentData } from '@/types/shipment';
import { ShipmentService } from '@/services/shipment.service';
import { logger } from '@/lib/logger';
import { vitals } from '@/lib/vitals';

/**
 * Public: Get tracking details
 */
export async function getTracking(trackingNumber: string): Promise<ShipmentData | null> {
    vitals.track('TRACKING_REQUESTED');
    try {
        return await ShipmentService.getByTracking(trackingNumber);
    } catch (error) {
        logger.error('Error fetching tracking', { trackingNumber, error });
        return null; // Return null instead of crashing, behaves like "not found"
    }
}

