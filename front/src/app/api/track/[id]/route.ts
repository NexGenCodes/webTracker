import { NextRequest, NextResponse } from 'next/server';
import sql from '@/lib/db';

export async function GET(
    request: NextRequest,
    { params }: { params: Promise<{ id: string }> }
) {
    const { id: trackingID } = await params;

    if (!trackingID) {
        return NextResponse.json({ error: 'tracking_id required' }, { status: 400 });
    }

    try {
        const [shipment] = await sql`
            SELECT * FROM Shipment WHERE tracking_id = ${trackingID}
        `;

        if (!shipment) {
            return NextResponse.json({ error: 'shipment not found' }, { status: 404 });
        }

        // Timeline generation logic (Directly from DB columns)
        const now = new Date();
        const timeline = [
            {
                status: 'Order Placed',
                timestamp: shipment.created_at,
                description: `Shipment registered at ${shipment.origin}`,
                isCompleted: true
            },
            {
                status: 'In Transit',
                timestamp: shipment.scheduled_transit_time,
                description: 'Package has left the origin facility and is on its way',
                isCompleted: now > new Date(shipment.scheduled_transit_time) || ['intransit', 'outfordelivery', 'delivered'].includes(shipment.status.toLowerCase())
            },
            {
                status: 'Delivered',
                timestamp: shipment.expected_delivery_time,
                description: 'Package has arrived at the destination',
                isCompleted: shipment.status.toLowerCase() === 'delivered'
            }
        ];

        const response = {
            tracking_id: shipment.tracking_id,
            status: shipment.status,
            origin: shipment.origin,
            destination: shipment.destination,
            recipient_country: shipment.destination,
            timeline: timeline,
            weight: shipment.weight,
            sender_name: redactName(shipment.sender_name),
            recipient_name: redactName(shipment.recipient_name),
            recipient_address: 'Redacted for privacy'
        };

        return NextResponse.json(response);
    } catch (error) {
        console.error('Tracking API Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}

function redactName(name: string | null): string {
    if (!name) return 'N/A';
    const parts = name.split(' ');
    if (parts[0].length <= 2) {
        return parts[0] + '***';
    }
    return parts[0].substring(0, 2) + '******';
}
