import { useState, useTransition, useRef, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { registerSchema, RegisterForm } from '@/lib/validations/auth';
import { registerIntentAction, verifyOtpAction } from '@/app/actions/auth';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';

export function useRegister(
    setError: (msg: string | null) => void,
    setEmailCache: (email: string) => void,
    setRegisterStep: (step: 'credentials' | 'otp') => void,
    registerStep: 'credentials' | 'otp'
) {
    const [isPending, startTransition] = useTransition();
    const { refreshAuth } = useMultiTenant();

    const [otpToken, setOtpToken] = useState<string | null>(null);
    const [otpTimer, setOtpTimer] = useState(0);
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    // Clean up interval on unmount to prevent memory leaks
    useEffect(() => {
        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        };
    }, []);

    const form = useForm<RegisterForm>({ resolver: zodResolver(registerSchema) });

    const startOtpTimer = () => {
        // Clear any existing interval before starting a new one (e.g. on resend)
        if (intervalRef.current) {
            clearInterval(intervalRef.current);
        }
        setOtpTimer(600);
        intervalRef.current = setInterval(() => {
            setOtpTimer((prev) => {
                if (prev <= 1) {
                    if (intervalRef.current) {
                        clearInterval(intervalRef.current);
                        intervalRef.current = null;
                    }
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);
    };

    const onRegisterIntent = (data: RegisterForm) => {
        setError(null);
        startTransition(async () => {
            const result = await registerIntentAction({
                companyName: data.companyName,
                email: data.email,
                password: data.password
            });

            if (!result.success) {
                setError(result.error || 'Registration failed. Please try again.');
                return;
            }

            if (result.data?.otp_token) {
                setOtpToken(result.data.otp_token);
                setEmailCache(data.email);
                setRegisterStep('otp');
                startOtpTimer();
            } else {
                setError('Failed to receive OTP verification token.');
            }
        });
    };

    const verifyOtp = async (fullOtp: string) => {
        if (fullOtp.length !== 6) {
            setError('Please enter all 6 digits.');
            return;
        }
        if (!otpToken) {
            setError('Missing verification token. Please start over.');
            setRegisterStep('credentials');
            return;
        }

        setError(null);
        startTransition(async () => {
            const result = await verifyOtpAction(fullOtp, otpToken);

            if (!result.success) {
                setError(result.error || 'Verification failed. Please check the code and try again.');
                return;
            }

            await refreshAuth();
            window.location.href = '/dashboard';
        });
    };

    return {
        form,
        loading: isPending,
        onRegisterIntent,
        verifyOtp,
        otpTimer
    };
}
