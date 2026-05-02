'use server';

import { cookies } from 'next/headers';
import { revalidatePath } from 'next/cache';
import { getApiUrl } from '@/lib/utils';
import { getServerSession } from '@/lib/auth';
import * as Sentry from '@sentry/nextjs';

/**
 * Standardized return type for all server actions.
 * Actions NEVER throw — they return { success: false, error } on failure.
 * This prevents Next.js error boundaries from destroying UI state.
 */
export interface ActionResult<T = undefined> {
    success: boolean;
    error?: string;
    data?: T;
}

interface LoginInput {
    email: string;
    password: string;
}

interface RegisterInput {
    companyName: string;
    email: string;
    password: string;
}

export async function checkAuthAction() {
    return await getServerSession();
}

export async function loginAction(data: LoginInput): Promise<ActionResult> {
    try {
        const res = await fetch(`${getApiUrl()}/api/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                email: data.email,
                password: data.password,
            }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Invalid email or password.' };
        }

        if (resData.token) {
            const cookieStore = await cookies();
            cookieStore.set('jwt', resData.token, {
                httpOnly: false, // Must be false for Supabase Realtime client to read it
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                path: '/',
                maxAge: 7 * 24 * 60 * 60 // 7 days
            });

            // Force revalidation of all critical routes
            revalidatePath('/', 'layout');
            revalidatePath('/', 'page');
            revalidatePath('/dashboard', 'layout');
            revalidatePath('/auth', 'layout');
        }

        return { success: true };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function registerIntentAction(data: RegisterInput): Promise<ActionResult<{ otp_token?: string }>> {
    try {
        const res = await fetch(`${getApiUrl()}/api/auth/register-intent`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                company_name: data.companyName.trim(),
                email: data.email,
                password: data.password,
            }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Registration failed.' };
        }

        return { success: true, data: { otp_token: resData.otp_token } };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function verifyOtpAction(otp: string, otpToken: string): Promise<ActionResult> {
    try {
        const res = await fetch(`${getApiUrl()}/api/auth/verify-otp`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-OTP-Token': otpToken
            },
            body: JSON.stringify({ otp }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Verification failed.' };
        }

        if (resData.token) {
            const cookieStore = await cookies();
            cookieStore.set('jwt', resData.token, {
                httpOnly: false, // Must be false for Supabase Realtime client to read it
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                path: '/',
                maxAge: 7 * 24 * 60 * 60
            });
            revalidatePath('/', 'layout');
            revalidatePath('/dashboard');
            revalidatePath('/auth');
        }

        return { success: true };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function forgotPasswordAction(email: string): Promise<ActionResult<{ reset_token?: string }>> {
    try {
        const res = await fetch(`${getApiUrl()}/api/auth/forgot-password`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to send reset code.' };
        }

        return { success: true, data: { reset_token: resData.reset_token } };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function resetPasswordAction(email: string, otp: string, newPassword: string, resetToken: string): Promise<ActionResult> {
    try {
        const res = await fetch(`${getApiUrl()}/api/auth/reset-password`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Reset-Token': resetToken
            },
            body: JSON.stringify({ email, otp, new_password: newPassword }),
        });

        const resData = await res.json();
        if (!res.ok) {
            return { success: false, error: resData.error || 'Failed to reset password.' };
        }

        return { success: true };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function logoutAction(): Promise<ActionResult> {
    const cookieStore = await cookies();
    cookieStore.delete('jwt');

    // Force revalidation of all critical routes
    revalidatePath('/', 'layout');
    revalidatePath('/', 'page');
    revalidatePath('/dashboard', 'layout');
    revalidatePath('/auth', 'layout');

    return { success: true };
}
