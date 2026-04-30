import React from 'react';
import { useI18n } from '@/components/providers/I18nContext';
import { Key, Mail, Loader2, ChevronLeft } from 'lucide-react';
import { AuthMode } from '@/lib/validations/auth';
import { usePasswordReset } from '@/hooks/auth/usePasswordReset';
import { AuthInput } from './ui/AuthInput';
import { AuthBanner } from './ui/AuthBanner';

interface ForgotPasswordViewProps {
    switchMode: (mode: AuthMode) => void;
    emailCache: string;
    setEmailCache: (email: string) => void;
    error: string | null;
    setError: (error: string | null) => void;
    successMessage: string | null;
    setSuccessMessage: (msg: string | null) => void;
}

export function ForgotPasswordView({ 
    switchMode, 
    emailCache, 
    setEmailCache, 
    error, 
    setError, 
    successMessage, 
    setSuccessMessage 
}: ForgotPasswordViewProps) {
    const { forgotPasswordForm, onForgotPassword, loading } = usePasswordReset(
        setError,
        setSuccessMessage,
        setEmailCache,
        switchMode,
        emailCache
    );
    const { dict } = useI18n();

    return (
        <form onSubmit={forgotPasswordForm.handleSubmit(onForgotPassword)} className="space-y-6 animate-fade-in w-full">
            <div className="flex flex-col items-center mb-10 text-center">
                <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6">
                    <Key className="text-white" size={32} />
                </div>
                <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">
                    {dict.auth?.resetPassword || 'Reset Password'}
                </h1>
                <p className="text-text-muted font-bold text-sm uppercase tracking-widest opacity-70">
                    {dict.auth?.enterEmail || 'Enter email to receive code'}
                </p>
            </div>

            <AuthBanner error={error} successMessage={successMessage} />

            <AuthInput
                label={dict.auth?.emailLabel || 'Email Address'}
                icon={Mail}
                type="email"
                placeholder="you@company.com"
                registration={forgotPasswordForm.register('email')}
                error={forgotPasswordForm.formState.errors.email?.message}
            />

            <button type="submit" disabled={loading} className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 transition-all duration-200 mt-6">
                {loading ? <Loader2 className="animate-spin" size={20} /> : (dict.auth?.sendResetCode || 'Send Reset Code')}
            </button>

            <div className="mt-8 pt-8 border-t border-border flex flex-col items-center gap-4">
                <button type="button" onClick={() => switchMode('signin')} className="text-text-muted hover:text-accent text-xs font-black uppercase tracking-widest transition-colors flex items-center gap-2">
                    <ChevronLeft size={16} /> {dict.auth?.backToSignIn || 'Back to Sign In'}
                </button>
            </div>
        </form>
    );
}
