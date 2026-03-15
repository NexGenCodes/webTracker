import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import sql from '@/lib/db';

export async function PATCH(
    request: NextRequest,
    { params }: { params: Promise<{ id: string }> }
) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const { id } = await params;
    try {
        const input = await request.json();
        const now = new Date();

        await sql`
            UPDATE Shipment SET
                status = ${input.status || 'pending'},
                sender_name = ${input.senderName},
                recipient_name = ${input.recipientName},
                destination = ${input.destination},
                origin = ${input.origin},
                updated_at = ${now}
            WHERE tracking_id = ${id}
        `;

        return NextResponse.json({ success: true });
    } catch (error) {
        console.error('Update Shipment API Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}

export async function DELETE(
    request: NextRequest,
    { params }: { params: Promise<{ id: string }> }
) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const { id } = await params;
    try {
        if (id === 'cleanup') {
            const result = await sql`
                DELETE FROM Shipment WHERE status = 'delivered' AND updated_at < NOW() - INTERVAL '30 days'
            `;
            return NextResponse.json({ deleted_count: result.count });
        } else {
            await sql`
                DELETE FROM Shipment WHERE tracking_id = ${id}
            `;
            return NextResponse.json({ success: true });
        }
    } catch (error) {
        console.error('Delete Shipment API Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
