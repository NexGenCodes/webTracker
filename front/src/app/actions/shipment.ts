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


/**
 * Admin: Create shipment
 */
export async function createShipmentAction(companyId: string, data: Record<string, unknown>) {
    return await ShipmentService.create(companyId, data);
}

/**
 * Admin: Update shipment
 */
export async function updateShipmentAction(id: string, companyId: string, data: Record<string, unknown>) {
    return await ShipmentService.update(id, companyId, data);
}

/**
 * Admin: Delete shipment
 */
export async function deleteShipmentAction(id: string, companyId: string) {
    return await ShipmentService.delete(id, companyId);
}

import { cookies } from 'next/headers';

/**
 * Admin: Parse Manifest with AI (Backend calls Gemini/Regex)
 */
export async function parseManifestAction(text: string) {
    try {
        const cookieStore = await cookies();
        const jwt = cookieStore.get('jwt')?.value;
        if (!jwt) return { success: false, error: 'Unauthorized' };

        return await ShipmentService.parse(text, jwt);
    } catch (error: unknown) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        return { success: false, error: message };
    }
}
