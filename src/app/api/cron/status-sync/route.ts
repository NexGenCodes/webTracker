import { NextResponse } from 'next/server';
import { prisma } from '@/lib/prisma';

export async function GET(request: Request) {
    const authHeader = request.headers.get('authorization');

    // Secure the endpoint with CRON_SECRET
    const isVercelCron = authHeader === `Bearer ${process.env.CRON_SECRET}`;
    const isExternalCron = authHeader === `Bearer ${process.env.EXTERNAL_CRON_SECRET}`;

    if (process.env.NODE_ENV === 'production' && !isVercelCron && !isExternalCron) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        // Find all PENDING shipments
        const pendingShipments = await prisma.shipment.findMany({
            where: {
                status: 'PENDING'
            }
        });

        if (pendingShipments.length === 0) {
            return NextResponse.json({
                success: true,
                message: 'No pending shipments to sync',
                updated: 0
            });
        }

        console.log(`[Status Sync] Processing ${pendingShipments.length} pending shipments`);

        let updatedCount = 0;
        let notificationsSent = 0;

        for (const shipment of pendingShipments) {
            // Update status to IN_TRANSIT
            await prisma.$transaction([
                prisma.shipment.update({
                    where: { id: shipment.id },
                    data: {
                        status: 'IN_TRANSIT',
                        lastTransitionAt: new Date()
                    }
                }),
                prisma.event.create({
                    data: {
                        shipmentId: shipment.id,
                        status: 'IN_TRANSIT',
                        location: shipment.senderCountry || 'Origin Center',
                        notes: 'Automatic status sync: Package in transit'
                    }
                })
            ]);

            updatedCount++;

            // Send WhatsApp notification if this was a WhatsApp-originated shipment
            // @ts-ignore - whatsappMessageId and whatsappFrom exist but may need Prisma regeneration
            if (shipment.whatsappMessageId && shipment.whatsappFrom) {
                try {
                    const message = `📦 *Status Update*\n\nYour package is now IN TRANSIT!\n\nTracking ID: *${shipment.trackingNumber}*\nStatus: IN_TRANSIT\n\nYou'll receive another update when it's out for delivery.`;

                    const response = await fetch(
                        `https://graph.facebook.com/v17.0/${process.env.WHATSAPP_PHONE_NUMBER_ID}/messages`,
                        {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                                Authorization: `Bearer ${process.env.WHATSAPP_TOKEN}`,
                            },
                            body: JSON.stringify({
                                messaging_product: 'whatsapp',
                                recipient_type: 'individual',
                                // @ts-ignore
                                to: shipment.whatsappFrom,
                                context: {
                                    // @ts-ignore
                                    message_id: shipment.whatsappMessageId
                                },
                                type: 'text',
                                text: { body: message },
                            }),
                        }
                    );

                    if (response.ok) {
                        notificationsSent++;
                        console.log(`[Status Sync] Notification sent for ${shipment.trackingNumber}`);
                    } else {
                        console.error(`[Status Sync] Failed to send notification for ${shipment.trackingNumber}`);
                    }
                } catch (error) {
                    console.error(`[Status Sync] WhatsApp error for ${shipment.trackingNumber}:`, error);
                }
            }
        }

        return NextResponse.json({
            success: true,
            message: 'Status sync completed',
            updated: updatedCount,
            notificationsSent,
            timestamp: new Date().toISOString()
        });

    } catch (error) {
        console.error('[Status Sync] Error:', error);
        return NextResponse.json({
            error: 'Status sync failed',
            details: error instanceof Error ? error.message : 'Unknown error'
        }, { status: 500 });
    }
}
