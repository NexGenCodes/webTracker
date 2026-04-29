import { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { signInSchema, SignInForm } from '@/lib/validations/auth';
import { getApiUrl } from '@/lib/utils';

const API_URL = getApiUrl();

export function useSignIn(setError: (msg: string | null) => void) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [loading, setLoading] = useState(false);

    const form = useForm<SignInForm>({ resolver: zodResolver(signInSchema) });

    const onSubmit = async (data: SignInForm) => {
        setLoading(true);
        setError(null);

        try {
            const res = await fetch(`${API_URL}/api/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    email: data.email,
                    password: data.password,
                }),
            });

            if (!res.ok) {
                const resData = await res.json();
                setError(resData.error || 'Invalid email or password.');
                setLoading(false);
                return;
            }

            const redirectUrl = searchParams.get('redirect') || searchParams.get('callbackUrl') || searchParams.get('returnUrl') || '/dashboard';
            router.push(redirectUrl);
            router.refresh();
        } catch {
            setError("Something went wrong. Please try again later.");
            setLoading(false);
        }
    };

    const handleGoogleSignIn = async () => {
        setError('Google sign-in is coming soon.');
    };

    return { form, onSubmit, loading, handleGoogleSignIn };
}
