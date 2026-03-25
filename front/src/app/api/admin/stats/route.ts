import { NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getBackendUrl, backendHeaders } from '@/lib/backend';

export async function GET() {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const res = await fetch(`${getBackendUrl()}/api/admin/stats`, {
            headers: backendHeaders(),
            cache: 'no-store'
        });

        if (!res.ok) throw new Error('Backend error');
        const data = await res.json();

        return NextResponse.json(data);
    } catch (error) {
        console.error('Stats Proxy Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
