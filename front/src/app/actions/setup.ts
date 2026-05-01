'use server';

import { getBackendUrl, backendHeaders } from '@/lib/backend';
import { getServerSession } from '@/lib/auth';
import { revalidatePath } from 'next/cache';

export interface WhatsAppPairResponse {
    success: boolean;
    pairingCode?: string;
    error?: string;
}

export async function pairWhatsApp(companyId: string, phone: string): Promise<WhatsAppPairResponse> {
    const { user } = await getServerSession();
    if (!user || user.company_id !== companyId) {
        throw new Error('Unauthorized');
    }

    if (!companyId || !phone) {
        throw new Error('company_id and phone are required');
    }

    const res = await fetch(`${getBackendUrl()}/api/company/pair`, {
        method: 'POST',
        headers: await backendHeaders({
            'X-Company-ID': companyId
        }),
        body: JSON.stringify({ phone })
    });
    
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Backend error' }));
        throw new Error(err.error || 'Backend error');
    }
    
    const data = await res.json();
    
    revalidatePath('/dashboard');
    return {
        success: true,
        pairingCode: data.pairing_code || data.code || data.data?.pairing_code || data.data?.code || data.pairingCode || data.data?.pairingCode
    };
}

export async function disconnectWhatsApp(companyId: string): Promise<{ success: boolean; error?: string }> {
    const { user } = await getServerSession();
    if (!user || user.company_id !== companyId) {
        throw new Error('Unauthorized');
    }

    if (!companyId) {
        throw new Error('company_id is required');
    }

    const res = await fetch(`${getBackendUrl()}/api/company/logout`, {
        method: 'POST',
        headers: await backendHeaders({
            'X-Company-ID': companyId
        })
    });
    
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Backend error' }));
        throw new Error(err.error || 'Backend error');
    }
    
    await res.json();
    revalidatePath('/dashboard');
    return { success: true };
}

export async function getWhatsAppQR(companyId: string): Promise<WhatsAppPairResponse> {
    const { user } = await getServerSession();
    if (!user || user.company_id !== companyId) {
        throw new Error('Unauthorized');
    }

    if (!companyId) {
        throw new Error('company_id is required');
    }

    const res = await fetch(`${getBackendUrl()}/api/company/qr`, {
        method: 'POST',
        headers: await backendHeaders({
            'X-Company-ID': companyId
        })
    });
    
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Backend error' }));
        throw new Error(err.error || 'Backend error');
    }
    
    const data = await res.json();
    
    return {
        success: true,
        pairingCode: data.code || data.data?.code
    };
}

export async function deleteAccount(companyId: string): Promise<{ success: boolean; error?: string }> {
    const { user } = await getServerSession();
    if (!user || user.company_id !== companyId) {
        throw new Error('Unauthorized');
    }

    if (!companyId) {
        throw new Error('company_id is required');
    }

    const res = await fetch(`${getBackendUrl()}/api/company/delete`, {
        method: 'DELETE',
        headers: await backendHeaders({
            'X-Company-ID': companyId
        })
    });

    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Backend error' }));
        throw new Error(err.error || 'Failed to delete account');
    }

    await res.json();
    revalidatePath('/dashboard');
    return { success: true };
}
