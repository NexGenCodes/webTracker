import { NextRequest, NextResponse } from 'next/server';
import { getBackendUrl, backendHeaders } from '@/lib/backend';
import { jwtVerify } from 'jose';

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET || '');

export async function POST(request: NextRequest) {
    try {
        // Verify JWT before proxying — prevents unauthenticated pairing
        const jwt = request.cookies.get('jwt')?.value;
        if (!jwt) {
            return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
        }
        try {
            await jwtVerify(jwt, JWT_SECRET);
        } catch {
            return NextResponse.json({ error: 'Invalid or expired token' }, { status: 401 });
        }

        const { company_id, phone } = await request.json();
        if (!company_id || !phone) {
            return NextResponse.json({ error: 'company_id and phone are required' }, { status: 400 });
        }

        const res = await fetch(`${getBackendUrl()}/api/company/pair`, {
            method: 'POST',
            headers: await backendHeaders({
                'X-Company-ID': company_id
            }),
            body: JSON.stringify({ phone })
        });
        
        if (!res.ok) {
            const err = await res.json().catch(() => ({ error: 'Backend error' }));
            return NextResponse.json(err, { status: res.status });
        }
        
        const data = await res.json();
        return NextResponse.json(data);
    } catch (error) {
        console.error('Pairing Proxy Error:', error);
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
    }
}
