import { NextResponse } from 'next/server';
import { pruneOldShipments } from '@/app/actions/shipment';

export async function GET(request: Request) {
    const authHeader = request.headers.get('authorization');

    // In production, Vercel sends the CRON_SECRET in the Authorization header
    if (process.env.NODE_ENV === 'production' && authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        await pruneOldShipments();
        return NextResponse.json({
            success: true,
            message: 'Database maintenance completed.',
            timestamp: new Date().toISOString()
        });
    } catch (error) {
        console.error('Cron job failed:', error);
        return NextResponse.json({ error: 'Maintenance failed' }, { status: 500 });
    }
}
