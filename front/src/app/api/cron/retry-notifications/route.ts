import { NextResponse } from 'next/server';
import { processNotificationQueue } from '@/app/actions/shipment';

export async function GET(request: Request) {
    const authHeader = request.headers.get('authorization');

    // Support dual authentication: Vercel native cron + external cron (cron-job.org)
    const isVercelCron = authHeader === `Bearer ${process.env.CRON_SECRET}`;
    const isExternalCron = authHeader === `Bearer ${process.env.EXTERNAL_CRON_SECRET}`;

    if (process.env.NODE_ENV === 'production' && !isVercelCron && !isExternalCron) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        await processNotificationQueue();
        return NextResponse.json({
            success: true,
            message: 'Notification retry queue processed.',
            timestamp: new Date().toISOString()
        });
    } catch (error) {
        console.error('Retry notifications cron job failed:', error);
        return NextResponse.json({ error: 'Retry processing failed' }, { status: 500 });
    }
}
