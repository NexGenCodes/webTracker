import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import sql from '@/lib/db';


export async function GET(request: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const shipments = await sql`
            SELECT * FROM Shipment ORDER BY created_at DESC
        `;
        return NextResponse.json(shipments);
    } catch (error) {
        console.error('List Shipments API Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}

export async function POST(request: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const input = await request.json();

        // Generate dynamic tracking ID
        const prefix = process.env.NEXT_PUBLIC_COMPANY_PREFIX || process.env.COMPANY_PREFIX || 'AWB';
        const randStr = Math.floor(Math.random() * 1000000000).toString().padStart(9, '0');
        const trackingId = `${prefix}-${randStr}`;

        // DB trigger auto-generates: scheduled_transit_time,
        // outfordelivery_time, expected_delivery_time, updated_at
        const result = await sql`
            INSERT INTO Shipment (
                tracking_id, user_jid, status,
                sender_name, origin, recipient_name, recipient_phone,
                recipient_email, recipient_address, destination,
                cargo_type, weight
            ) VALUES (
                ${trackingId}, 'admin-ui', 'pending',
                ${input.senderName || null}, ${input.senderCountry}, ${input.receiverName},
                ${input.number || null}, ${input.email || null}, ${input.address || null},
                ${input.receiverCountry}, ${input.cargoType || null}, ${input.weight || 0}
            )
            RETURNING tracking_id
        `;

        return NextResponse.json({ tracking_id: result[0].tracking_id });
    } catch (error) {
        console.error('Create Shipment API Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}

