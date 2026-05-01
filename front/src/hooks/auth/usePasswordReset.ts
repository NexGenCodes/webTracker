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
import { forgotPasswordAction, resetPasswordAction } from '@/app/actions/auth';

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
            const resData = await forgotPasswordAction(data.email);

            if (resData.reset_token) {
                sessionStorage.setItem('reset_token', resData.reset_token);
            }

            setEmailCache(data.email);
            setSuccessMessage("A 6-digit reset code has been sent to your email.");
            setTimeout(() => switchMode('reset-password'), 2000);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Network error. Please try again.');
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
            const resetToken = sessionStorage.getItem('reset_token') || '';
            await resetPasswordAction(emailCache, code, data.password, resetToken);

            setSuccessMessage("Password reset successfully. You can now sign in.");
            setTimeout(() => switchMode('signin'), 2000);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Network error. Please try again.');
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
