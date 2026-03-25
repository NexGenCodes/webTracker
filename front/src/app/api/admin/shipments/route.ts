import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getBackendUrl, backendHeaders } from '@/lib/backend';

export async function GET(request: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const { search } = new URL(request.url);
        const res = await fetch(`${getBackendUrl()}/api/admin/shipments${search}`, {
            headers: backendHeaders(),
            cache: 'no-store'
        });
        
        if (!res.ok) throw new Error('Backend error');
        const data = await res.json();
        
        return NextResponse.json(data);
    } catch (error) {
        console.error('List Shipments Proxy Error:', error);
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


