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

            if (result.data?.reset_token) {
                // Store the reset token in sessionStorage. This is tab-scoped,
                // cleared on tab close, and avoids polluting React state across
                // mode transitions. The backend enforces JWT expiry on the token.
                sessionStorage.setItem('reset_token', result.data.reset_token);
                setEmailCache(data.email);
                setSuccessMessage('Password reset code sent to your email.');
                switchMode('reset-password');
            } else {
                setError('Failed to receive reset token.');
            }
        });
    };

    const onResetPassword = (data: ResetPasswordForm, otp: string) => {
        const resetToken = sessionStorage.getItem('reset_token');
        if (!resetToken) {
            setError("Session expired. Please start over.");
            switchMode('forgot-password');
            return;
        }

        setError(null);
        startTransition(async () => {
            const result = await resetPasswordAction(emailCache, otp, data.password, resetToken);

            if (!result.success) {
                setError(result.error || 'Failed to reset password. Please verify the code and try again.');
                return;
            }

            sessionStorage.removeItem('reset_token');
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
