import { prisma } from '@/lib/prisma';
import { CreateShipmentDto, ShipmentData, ServiceResult } from '@/types/shipment';
import crypto from 'crypto';
import { COUNTRY_COORDS, TRACKING_PREFIX } from "@/lib/constants";
import { logger } from '@/lib/logger';

export class ShipmentService {
    /**
     * Helper: Normalize coordinates based on country name
     */
    private static getCoordinates(country: string | null): [number, number] {
        if (!country) return [20.0, 0.0];
        const normalized = country.trim().toLowerCase();

        for (const [key, coords] of Object.entries(COUNTRY_COORDS)) {
            if (key.toLowerCase() === normalized) return coords;
        }

        for (const [key, coords] of Object.entries(COUNTRY_COORDS)) {
            if (key.toLowerCase().includes(normalized) || normalized.includes(key.toLowerCase())) return coords;
        }

        return [20.0, 0.0];
    }

    /**
     * Create a new shipment
     */
    static async create(data: CreateShipmentDto): Promise<ServiceResult<{ trackingNumber: string }>> {
        try {
            // Aligned with Go backend standard: Exactly 9 characters total
            const randomLen = Math.max(9 - TRACKING_PREFIX.length - 1, 3);
            const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
            let randomPart = "";
            for (let i = 0; i < randomLen; i++) {
                randomPart += charset.charAt(Math.floor(Math.random() * charset.length));
            }
            const trackingNumber = `${TRACKING_PREFIX}-${randomPart}`;

            const shipment = await prisma.shipment.create({
                data: {
                    trackingNumber,
                    status: 'IN_TRANSIT',
                    senderName: data.senderName,
                    senderCountry: data.senderCountry,
                    receiverName: data.receiverName,
                    receiverAddress: data.receiverAddress,
                    receiverCountry: data.receiverCountry,
                    receiverPhone: data.receiverPhone,
                    events: {
                        create: {
                            status: 'IN_TRANSIT',
                            location: data.senderCountry || 'Origin',
                            notes: 'Shipment created'
                        }
                    }
                }
            });

            return { success: true, data: { trackingNumber: shipment.trackingNumber } };
        } catch (error) {
            logger.error('[ShipmentService] Create error', error);
            return { success: false, error: 'Database error' };
        }
    }

    /**
     * Fetch tracking details with self-healing logic
     */
    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        const shipment = await prisma.shipment.findUnique({
            where: { trackingNumber },
            include: { events: { orderBy: { timestamp: 'desc' } } }
        });

        if (!shipment) return null;

        // Self-Healing: Auto-transition PENDING shipments older than 1 hour
        if (shipment.status === 'PENDING') {
            const oneHourAgo = new Date();
            oneHourAgo.setHours(oneHourAgo.getHours() - 1);

            if (new Date(shipment.createdAt) < oneHourAgo) {
                await prisma.$transaction([
                    prisma.shipment.update({
                        where: { id: shipment.id },
                        data: { status: 'IN_TRANSIT' }
                    }),
                    prisma.event.create({
                        data: {
                            shipmentId: shipment.id,
                            status: 'IN_TRANSIT',
                            location: shipment.senderCountry || 'Origin Center',
                            notes: 'Automatic transition: Intake complete'
                        }
                    })
                ]);
                shipment.status = 'IN_TRANSIT';
                // Note: The caller (action) should handle the WhatsApp notification if needed
            }
        }

        if (shipment.isArchived) {
            return {
                trackingNumber: shipment.trackingNumber,
                status: 'DELIVERED',
                isArchived: true,
                id: shipment.id,
                events: []
            } as any;
        }

        const createdAt = new Date(shipment.createdAt);
        const estimatedDelivery = new Date(createdAt);
        estimatedDelivery.setDate(createdAt.getDate() + 1);
        estimatedDelivery.setHours(8, 0, 0, 0);

        const originCountry = shipment.senderCountry || shipment.events[shipment.events.length - 1]?.location || '';
        const destinationCountry = shipment.receiverCountry || '';

        return {
            ...shipment,
            originCoords: this.getCoordinates(originCountry),
            destinationCoords: this.getCoordinates(destinationCountry),
            estimatedDelivery: estimatedDelivery.toISOString(),
        } as unknown as ShipmentData;
    }

    /**
     * Standard status updates
     */
    static async updateStatus(trackingNumber: string, status: string, location: string): Promise<ServiceResult<void>> {
        try {
            const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } });
            if (!shipment) return { success: false, error: 'Not found' };

            await prisma.$transaction([
                prisma.event.create({
                    data: {
                        shipmentId: shipment.id,
                        status,
                        location,
                        notes: `Status updated to ${status}`
                    }
                }),
                prisma.shipment.update({
                    where: { id: shipment.id },
                    data: { status }
                })
            ]);

            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Status update error', error);
            return { success: false, error: 'Database error' };
        }
    }

    /**
     * Mark as delivered and archive sensitive data
     */
    static async markDelivered(trackingNumber: string): Promise<ServiceResult<void>> {
        try {
            const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } });
            if (!shipment) return { success: false, error: 'Not found' };

            await prisma.$transaction([
                prisma.event.create({
                    data: {
                        shipmentId: shipment.id,
                        status: 'DELIVERED',
                        location: 'Destination',
                        notes: 'Delivered to recipient'
                    }
                }),
                prisma.shipment.update({
                    where: { id: shipment.id },
                    data: {
                        status: 'DELIVERED',
                        isArchived: true,
                        senderName: null,
                        senderCountry: null,
                        receiverName: null,
                        receiverEmail: null,
                        receiverAddress: null,
                        receiverCountry: null,
                        receiverPhone: null,
                    }
                })
            ]);

            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Delivery error', error);
            return { success: false, error: 'Database error' };
        }
    }

    /**
     * Admin: Delete single shipment
     */
    static async delete(trackingNumber: string): Promise<ServiceResult<void>> {
        try {
            const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } });
            if (!shipment) return { success: false, error: 'Shipment not found' };

            await prisma.event.deleteMany({ where: { shipmentId: shipment.id } });
            await prisma.shipment.delete({ where: { id: shipment.id } });

            return { success: true };
        } catch (error) {
            logger.error('[ShipmentService] Delete error', error);
            return { success: false, error: 'Database error' };
        }
    }

    /**
     * Admin: Bulk delete delivered
     */
    static async bulkDeleteDelivered(): Promise<ServiceResult<void>> {
        try {
            const archived = await prisma.shipment.findMany({
                where: { isArchived: true },
                select: { id: true }
            });
            const ids = archived.map(s => s.id);

            await prisma.event.deleteMany({ where: { shipmentId: { in: ids } } });
            const result = await prisma.shipment.deleteMany({ where: { isArchived: true } });

            return { success: true, count: result.count };
        } catch (error) {
            logger.error('[ShipmentService] Bulk delete error', error);
            return { success: false, error: 'Database error' };
        }
    }

    /**
     * Maintenance: Prune stale data
     */
    static async pruneStale(): Promise<void> {
        const oneWeekAgo = new Date();
        oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
        try {
            const oldOnes = await prisma.shipment.findMany({
                where: { createdAt: { lt: oneWeekAgo } },
                select: { id: true }
            });
            if (oldOnes.length === 0) return;
            const ids = oldOnes.map(s => s.id);
            await prisma.event.deleteMany({ where: { shipmentId: { in: ids } } });
            await prisma.shipment.deleteMany({ where: { id: { in: ids } } });
            logger.info(`Pruned ${ids.length} stale manifests.`);
        } catch (error) {
            logger.error('[ShipmentService] Pruning error', error);
        }
    }

    /**
     * Dashboard: Fetch stats and list
     */
    static async getDashboardData(): Promise<ServiceResult<{ shipments: any[], stats: any }>> {
        try {
            const shipments = await prisma.shipment.findMany({
                include: { events: { orderBy: { timestamp: 'desc' }, take: 1 } },
                orderBy: { createdAt: 'desc' }
            });

            const stats = {
                total: shipments.length,
                inTransit: shipments.filter((s) => s.status === 'IN_TRANSIT').length,
                delivered: shipments.filter((s) => s.isArchived).length,
                pending: shipments.filter((s) => s.status === 'PENDING').length,
                canceled: shipments.filter((s) => s.status === 'CANCELED').length,
            };

            return { success: true, data: { shipments, stats } };
        } catch (error) {
            logger.error('[ShipmentService] Dashboard error', error);
            return { success: false, error: 'Database error' };
        }
    }
}
