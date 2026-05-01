'use server';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { revalidatePath } from 'next/cache';
import { getApiUrl } from '@/lib/utils';
import { getServerSession } from '@/lib/auth';

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

export async function loginAction(data: LoginInput) {
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
        throw new Error(resData.error || 'Invalid email or password.');
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
        revalidatePath('/', 'layout');
    }

    return resData;
}

export async function registerIntentAction(data: RegisterInput) {
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
        throw new Error(resData.error || 'Registration failed.');
    }

    return resData;
}

export async function verifyOtpAction(otp: string, otpToken: string) {
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
        throw new Error(resData.error || 'Verification failed.');
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
    }

    return resData;
}

export async function forgotPasswordAction(email: string) {
    const res = await fetch(`${getApiUrl()}/api/auth/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
    });

    const resData = await res.json();
    if (!res.ok) {
        throw new Error(resData.error || 'Failed to send reset code.');
    }

    return resData;
}

export async function resetPasswordAction(email: string, otp: string, newPassword: string, resetToken: string) {
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
        throw new Error(resData.error || 'Failed to reset password.');
    }

    return resData;
}

export async function logoutAction() {
    // We clear the cookie by deleting it in the Next.js server environment
    const cookieStore = await cookies();
    cookieStore.delete('jwt');

    // Redirect the user to the auth page
    redirect('/auth');
}
