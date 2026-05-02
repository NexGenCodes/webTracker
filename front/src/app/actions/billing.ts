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

export async function waitForPaymentAction(reference: string): Promise<ActionResult<{ status: string }>> {
    try {
        const session = await getServerSession();
        if (!session?.user?.company_id) {
            return { success: false, error: 'Unauthorized.' };
        }

        const maxAttempts = 12;
        const delayMs = 2500; // 30 seconds total max wait

        for (let i = 0; i < maxAttempts; i++) {
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
                if (data.status === 'active') {
                    return { success: true, data: { status: 'active' } };
                }
            }

            // Wait before next attempt
            await new Promise(resolve => setTimeout(resolve, delayMs));
        }

        return { success: false, error: 'Payment verification is taking longer than usual. It will update automatically once processed.' };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error.' };
    }
}
