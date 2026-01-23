import { CreateShipmentDto, ShipmentData, ServiceResult } from '@/types/shipment';
import { logger } from '@/lib/logger';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_TOKEN = process.env.API_AUTH_TOKEN || '';

export class ShipmentService {
    /**
     * Create a new shipment via Go API
     */
    static async create(data: CreateShipmentDto): Promise<ServiceResult<{ trackingNumber: string }>> {
        try {
            const response = await fetch(`${API_URL}/api/shipments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                },
                body: JSON.stringify(data),
            });

            if (!response.ok) throw new Error(`API error: ${response.statusText}`);
            const result = await response.json();

            return { success: true, data: { trackingNumber: result.tracking_id } };
        } catch (error) {
            logger.error('[ShipmentService] Create error', error);
            return { success: false, error: 'API connection failed' };
        }
    }

    /**
     * Fetch tracking details from Go Backend API
     */
    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        try {
            const response = await fetch(`${API_URL}/api/track/${trackingNumber}`, {
                next: { revalidate: 0 }
            });

            if (!response.ok) {
                if (response.status === 404) return null;
                throw new Error(`API error: ${response.statusText}`);
            }

            const data = await response.json();

            // Map backend simple status to frontend typed status
            const normalizeStatus = (s: string): string => {
                const upper = s.toUpperCase();
                if (upper === 'INTRANSIT') return 'IN_TRANSIT';
                if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
                if (upper === 'CANCELLED') return 'CANCELED'; // Handle legacy
                return upper;
            };

            const shipment: ShipmentData = {
                id: data.tracking_id,
                trackingNumber: data.tracking_id,
                status: normalizeStatus(data.status) as any,
                senderName: data.sender_name || 'N/A',
                receiverName: data.recipient_name || 'N/A',
                receiverPhone: data.recipient_phone || null,
                receiverEmail: data.recipient_email || null,
                receiverAddress: data.destination || null,
                receiverCountry: data.destination || 'N/A',
                senderCountry: data.origin || 'N/A',
                timeline: data.timeline || [],
                events: [],
                isArchived: data.status === 'delivered',
            };
            return shipment;
        } catch (error) {
            logger.error('[ShipmentService] Fetch error', error);
            return null;
        }
    }

    /**
     * Admin: Update status via Go API
     */
    static async updateStatus(trackingNumber: string, status: string, location: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetch(`${API_URL}/api/shipments/${trackingNumber}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                },
                body: JSON.stringify({ status: status.toLowerCase(), destination: location }),
            });

            if (!response.ok) throw new Error(`API error: ${response.statusText}`);
            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Update error', error);
            return { success: false, error: 'API connection failed' };
        }
    }

    /**
     * Admin: Mark as delivered
     */
    static async markDelivered(trackingNumber: string): Promise<ServiceResult<void>> {
        return this.updateStatus(trackingNumber, 'delivered', 'Destination');
    }

    /**
     * Admin: Delete shipment via Go API
     */
    static async delete(trackingNumber: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetch(`${API_URL}/api/shipments/${trackingNumber}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                }
            });

            if (!response.ok) throw new Error(`API error: ${response.statusText}`);
            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Delete error', error);
            return { success: false, error: 'API connection failed' };
        }
    }

    /**
     * Admin: Dashboard data via Go API
     */
    static async getDashboardData(): Promise<ServiceResult<{ shipments: any[], stats: any }>> {
        try {
            const headers = { 'Authorization': `Bearer ${AUTH_TOKEN}` };

            // Fetch List
            const listRes = await fetch(`${API_URL}/api/shipments`, {
                headers,
                next: { revalidate: 0 }
            });
            if (!listRes.ok) throw new Error('Failed to fetch shipments');
            const apiShipments = await listRes.json();

            const normalizeStatus = (s: string): string => {
                const upper = s.toUpperCase();
                if (upper === 'INTRANSIT') return 'IN_TRANSIT';
                if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
                if (upper === 'CANCELLED') return 'CANCELED';
                return upper;
            };

            // Map to frontend expected keys (camelCase)
            const shipments = apiShipments.map((s: any) => ({
                id: s.tracking_id,
                trackingNumber: s.tracking_id,
                status: normalizeStatus(s.status),
                senderName: s.sender_name,
                receiverName: s.recipient_name,
                receiverPhone: s.recipient_phone,
                receiverEmail: s.recipient_email,
                receiverAddress: s.destination,
                senderCountry: s.origin,
                createdAt: s.created_at,
                isArchived: s.status === 'delivered',
            }));

            // Fetch Stats
            const statsRes = await fetch(`${API_URL}/api/stats`, {
                headers,
                next: { revalidate: 0 }
            });
            if (!statsRes.ok) throw new Error('Failed to fetch stats');
            const apiStats = await statsRes.json();

            const stats = {
                total: shipments.length,
                inTransit: shipments.filter((s: any) => s.status === 'IN_TRANSIT').length,
                outForDelivery: shipments.filter((s: any) => s.status === 'OUT_FOR_DELIVERY').length,
                delivered: shipments.filter((s: any) => s.status === 'DELIVERED').length,
                pending: shipments.filter((s: any) => s.status === 'PENDING').length,
                canceled: shipments.filter((s: any) => s.status === 'CANCELED').length,
            };

            return { success: true, data: { shipments, stats } };
        } catch (error) {
            logger.error('[ShipmentService] Dashboard error', error);
            return { success: false, error: 'API connection failed' };
        }
    }

    // Unused legacy methods
    /**
     * Admin: Bulk cleanup of delivered shipments
     */
    static async bulkDeleteDelivered(): Promise<ServiceResult<void>> {
        try {
            const response = await fetch(`${API_URL}/api/shipments/cleanup`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                }
            });

            if (!response.ok) throw new Error(`API error: ${response.statusText}`);
            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Bulk Delete error', error);
            return { success: false, error: 'API connection failed' };
        }
    }
    static async pruneStale(): Promise<void> { }
}
