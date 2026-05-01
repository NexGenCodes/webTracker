'use server';

import { getBackendUrl, backendHeaders } from '@/lib/backend';
import { getServerSession } from '@/lib/auth';
import { revalidatePath } from 'next/cache';

export async function subscribeAction(plan: string, callback_url: string) {
    const { user } = await getServerSession();
    if (!user || !user.company_id) {
        throw new Error('Unauthorized');
    }

    if (!plan) {
        throw new Error('plan is required');
    }

    const res = await fetch(`${getBackendUrl()}/api/company/subscribe`, {
        method: 'POST',
        headers: await backendHeaders({
            'X-Company-ID': user.company_id,
            'Content-Type': 'application/json'
        }),
        body: JSON.stringify({
            callback_url,
            plan
        })
    });
    
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Backend error' }));
        throw new Error(err.error || 'Backend error');
    }
    const data = await res.json();
    revalidatePath('/dashboard');
    return data;
}
