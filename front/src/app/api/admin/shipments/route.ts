import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from '@/lib/auth';
import { getBackendUrl, backendHeaders } from '@/lib/backend';

// GET reads are now handled directly via Supabase in ShipmentService.
// Only the POST (write) proxy to Go remains here.

export async function POST(request: NextRequest) {
    const { authenticated } = await getServerSession();
    if (!authenticated) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const input = await request.json();
        const res = await fetch(`${getBackendUrl()}/api/admin/shipments`, {
            method: 'POST',
            headers: backendHeaders(),
            body: JSON.stringify(input)
        });
        
        if (!res.ok) throw new Error('Backend error');
        const data = await res.json();

        return NextResponse.json(data);
    } catch (error) {
        console.error('Create Shipment Proxy Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
