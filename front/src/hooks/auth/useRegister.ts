import { useState, useEffect, useTransition } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { registerSchema, RegisterForm, RegisterStep } from '@/lib/validations/auth';
import { registerIntentAction, verifyOtpAction } from '@/app/actions/auth';

export function useRegister(
    setError: (msg: string | null) => void,
    setEmailCache: (email: string) => void,
    setRegisterStep: (step: RegisterStep) => void,
    currentStep: RegisterStep
) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [isPending, startTransition] = useTransition();
    const [otpTimer, setOtpTimer] = useState(600); // 10 minutes

    const form = useForm<RegisterForm>({ 
        resolver: zodResolver(registerSchema),
        defaultValues: {
            companyName: '',
            email: '',
            password: '',
            confirmPassword: '',
            acceptTerms: false
        }
    });

    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (currentStep === 'otp' && otpTimer > 0) {
            interval = setInterval(() => {
                setOtpTimer((prev) => prev - 1);
            }, 1000);
        }
        return () => clearInterval(interval);
    }, [currentStep, otpTimer]);

    const onRegisterIntent = (data: RegisterForm) => {
        setError(null);
        startTransition(async () => {
            try {
                const resData = await registerIntentAction(data);
                
                if (resData.otp_token) {
                    sessionStorage.setItem('otp_token', resData.otp_token);
                }

                setEmailCache(data.email);
                setOtpTimer(600);
                setRegisterStep('otp');
            } catch (err: unknown) {
                setError(err instanceof Error ? err.message : 'Network error. Please try again.');
            }
        });
    };

    const verifyOtp = (code: string) => {
        setError(null);
        if (code.length !== 6) {
            setError("Please enter the full 6-digit code.");
            return;
        }

        startTransition(async () => {
            try {
                const otpToken = sessionStorage.getItem('otp_token') || '';
                await verifyOtpAction(code, otpToken);

                const redirectUrl = searchParams.get('redirect') || searchParams.get('callbackUrl') || searchParams.get('returnUrl') || '/dashboard';
                router.push(redirectUrl);
                router.refresh();
            } catch (err: unknown) {
                setError(err instanceof Error ? err.message : 'Network error. Please try again.');
            }
        });
    };

    return { form, onRegisterIntent, verifyOtp, loading: isPending, otpTimer, setOtpTimer };
}
