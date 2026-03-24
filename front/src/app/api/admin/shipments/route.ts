import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import sql from '@/lib/db';


export async function GET(request: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const { searchParams } = new URL(request.url);
    const page = parseInt(searchParams.get('page') || '1');
    const limit = parseInt(searchParams.get('limit') || '20');
    const search = searchParams.get('search') || '';
    const status = searchParams.get('status') || '';
    const offset = (page - 1) * limit;

    try {
        // Dynamic search and status filtering
        const shipments = await sql`
            SELECT *, COUNT(*) OVER() as full_count 
            FROM Shipment 
            WHERE 1=1
            ${status ? sql`AND status = ${status}` : sql``}
            ${search ? sql`AND (tracking_id ILIKE ${'%' + search + '%'} OR recipient_name ILIKE ${'%' + search + '%'} OR recipient_phone ILIKE ${'%' + search + '%'})` : sql``}
            ORDER BY created_at DESC
            LIMIT ${limit} OFFSET ${offset}
        `;
        
        const total = shipments.length > 0 ? parseInt(shipments[0].full_count) : 0;

        return NextResponse.json({
            data: shipments.map(s => {
                const { full_count, ...shipment } = s;
                return shipment;
            }),
            pagination: {
                total,
                page,
                limit,
                totalPages: Math.ceil(total / limit)
            }
        });
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
        const { generateAbbreviation } = await import('@/lib/utils');
        const companyName = process.env.NEXT_PUBLIC_COMPANY_NAME || 'Airwaybill';
        const prefix = process.env.NEXT_PUBLIC_COMPANY_PREFIX || process.env.COMPANY_PREFIX || generateAbbreviation(companyName);
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
                ${input.receiverPhone || null}, ${input.receiverEmail || null}, ${input.receiverAddress || null},
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

