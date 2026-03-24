import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';

function getBackendUrl() {
    return process.env.BACKEND_URL || 'http://localhost:5000';
}

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
        const res = await fetch(`${getBackendUrl()}/api/admin/shipments/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(input)
        });
        
        if (!res.ok) throw new Error('Backend error');
        return NextResponse.json(await res.json());
    } catch (error) {
        console.error('Update Shipment Proxy Error:', error);
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
        const endpoint = id === 'cleanup' ? 'cleanup' : id;
        const res = await fetch(`${getBackendUrl()}/api/admin/shipments/${endpoint}`, {
            method: 'DELETE'
        });
        
        if (!res.ok) throw new Error('Backend error');
        return NextResponse.json(await res.json());
    } catch (error) {
        console.error('Delete Shipment Proxy Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
