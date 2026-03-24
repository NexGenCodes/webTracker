import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';

function getBackendUrl() {
    return process.env.BACKEND_URL || 'http://localhost:5000';
}

export async function GET(request: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
        const res = await fetch(`${getBackendUrl()}/api/admin/stats`, {
            headers: { 'Content-Type': 'application/json' },
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
