import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { registerSchema, RegisterForm, RegisterStep } from '@/lib/validations/auth';
import { getApiUrl } from '@/lib/utils';

const API_URL = getApiUrl();

export function useRegister(
    setError: (msg: string | null) => void,
    setEmailCache: (email: string) => void,
    setRegisterStep: (step: RegisterStep) => void,
    currentStep: RegisterStep
) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [loading, setLoading] = useState(false);
    const [otpTimer, setOtpTimer] = useState(600); // 10 minutes

    const form = useForm<RegisterForm>({ resolver: zodResolver(registerSchema) });

    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (currentStep === 'otp' && otpTimer > 0) {
            interval = setInterval(() => {
                setOtpTimer((prev) => prev - 1);
            }, 1000);
        }
        return () => clearInterval(interval);
    }, [currentStep, otpTimer]);

    const onRegisterIntent = async (data: RegisterForm) => {
        setError(null);
        setLoading(true);
        try {
            const res = await fetch(`${API_URL}/api/auth/register-intent`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    company_name: data.companyName.trim(),
                    email: data.email,
                    password: data.password,
                }),
            });

            if (!res.ok) {
                const resData = await res.json();
                setError(resData.error || 'Registration failed.');
                setLoading(false);
                return;
            }

            const resData = await res.json();
            if (resData.otp_token) {
                sessionStorage.setItem('otp_token', resData.otp_token);
            }

            setEmailCache(data.email);
            setOtpTimer(600);
            setRegisterStep('otp');
            setLoading(false);
        } catch {
            setError("Network error. Please try again.");
            setLoading(false);
        }
    };

    const verifyOtp = async (code: string) => {
        setError(null);
        if (code.length !== 6) {
            setError("Please enter the full 6-digit code.");
            return;
        }

        setLoading(true);
        try {
            const otpToken = sessionStorage.getItem('otp_token') || '';
            const res = await fetch(`${API_URL}/api/auth/verify-otp`, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'X-OTP-Token': otpToken
                },
                credentials: 'include',
                body: JSON.stringify({ otp: code }),
            });

            if (!res.ok) {
                const resData = await res.json();
                setError(resData.error || 'Verification failed.');
                setLoading(false);
                return;
            }

            const redirectUrl = searchParams.get('redirect') || searchParams.get('callbackUrl') || searchParams.get('returnUrl') || '/dashboard';
            router.push(redirectUrl);
            router.refresh();
        } catch {
            setError("Network error. Please try again.");
            setLoading(false);
        }
    };

    return { form, onRegisterIntent, verifyOtp, loading, otpTimer, setOtpTimer };
}
