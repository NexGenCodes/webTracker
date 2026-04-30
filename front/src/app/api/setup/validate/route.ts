import { NextRequest, NextResponse } from 'next/server';
import { getBackendUrl, backendHeaders } from '@/lib/backend';

export async function POST(request: NextRequest) {
    try {
        const { token } = await request.json();
        if (!token) {
            return NextResponse.json({ error: 'token is required' }, { status: 400 });
        }

        const res = await fetch(`${getBackendUrl()}/api/company/setup/${token}`, {
            headers: await backendHeaders(),
            cache: 'no-store'
        });

        if (!res.ok) {
            return NextResponse.json({ error: 'Invalid or expired setup token' }, { status: 404 });
        }

        const data = await res.json();
        return NextResponse.json(data);
    } catch (error) {
        console.error('Setup Validate Proxy Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
