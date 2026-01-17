'use server';

import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import { revalidatePath } from 'next/cache';
import { CreateShipmentDto, ShipmentData, ServiceResult } from '@/types/shipment';
import { ShipmentService } from '@/services/shipment.service';
import { logger } from '@/lib/logger';
import { vitals } from '@/lib/vitals';

/**
 * Common Authorization wrapper
 */
async function isAdmin() {
    const session = await getServerSession(authOptions);
    return !!session;
}

/**
 * Create a new shipment from Admin Portal
 */
export async function createShipment(data: CreateShipmentDto): Promise<ServiceResult<{ trackingNumber: string }>> {
    vitals.track('SHIPMENT_CREATED');
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };

    logger.info('Creating shipment', { data });
    const result = await ShipmentService.create(data);
    if (result.success) revalidatePath('/admin');
    return result;
}

/**
 * Public: Get tracking details
 */
export async function getTracking(trackingNumber: string): Promise<ShipmentData | null> {
    vitals.track('TRACKING_REQUESTED');
    return await ShipmentService.getByTracking(trackingNumber);
}

/**
 * Admin: Update status manually
 */
export async function updateShipmentStatus(
    trackingNumber: string,
    status: 'PENDING' | 'IN_TRANSIT' | 'OUT_FOR_DELIVERY' | 'DELIVERED',
    location: string
): Promise<ServiceResult<void>> {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };

    logger.info('Updating shipment status', { trackingNumber, status, location });

    const result = await ShipmentService.updateStatus(trackingNumber, status, location);
    if (result.success) {
        revalidatePath('/');
        revalidatePath('/admin');
    }
    return result;
}

/**
 * Admin: Quick Deliver action
 */
export async function markAsDelivered(trackingNumber: string) {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };

    const result = await ShipmentService.markDelivered(trackingNumber);
    if (result.success) {
        revalidatePath('/');
        revalidatePath('/admin');
    }
    return result;
}

/**
 * Admin: Delete shipment
 */
export async function deleteShipment(trackingNumber: string) {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };

    const result = await ShipmentService.delete(trackingNumber);
    if (result.success) revalidatePath('/admin');
    return result;
}

/**
 * Admin: Bulk cleanup
 */
export async function bulkDeleteDelivered() {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };

    const result = await ShipmentService.bulkDeleteDelivered();
    if (result.success) revalidatePath('/admin');
    return result;
}

/**
 * Admin: Dashboard Stats
 */
export async function getAdminDashboardData() {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized', shipments: [], stats: null };
    return await ShipmentService.getDashboardData();
}

/**
 * Cron: Maintenance tasks
 */
export async function pruneOldShipments() {
    await ShipmentService.pruneStale();
}
