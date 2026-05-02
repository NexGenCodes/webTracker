'use server';

import { getApiUrl } from '@/lib/utils';
import { getServerSession } from '@/lib/auth';
import { ActionResult } from './auth';
import * as Sentry from '@sentry/nextjs';

export interface PlanData {
    id: string;
    name: string;
    name_key: string;
    desc_key: string;
    price: number;
    currency: string;
    interval_key: string;
    popular: boolean;
    trial_key?: string;
    btn_key: string;
    features: string[];
}

export async function getPlansAction(): Promise<ActionResult<PlanData[]>> {
    try {
        const res = await fetch(`${getApiUrl()}/api/billing/plans`, {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' },
            next: { revalidate: 300, tags: ['plans', 'billing'] } // Cache for 5 minutes and tag for targeted invalidation
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: 'Failed to fetch plans.' };
        }

        return { success: true, data: resData };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error.' };
    }
}

export async function subscribeAction(plan: string, callback_url: string): Promise<ActionResult<{ authorization_url?: string, reference?: string }>> {
    try {
        const session = await getServerSession();
        if (!session?.user?.company_id) {
            return { success: false, error: 'Unauthorized.' };
        }

        if (!plan) {
            return { success: false, error: 'Plan is required.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/subscribe`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${session.token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                callback_url,
                plan
            })
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to start subscription.' };
        }

        return { success: true, data: { authorization_url: resData.authorization_url, reference: resData.reference } };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}
export async function getSubscriptionStatusAction(): Promise<ActionResult<{ status: string, plan: string, expiry: string }>> {
    try {
        const session = await getServerSession();
        if (!session?.user?.company_id) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/subscription-status`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${session.token}`,
                'Content-Type': 'application/json'
            }
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to fetch status.' };
        }

        return { success: true, data: resData };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error.' };
    }
}

/**
 * Performs a single check of the subscription status after payment.
 * 
 * NOTE: Do NOT use polling here — the frontend should subscribe to
 * Supabase Realtime `postgres_changes` on the `companies` table to
 * detect status transitions instantly. This action is a one-shot
 * fallback for the initial check after redirect.
 */
export async function checkPaymentStatusAction(): Promise<ActionResult<{ status: string }>> {
    try {
        const session = await getServerSession();
        if (!session?.user?.company_id) {
            return { success: false, error: 'Unauthorized.' };
        }

        const res = await fetch(`${getApiUrl()}/api/company/subscription-status`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${session.token}`,
                'Content-Type': 'application/json'
            },
            cache: 'no-store'
        });

        if (res.ok) {
            const data = await res.json();
            return { success: true, data: { status: data.status || 'pending' } };
        }

        return { success: false, error: 'Failed to check payment status.' };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error.' };
    }
}

