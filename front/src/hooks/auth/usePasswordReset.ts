import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { 
    forgotPasswordSchema, 
    resetPasswordSchema, 
    ForgotPasswordForm, 
    ResetPasswordForm,
    AuthMode 
} from '@/lib/validations/auth';
import { getApiUrl } from '@/lib/utils';

const API_URL = getApiUrl();

export function usePasswordReset(
    setError: (msg: string | null) => void,
    setSuccessMessage: (msg: string | null) => void,
    setEmailCache: (email: string) => void,
    switchMode: (mode: AuthMode) => void,
    emailCache: string
) {
    const [loading, setLoading] = useState(false);

    const forgotPasswordForm = useForm<ForgotPasswordForm>({ 
        resolver: zodResolver(forgotPasswordSchema) 
    });
    
    const resetPasswordForm = useForm<ResetPasswordForm>({ 
        resolver: zodResolver(resetPasswordSchema) 
    });

    const onForgotPassword = async (data: ForgotPasswordForm) => {
        setError(null);
        setLoading(true);
        try {
            const res = await fetch(`${API_URL}/api/auth/forgot-password`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: data.email }),
            });

            if (!res.ok) {
                const resData = await res.json();
                setError(resData.error || "Failed to send reset code.");
                return;
            }

            setEmailCache(data.email);
            setSuccessMessage("A 6-digit reset code has been sent to your email.");
            setTimeout(() => switchMode('reset-password'), 2000);
        } catch {
            setError("Network error. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    const onResetPassword = async (data: ResetPasswordForm, code: string) => {
        setError(null);
        if (code.length !== 6) {
            setError("Please enter the full 6-digit code.");
            return;
        }

        setLoading(true);
        try {
            const res = await fetch(`${API_URL}/api/auth/reset-password`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: emailCache, otp: code, new_password: data.password }),
            });

            if (!res.ok) {
                const resData = await res.json();
                setError(resData.error || "Failed to reset password.");
                return;
            }

            setSuccessMessage("Password reset successfully. You can now sign in.");
            setTimeout(() => switchMode('signin'), 2000);
        } catch {
            setError("Network error. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    return { 
        forgotPasswordForm, 
        resetPasswordForm, 
        onForgotPassword, 
        onResetPassword, 
        loading 
    };
}
