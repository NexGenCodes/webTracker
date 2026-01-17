import { prisma } from '@/lib/prisma';
import { logger } from '@/lib/logger';

export class NotificationService {
    /**
     * Send WhatsApp notification with auto-retry queueing
     */
    static async sendWhatsApp(
        whatsappMessageId: string | null,
        from: string,
        trackingNumber: string,
        status: string
    ): Promise<void> {
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

            if (!response.ok) throw new Error(`WhatsApp API error: ${response.status}`);
            logger.info('WhatsApp notification sent', { trackingNumber, status });
        } catch (error) {
            logger.error('WhatsApp notification failed, queueing', error);
            try {
                // @ts-ignore
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
            } catch (queueError) {
                logger.error('WhatsApp queueing failed', queueError);
            }
        }
    }

    /**
     * Process retries for failed notifications
     */
    static async processRetries(): Promise<void> {
        const fiveMinutesAgo = new Date();
        fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);

        try {
            // @ts-ignore
            const pending = await prisma.notificationQueue.findMany({
                where: { retryCount: { lt: 3 }, lastAttempt: { lt: fiveMinutesAgo } }
            });

            for (const n of pending) {
                const response = await fetch(`https://graph.facebook.com/v17.0/${process.env.WHATSAPP_PHONE_NUMBER_ID}/messages`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${process.env.WHATSAPP_TOKEN}`,
                    },
                    body: JSON.stringify({
                        messaging_product: "whatsapp",
                        recipient_type: "individual",
                        to: n.whatsappFrom,
                        context: { message_id: n.whatsappMessageId },
                        type: "text",
                        text: { body: n.message },
                    }),
                });

                if (response.ok) {
                    // @ts-ignore
                    await prisma.notificationQueue.delete({ where: { id: n.id } });
                } else {
                    // @ts-ignore
                    await prisma.notificationQueue.update({
                        where: { id: n.id },
                        data: { retryCount: n.retryCount + 1, lastAttempt: new Date() }
                    });
                }
            }
        } catch (error) {
            logger.error('[NotificationService] Processing retries error', error);
        }
    }
}
