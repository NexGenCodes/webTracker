'use server'

import { prisma } from '@/lib/prisma'
import { CreateShipmentDto } from '@/types/shipment'
import { revalidatePath } from 'next/cache'
import { ShipmentData } from '@/types/shipment'
import crypto from 'crypto'
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import { COUNTRY_COORDS } from "@/lib/constants";

// Helper: Send WhatsApp notification for status changes
async function sendWhatsAppNotification(
    whatsappMessageId: string | null,
    from: string,
    trackingNumber: string,
    status: string
) {
    if (!whatsappMessageId || !from) return;

    const statusMessages: Record<string, string> = {
        'IN_TRANSIT': `📦 *Status Update*\n\nYour package is now in transit!\n\nTracking ID: *${trackingNumber}*\nStatus: IN_TRANSIT\n\nYou'll receive another update when it's out for delivery.`,
        'OUT_FOR_DELIVERY': `🚚 *Out for Delivery*\n\nYour package is on its way to you!\n\nTracking ID: *${trackingNumber}*\nExpected delivery: Today`,
        'DELIVERED': `✅ *Package Delivered*\n\nYour package has been successfully delivered!\n\nTracking ID: *${trackingNumber}*\n\nThank you for using our service!`
    };

    const message = statusMessages[status];
    if (!message) return;

    try {
        const response = await fetch(`https://graph.facebook.com/v17.0/${process.env.WHATSAPP_PHONE_NUMBER_ID}/messages`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${process.env.WHATSAPP_TOKEN}`,
            },
            body: JSON.stringify({
                messaging_product: "whatsapp",
                recipient_type: "individual",
                to: from,
                context: { message_id: whatsappMessageId },
                type: "text",
                text: { body: message },
            }),
        });

        if (!response.ok) {
            throw new Error(`WhatsApp API error: ${response.status}`);
        }

        console.log(`[WhatsApp] Notification sent for ${trackingNumber}: ${status}`);
    } catch (error) {
        console.error(`[WhatsApp] Failed to send notification:`, error);

        // Queue the notification for retry
        try {
            await prisma.notificationQueue.create({
                data: {
                    trackingNumber,
                    whatsappMessageId,
                    whatsappFrom: from,
                    status,
                    message,
                    retryCount: 0
                }
            });
            console.log(`[WhatsApp] Notification queued for retry: ${trackingNumber}`);
        } catch (queueError) {
            console.error(`[WhatsApp] Failed to queue notification:`, queueError);
        }
    }
}

export async function createShipment(data: CreateShipmentDto) {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    try {
        const hashInput = `${data.receiverName}${data.receiverPhone}${data.receiverCountry}${data.senderName}`;
        const hash = crypto.createHash('shake256', { outputLength: 5 })
            .update(hashInput)
            .digest('hex');

        const trackingNumber = `AWB-${hash}`;

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
                        location: data.senderCountry,
                        notes: 'Shipment created'
                    }
                }
            }
        })

        return { success: true, trackingNumber: shipment.trackingNumber }
    } catch (error) {
        console.error('Failed to create shipment:', error)
        return { success: false, error: 'Database error' }
    }
}

const getCoordinates = (country: string | null): [number, number] => {
    if (!country) return [20.0, 0.0];

    // Normalize: trim and lowercase for fuzzy matching
    const normalized = country.trim().toLowerCase();

    // Try exact match first (case-insensitive)
    for (const [key, coords] of Object.entries(COUNTRY_COORDS)) {
        if (key.toLowerCase() === normalized) {
            return coords;
        }
    }

    // Fallback: partial match (e.g., "usa" matches "United States")
    for (const [key, coords] of Object.entries(COUNTRY_COORDS)) {
        if (key.toLowerCase().includes(normalized) || normalized.includes(key.toLowerCase())) {
            return coords;
        }
    }

    // Default fallback
    return [20.0, 0.0];
};

export async function getTracking(trackingNumber: string): Promise<ShipmentData | null> {
    if (!trackingNumber) return null

    const shipment = await prisma.shipment.findUnique({
        where: { trackingNumber },
        include: { events: { orderBy: { timestamp: 'desc' } } }
    })

    if (!shipment) return null

    // Self-Healing: Auto-transition PENDING shipments older than 1 hour
    if (shipment.status === 'PENDING') {
        const oneHourAgo = new Date();
        oneHourAgo.setHours(oneHourAgo.getHours() - 1);

        if (new Date(shipment.createdAt) < oneHourAgo) {
            // Perform on-the-fly transition to IN_TRANSIT
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

            // Update local object to reflect the change
            shipment.status = 'IN_TRANSIT';

            // Send WhatsApp notification if this was a WhatsApp-originated shipment
            // @ts-ignore - whatsappMessageId and whatsappFrom exist but Prisma client needs regeneration
            if (shipment.whatsappMessageId && shipment.whatsappFrom) {
                await sendWhatsAppNotification(
                    // @ts-ignore
                    shipment.whatsappMessageId,
                    // @ts-ignore
                    shipment.whatsappFrom,
                    shipment.trackingNumber,
                    'IN_TRANSIT'
                );
            }
        }
    }

    if (shipment.isArchived) {
        return {
            trackingNumber: shipment.trackingNumber,
            status: 'DELIVERED',
            isArchived: true,
            senderName: null,
            receiverName: null,
            receiverAddress: null,
            receiverCountry: null,
            receiverPhone: null,
            id: shipment.id,
            events: []
        } as ShipmentData;
    }

    // Calculate next-day delivery (8am-10am destination time)
    const createdAt = new Date(shipment.createdAt);
    const estimatedDelivery = new Date(createdAt);
    estimatedDelivery.setDate(createdAt.getDate() + 1); // Next day
    estimatedDelivery.setHours(8, 0, 0, 0); // 8am destination time

    // Get origin from shipment data or first event or default
    const originCountry = shipment.senderCountry || shipment.events[shipment.events.length - 1]?.location || '';
    const destinationCountry = shipment.receiverCountry || '';

    return {
        ...shipment,
        originCoords: getCoordinates(originCountry),
        destinationCoords: getCoordinates(destinationCountry),
        senderCountry: originCountry,
        estimatedDelivery: estimatedDelivery.toISOString(),
    } as unknown as ShipmentData;
}

export async function markAsDelivered(trackingNumber: string) {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } })
    if (!shipment) return { success: false, error: 'Not found' }

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
    ])

    revalidatePath('/')
    return { success: true }
}

export async function cancelShipment(trackingNumber: string) {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } })
    if (!shipment) return { success: false, error: 'Not found' }

    // Only allow cancellation if not delivered (isArchived)
    if (shipment.isArchived) {
        return { success: false, error: 'Cannot cancel delivered shipment' }
    }

    await prisma.$transaction([
        prisma.event.create({
            data: {
                shipmentId: shipment.id,
                status: 'CANCELED',
                location: 'Admin Center',
                notes: 'Shipment canceled by admin'
            }
        }),
        prisma.shipment.update({
            where: { id: shipment.id },
            data: {
                status: 'CANCELED'
            }
        })
    ])

    revalidatePath('/')
    revalidatePath('/admin')
    return { success: true }
}

// New admin actions
export async function deleteShipment(trackingNumber: string) {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    try {
        const shipment = await prisma.shipment.findUnique({
            where: { trackingNumber },
            include: { events: true }
        })

        if (!shipment) return { success: false, error: 'Shipment not found' }

        // Delete events first (cascade)
        await prisma.event.deleteMany({
            where: { shipmentId: shipment.id }
        })

        // Delete shipment
        await prisma.shipment.delete({
            where: { id: shipment.id }
        })

        revalidatePath('/admin')
        return { success: true }
    } catch (error) {
        console.error('Failed to delete shipment:', error)
        return { success: false, error: 'Database error' }
    }
}

export async function bulkDeleteDelivered() {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    try {
        const deliveredShipments = await prisma.shipment.findMany({
            where: { isArchived: true },
            select: { id: true }
        })

        const shipmentIds = deliveredShipments.map(s => s.id)

        // Delete events first
        await prisma.event.deleteMany({
            where: { shipmentId: { in: shipmentIds } }
        })

        // Delete shipments
        const result = await prisma.shipment.deleteMany({
            where: { isArchived: true }
        })

        revalidatePath('/admin')
        return { success: true, count: result.count }
    } catch (error) {
        console.error('Failed to bulk delete:', error)
        return { success: false, error: 'Database error' }
    }
}

export async function pruneOldShipments() {
    const oneWeekAgo = new Date();
    oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);

    try {
        // Find ALL shipments older than 7 days, regardless of status
        const oldShipments = await prisma.shipment.findMany({
            where: {
                createdAt: { lt: oneWeekAgo }
            },
            select: { id: true }
        });

        if (oldShipments.length === 0) {
            return;
        }

        const ids = oldShipments.map(s => s.id);

        // Delete associated events first
        await prisma.event.deleteMany({
            where: { shipmentId: { in: ids } }
        });

        // Delete shipments
        await prisma.shipment.deleteMany({
            where: { id: { in: ids } }
        });

        console.log(`[Uplink] Maintenance complete: Pruned ${ids.length} stale manifests.`);
    } catch (error) {
        console.error('Pruning error:', error);
    }
}

export async function autoTransitionShipments() {
    const oneHourAgo = new Date();
    oneHourAgo.setHours(oneHourAgo.getHours() - 1);

    try {
        // Find Pending shipments older than 1 hour
        const pendingShipments = await prisma.shipment.findMany({
            where: {
                status: 'PENDING',
                createdAt: { lt: oneHourAgo }
            }
        });

        if (pendingShipments.length === 0) return;

        console.log(`[Uplink] Transitioning ${pendingShipments.length} manifests to IN_TRANSIT`);

        // Update each shipment and create an event
        // Using a transaction for efficiency and consistency
        await prisma.$transaction(
            pendingShipments.map(s => [
                prisma.shipment.update({
                    where: { id: s.id },
                    data: { status: 'IN_TRANSIT' }
                }),
                prisma.event.create({
                    data: {
                        shipmentId: s.id,
                        status: 'IN_TRANSIT',
                        location: s.senderCountry || 'Origin Center',
                        notes: 'Automatic transition: Intake complete'
                    }
                })
            ]).flat()
        );

        // Send WhatsApp notifications for WhatsApp-originated shipments
        for (const shipment of pendingShipments) {
            // @ts-ignore - whatsappMessageId and whatsappFrom exist but Prisma client needs regeneration
            if (shipment.whatsappMessageId && shipment.whatsappFrom) {
                await sendWhatsAppNotification(
                    // @ts-ignore
                    shipment.whatsappMessageId,
                    // @ts-ignore
                    shipment.whatsappFrom,
                    shipment.trackingNumber,
                    'IN_TRANSIT'
                );
            }
        }

    } catch (error) {
        console.error('Auto-transition error:', error);
    }
}

export async function processNotificationQueue() {
    const fiveMinutesAgo = new Date();
    fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);

    try {
        // Find notifications that need retry
        // @ts-ignore - notificationQueue exists but Prisma client needs regeneration
        const pendingNotifications = await prisma.notificationQueue.findMany({
            where: {
                retryCount: { lt: 3 },
                lastAttempt: { lt: fiveMinutesAgo }
            }
        });

        if (pendingNotifications.length === 0) {
            console.log('[WhatsApp] No notifications to retry');
            return;
        }

        console.log(`[WhatsApp] Processing ${pendingNotifications.length} queued notifications`);

        for (const notification of pendingNotifications) {
            try {
                const response = await fetch(`https://graph.facebook.com/v17.0/${process.env.WHATSAPP_PHONE_NUMBER_ID}/messages`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${process.env.WHATSAPP_TOKEN}`,
                    },
                    body: JSON.stringify({
                        messaging_product: "whatsapp",
                        recipient_type: "individual",
                        to: notification.whatsappFrom,
                        context: { message_id: notification.whatsappMessageId },
                        type: "text",
                        text: { body: notification.message },
                    }),
                });

                if (response.ok) {
                    // Success - delete from queue
                    // @ts-ignore
                    await prisma.notificationQueue.delete({
                        where: { id: notification.id }
                    });
                    console.log(`[WhatsApp] Retry successful for ${notification.trackingNumber}`);
                } else {
                    // Failed - increment retry count
                    // @ts-ignore
                    await prisma.notificationQueue.update({
                        where: { id: notification.id },
                        data: {
                            retryCount: notification.retryCount + 1,
                            lastAttempt: new Date()
                        }
                    });
                    console.log(`[WhatsApp] Retry ${notification.retryCount + 1}/3 failed for ${notification.trackingNumber}`);
                }
            } catch (error) {
                console.error(`[WhatsApp] Error retrying notification for ${notification.trackingNumber}:`, error);
                // Increment retry count even on error
                // @ts-ignore
                await prisma.notificationQueue.update({
                    where: { id: notification.id },
                    data: {
                        retryCount: notification.retryCount + 1,
                        lastAttempt: new Date()
                    }
                });
            }
        }

    } catch (error) {
        console.error('[WhatsApp] Notification queue processing error:', error);
    }
}

// Keep the old name as an alias for compatibility if needed, but I'll update the cron route too
export const pruneDeliveredShipments = pruneOldShipments;

export async function getAdminDashboardData() {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized', shipments: [], stats: null };

    // Pruning is now handled by a scheduled Vercel Cron Job

    try {
        const shipments = await prisma.shipment.findMany({
            include: {
                events: {
                    orderBy: { timestamp: 'desc' },
                    take: 1
                }
            },
            orderBy: { createdAt: 'desc' }
        })

        const total = shipments.length;
        const inTransit = shipments.filter((s) => s.status === 'IN_TRANSIT').length;
        const delivered = shipments.filter((s) => s.isArchived).length;
        const pending = shipments.filter((s) => s.status === 'PENDING').length;
        const canceled = shipments.filter((s) => s.status === 'CANCELED').length;

        return {
            success: true,
            shipments,
            stats: { total, inTransit, delivered, pending, canceled }
        }
    } catch (error) {
        console.error('Failed to get dashboard data:', error)
        return { success: false, error: 'Database error', shipments: [], stats: null }
    }
}

export async function getAllShipments() {
    return getAdminDashboardData();
}

export async function updateShipmentStatus(trackingNumber: string, status: 'PENDING' | 'IN_TRANSIT' | 'OUT_FOR_DELIVERY' | 'DELIVERED', location: string) {
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    try {
        const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } })
        if (!shipment) return { success: false, error: 'Not found' }

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
        ])

        revalidatePath('/')
        revalidatePath('/admin')
        return { success: true }
    } catch (error) {
        console.error('Failed to update status:', error)
        return { success: false, error: 'Database error' }
    }
}
