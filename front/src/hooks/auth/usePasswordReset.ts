import { useTransition } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { forgotPasswordSchema, resetPasswordSchema, ForgotPasswordForm, ResetPasswordForm } from '@/lib/validations/auth';
import { forgotPasswordAction, resetPasswordAction } from '@/app/actions/auth';

export function usePasswordReset(
    setError: (msg: string | null) => void,
    setSuccessMessage: (msg: string | null) => void,
    setEmailCache: (email: string) => void,
    switchMode: (mode: 'signin' | 'register' | 'forgot-password' | 'reset-password') => void,
    emailCache: string
) {
    const [isPending, startTransition] = useTransition();

    const forgotPasswordForm = useForm<ForgotPasswordForm>({
        resolver: zodResolver(forgotPasswordSchema),
        defaultValues: { email: emailCache }
    });
    const resetPasswordForm = useForm<ResetPasswordForm>({ resolver: zodResolver(resetPasswordSchema) });

    const onForgotPassword = (data: ForgotPasswordForm) => {
        setError(null);
        setSuccessMessage(null);
        startTransition(async () => {
            const result = await forgotPasswordAction(data.email);

            if (!result.success) {
                setError(result.error || 'Failed to send reset code. Please try again.');
                return;
            }

            setEmailCache(data.email);
            setSuccessMessage('Password reset code sent to your email.');
            switchMode('reset-password');
        });
    };

    const onResetPassword = (data: ResetPasswordForm, otp: string) => {
        setError(null);
        startTransition(async () => {
            const result = await resetPasswordAction(emailCache, otp, data.password);

            if (!result.success) {
                if (result.error?.includes('expired')) {
                    switchMode('forgot-password');
                }
                setError(result.error || 'Failed to reset password. Please verify the code and try again.');
                return;
            }

            setSuccessMessage("Password reset successfully. Please sign in.");
            switchMode('signin');
        });
    };

    return {
        forgotPasswordForm,
        resetPasswordForm,
        loading: isPending,
        onForgotPassword,
        onResetPassword
    };
}
