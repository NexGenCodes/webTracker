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

/**
 * Returns the raw JWT string from the HttpOnly cookie.
 * This is used by server components to pass the token to client
 * components that need it for Supabase Realtime subscriptions.
 */
export async function getJwtTokenAction(): Promise<string | undefined> {
    const cookieStore = await cookies();
    return cookieStore.get('jwt')?.value;
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
                httpOnly: true,
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

export async function registerIntentAction(data: RegisterInput): Promise<ActionResult> {
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

        // Store OTP token in an HttpOnly cookie — never expose to client JS
        if (resData.otp_token) {
            const cookieStore = await cookies();
            cookieStore.set('otp_token', resData.otp_token, {
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                path: '/',
                maxAge: 10 * 60 // 10 minutes
            });
        }

        return { success: true };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function verifyOtpAction(otp: string): Promise<ActionResult> {
    try {
        // Read OTP token from server-side HttpOnly cookie — never from client
        const cookieStore = await cookies();
        const otpToken = cookieStore.get('otp_token')?.value;
        if (!otpToken) {
            return { success: false, error: 'OTP session expired. Please start over.' };
        }

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
            cookieStore.set('jwt', resData.token, {
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                path: '/',
                maxAge: 7 * 24 * 60 * 60
            });
            // Clear the OTP token cookie
            cookieStore.delete('otp_token');
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

export async function forgotPasswordAction(email: string): Promise<ActionResult> {
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

        // Store reset token in an HttpOnly cookie — never expose to client JS
        if (resData.reset_token) {
            const cookieStore = await cookies();
            cookieStore.set('reset_token', resData.reset_token, {
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                path: '/',
                maxAge: 15 * 60 // 15 minutes
            });
        }

        return { success: true };
    } catch (error) {
        Sentry.captureException(error);
        return { success: false, error: 'Network error. Please check your connection.' };
    }
}

export async function resetPasswordAction(email: string, otp: string, newPassword: string): Promise<ActionResult> {
    try {
        // Read reset token from server-side HttpOnly cookie — never from client
        const cookieStore = await cookies();
        const resetToken = cookieStore.get('reset_token')?.value;
        if (!resetToken) {
            return { success: false, error: 'Reset session expired. Please start over.' };
        }

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

        // Clear the reset token cookie
        cookieStore.delete('reset_token');

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
