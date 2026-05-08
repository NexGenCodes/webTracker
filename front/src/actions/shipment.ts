'use server';

import { ShipmentData, ShipmentStatus } from '@/types/shipment';
import { logger } from '@/lib/logger';
import { vitals } from '@/lib/vitals';
import { cookies } from 'next/headers';
import { createClient } from '@/lib/supabase/server';

const BACKEND_URL = process.env.BACKEND_URL ?? '';

// ─── Helpers ────────────────────────────────────────────────────────────────

async function getJwt(): Promise<string | null> {
    const cookieStore = await cookies();
    return cookieStore.get('jwt')?.value ?? null;
}

const normalizeStatus = (s: string): string => {
    const upper = s.toUpperCase();
    if (upper === 'INTRANSIT') return 'IN_TRANSIT';
    if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
    if (upper === 'CANCELLED') return 'CANCELED';
    return upper;
};

async function goAPI(
    path: string,
    method: string,
    jwt: string,
    body?: unknown,
): Promise<{ success: boolean; data?: unknown; error?: string }> {
    try {
        const res = await fetch(`${BACKEND_URL}${path}`, {
            method,
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${jwt}`,
            },
            body: body !== undefined ? JSON.stringify(body) : undefined,
        });
        const json = await res.json().catch(() => ({})) as Record<string, unknown>;
        if (!res.ok) {
            return { success: false, error: (json.error as string) || `Server error ${res.status}` };
        }
        return { success: true, data: json };
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Network error';
        return { success: false, error: message };
    }
}

// ─── Queries ────────────────────────────────────────────────────────────────

/**
 * Public: Get tracking details (Supabase RPC)
 */
export async function getTrackingAction(trackingNumber: string): Promise<ShipmentData | null> {
    vitals.track('TRACKING_REQUESTED');
    if (!trackingNumber) return null;

    try {
        const supabase = await createClient();

        const { data: rawData, error } = await supabase
            .rpc('get_public_shipment', { p_tracking_id: trackingNumber })
            .single();
        
        if (error || !rawData) {
            if (error?.code === 'PGRST116') return null;
            throw new Error(`Supabase error: ${error?.message}`);
        }

        const data = rawData as Record<string, unknown>;
        const timelineStr = (val: unknown) => typeof val === 'string' ? val : '';
        const statusStr = typeof data.status === 'string' ? data.status.toLowerCase() : '';
        const scheduledTransit = timelineStr(data.scheduled_transit_time);
        const expectedDelivery = timelineStr(data.expected_delivery_time);

        const timeline = [
            {
                status: 'Order Placed',
                timestamp: timelineStr(data.created_at),
                description: `Shipment registered at ${timelineStr(data.origin) || 'origin'}`,
                is_completed: true
            },
            {
                status: 'In Transit',
                timestamp: scheduledTransit,
                description: 'Package has left the origin facility and is on its way',
                is_completed: ['intransit', 'outfordelivery', 'delivered'].includes(statusStr)
            },
            {
                status: 'Out for Delivery',
                timestamp: timelineStr(data.outfordelivery_time),
                description: 'Package is with our local agent for final delivery',
                is_completed: ['outfordelivery', 'delivered'].includes(statusStr)
            },
            {
                status: 'Delivered',
                timestamp: expectedDelivery,
                description: 'Package has been successfully delivered',
                is_completed: statusStr === 'delivered'
            }
        ];

        const redactName = (name: unknown): string => {
            if (typeof name !== 'string' || !name) return 'N/A';
            const parts = name.split(' ');
            if (parts[0].length <= 2) return parts[0] + '***';
            return parts[0].substring(0, 2) + '******';
        };

        return {
            id: timelineStr(data.tracking_id),
            trackingNumber: timelineStr(data.tracking_id),
            status: normalizeStatus(statusStr) as ShipmentStatus,
            senderName: redactName(data.sender_name),
            receiverName: redactName(data.recipient_name),
            receiverPhone: typeof data.recipient_phone === 'string' ? data.recipient_phone : null,
            receiverEmail: typeof data.recipient_email === 'string' ? data.recipient_email : null,
            receiverAddress: typeof data.recipient_address === 'string' ? data.recipient_address : null,
            receiverCountry: timelineStr(data.destination) || 'N/A',
            weight: typeof data.weight === 'number' ? data.weight : (typeof data.weight === 'string' ? parseFloat(data.weight) : 0),
            senderCountry: timelineStr(data.origin) || 'N/A',
            timeline: timeline,
            events: [],
            createdAt: timelineStr(data.created_at),
            scheduledTransitTime: scheduledTransit,
            outfordeliveryTime: timelineStr(data.outfordelivery_time),
            expectedDeliveryTime: expectedDelivery,
            isArchived: statusStr === 'delivered',
        };
    } catch (error) {
        logger.error(`Error fetching tracking`, error);
        return null;
    }
}

// ─── Mutations (Go API) ─────────────────────────────────────────────────────

export async function createShipmentAction(data: {
    sender_name: string; sender_phone?: string; origin: string;
    recipient_name: string; recipient_phone: string; recipient_email: string;
    recipient_address: string; destination: string;
    cargo_type?: string; weight: number; cost?: number;
}) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };

    return goAPI('/api/admin/shipments/', 'POST', jwt, {
        senderName: data.sender_name,
        senderPhone: data.sender_phone,
        senderCountry: data.origin,
        receiverName: data.recipient_name,
        receiverPhone: data.recipient_phone,
        receiverEmail: data.recipient_email,
        receiverAddress: data.recipient_address,
        receiverCountry: data.destination,
        cargoType: data.cargo_type,
        weight: data.weight,
        cost: data.cost,
    });
}

export async function editShipmentAction(
    trackingId: string,
    data: Record<string, unknown>,
) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI(`/api/admin/shipments/${trackingId}`, 'PUT', jwt, data);
}

export async function updateStatusAction(
    trackingId: string,
    status: string,
    destination: string,
) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI(`/api/admin/shipments/${trackingId}`, 'PATCH', jwt, { status, destination });
}

export async function deleteShipmentAction(trackingId: string) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI(`/api/admin/shipments/${trackingId}`, 'DELETE', jwt);
}

export async function bulkDeleteAction(ids: string[]) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI('/api/admin/shipments/bulk_delete', 'DELETE', jwt, { ids });
}

export async function bulkStatusAction(ids: string[], status: string) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI('/api/admin/shipments/bulk_status', 'PATCH', jwt, { ids, status });
}

export async function parseManifestAction(text: string) {
    const jwt = await getJwt();
    if (!jwt) return { success: false, error: 'Unauthorized' };
    return goAPI('/api/admin/shipments/parse', 'POST', jwt, { text });
}
