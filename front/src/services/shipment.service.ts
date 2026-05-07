import { ShipmentData, ShipmentStatus } from '@/types/shipment';
import { logger } from '@/lib/logger';
import { createClient } from '@/lib/supabase/server';

const normalizeStatus = (s: string): string => {
    const upper = s.toUpperCase();
    if (upper === 'INTRANSIT') return 'IN_TRANSIT';
    if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
    if (upper === 'CANCELLED') return 'CANCELED';
    return upper;
};

export class ShipmentService {
    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        try {
            const supabase = await createClient();

            const { data: rawData, error } = await supabase
                .rpc('get_public_shipment', { p_tracking_id: trackingNumber })
                .single();
            const data = rawData as Record<string, unknown>;

            if (error || !data) {
                if (error?.code === 'PGRST116') return null;
                throw new Error(`Supabase error: ${error?.message}`);
            }

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

            const shipment: ShipmentData = {
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
            return shipment;
        } catch (error) {
            logger.error(`[ShipmentService] Fetch tracking`, error);
            return null;
        }
    }
    static async create(companyId: string, data: Record<string, unknown>): Promise<{ success: boolean; error?: string }> {
        try {
            const supabase = await createClient();
            const { error } = await supabase
                .from('shipment')
                .insert([{
                    ...data,
                    company_id: companyId,
                    status: 'pending'
                }]);

            if (error) throw error;
            return { success: true };
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            logger.error(`[ShipmentService] Create shipment`, error);
            return { success: false, error: message };
        }
    }

    static async update(id: string, companyId: string, data: Record<string, unknown>): Promise<{ success: boolean; error?: string }> {
        try {
            const supabase = await createClient();
            const { error } = await supabase
                .from('shipment')
                .update(data)
                .eq('id', id)
                .eq('company_id', companyId);

            if (error) throw error;
            return { success: true };
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            logger.error(`[ShipmentService] Update shipment`, error);
            return { success: false, error: message };
        }
    }

    static async delete(id: string, companyId: string): Promise<{ success: boolean; error?: string }> {
        try {
            const supabase = await createClient();
            const { error } = await supabase
                .from('shipment')
                .delete()
                .eq('id', id)
                .eq('company_id', companyId);

            if (error) throw error;
            return { success: true };
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            logger.error(`[ShipmentService] Delete shipment`, error);
            return { success: false, error: message };
        }
    }

    static async parse(text: string, jwt: string): Promise<{ success: boolean; data?: unknown; error?: string }> {
        try {
            const response = await fetch(`${process.env.BACKEND_URL}/api/admin/shipments/parse`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${jwt}`
                },
                body: JSON.stringify({ text })
            });

            if (!response.ok) {
                const errData = await response.json().catch(() => ({}));
                throw new Error(errData.error || `Server responded with ${response.status}`);
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            logger.error(`[ShipmentService] Parse manifest`, error);
            return { success: false, error: message };
        }
    }
}
