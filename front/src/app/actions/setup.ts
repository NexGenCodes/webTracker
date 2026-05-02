'use server';

import { getApiUrl } from '@/lib/utils';
import { getServerSession } from '@/lib/auth';
import { ActionResult } from './auth';

export async function pairWhatsApp(companyId: string, phone: string): Promise<ActionResult<{ code?: string }>> {
    try {
        const session = await getServerSession();
        if (!session.user?.company_id || session.user.company_id !== companyId) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/pair`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${session.token}`
            },
            body: JSON.stringify({ phone }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to request pairing code.' };
        }

        return { success: true, data: { code: resData.code } };
    } catch {
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function checkWhatsAppStatus(companyId: string): Promise<ActionResult<{ connected: boolean; phone?: string }>> {
    try {
        const session = await getServerSession();
        if (!session.user?.company_id || session.user.company_id !== companyId) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/status`, {
            headers: {
                'Authorization': `Bearer ${session.token}`
            }
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to check status.' };
        }

        return { success: true, data: { connected: resData.connected, phone: resData.phone } };
    } catch {
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function disconnectWhatsApp(companyId: string): Promise<ActionResult> {
    try {
        const session = await getServerSession();
        if (!session.user?.company_id || session.user.company_id !== companyId) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/logout`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${session.token}`
            }
        });

        if (!res.ok) {
            const resData = await res.json();
            return { success: false, error: resData.error || 'Failed to disconnect.' };
        }

        return { success: true };
    } catch {
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function deleteAccount(companyId: string): Promise<ActionResult> {
    try {
        const session = await getServerSession();
        if (!session.user?.company_id || session.user.company_id !== companyId) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/delete`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${session.token}`
            }
        });

        if (!res.ok) {
            const resData = await res.json();
            return { success: false, error: resData.error || 'Failed to delete account.' };
        }

        return { success: true };
    } catch {
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function getWhatsAppQR(companyId: string): Promise<ActionResult<{ pairingCode?: string }>> {
    try {
        const session = await getServerSession();
        if (!session.user?.company_id || session.user.company_id !== companyId) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/qr`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${session.token}`
            }
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to fetch QR code.' };
        }

        return { success: true, data: { pairingCode: resData.code || resData.data?.code } };
    } catch {
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}
