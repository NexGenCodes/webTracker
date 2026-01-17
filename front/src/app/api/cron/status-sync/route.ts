import { NextResponse } from 'next/server';
import { prisma } from '@/lib/prisma';
import { NotificationService } from '@/services/notification.service';
import { logger } from '@/lib/logger';

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

        logger.info(`[Status Sync] Processing ${pendingShipments.length} pending shipments`);

        let updatedCount = 0;

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

            // Use NotificationService for WhatsApp alerts
            // @ts-ignore
            if (shipment.whatsappMessageId && shipment.whatsappFrom) {
                await NotificationService.sendWhatsApp(
                    // @ts-ignore
                    shipment.whatsappMessageId,
                    // @ts-ignore
                    shipment.whatsappFrom,
                    shipment.trackingNumber,
                    'IN_TRANSIT'
                );
            }
        }

        return NextResponse.json({
            success: true,
            message: 'Status sync completed',
            updated: updatedCount,
            timestamp: new Date().toISOString()
        });

    } catch (error) {
        logger.error('[Status Sync] Error', error);
        return NextResponse.json({
            error: 'Status sync failed',
            details: error instanceof Error ? error.message : 'Unknown error'
        }, { status: 500 });
    }
}
