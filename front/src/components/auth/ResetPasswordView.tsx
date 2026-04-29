import React, { useState } from 'react';
import { Key, Lock, Loader2, ChevronLeft } from 'lucide-react';
import { AuthMode } from '@/lib/validations/auth';
import { usePasswordReset } from '@/hooks/auth/usePasswordReset';
import { AuthInput } from './ui/AuthInput';
import { OTPInput } from './ui/OTPInput';
import { AuthBanner } from './ui/AuthBanner';

interface ResetPasswordViewProps {
    switchMode: (mode: AuthMode) => void;
    emailCache: string;
    setEmailCache: (email: string) => void;
    error: string | null;
    setError: (error: string | null) => void;
    successMessage: string | null;
    setSuccessMessage: (msg: string | null) => void;
}

export function ResetPasswordView({ 
    switchMode, 
    emailCache, 
    setEmailCache, 
    error, 
    setError, 
    successMessage, 
    setSuccessMessage 
}: ResetPasswordViewProps) {
    const [otp, setOtp] = useState(['', '', '', '', '', '']);
    const { resetPasswordForm, onResetPassword, loading } = usePasswordReset(
        setError,
        setSuccessMessage,
        setEmailCache,
        switchMode,
        emailCache
    );

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const code = otp.join('');
        await resetPasswordForm.handleSubmit((data) => onResetPassword(data, code))(e);
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-6 animate-fade-in w-full">
            <div className="flex flex-col items-center mb-10 text-center">
                <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6">
                    <Key className="text-white" size={32} />
                </div>
                <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">
                    New Password
                </h1>
                <p className="text-text-muted font-bold text-sm uppercase tracking-widest opacity-70">
                    Enter the code and your new password
                </p>
            </div>

            <AuthBanner error={error} successMessage={successMessage} />

            <div className="space-y-4">
                <label className="text-[10px] font-black uppercase tracking-[0.2em] text-accent/80 ml-1 text-center block">
                    6-Digit Reset Code
                </label>
                <OTPInput otp={otp} setOtp={setOtp} />
            </div>

            <div className="mt-6 space-y-6">
                <AuthInput
                    label="New Password"
                    icon={Lock}
                    type="password"
                    placeholder="Min 8 characters"
                    registration={resetPasswordForm.register('password')}
                    error={resetPasswordForm.formState.errors.password?.message}
                />

                <AuthInput
                    label="Confirm New Password"
                    icon={Lock}
                    type="password"
                    placeholder="••••••••"
                    registration={resetPasswordForm.register('confirmPassword')}
                    error={resetPasswordForm.formState.errors.confirmPassword?.message}
                />
            </div>

            <button type="submit" disabled={loading || otp.join('').length < 6} className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 transition-all duration-200 mt-6">
                {loading ? <Loader2 className="animate-spin" size={20} /> : "Update Password"}
            </button>

            <div className="mt-8 pt-8 border-t border-border flex flex-col items-center gap-4">
                <button type="button" onClick={() => switchMode('signin')} className="text-text-muted hover:text-accent text-xs font-black uppercase tracking-widest transition-colors flex items-center gap-2">
                    <ChevronLeft size={16} /> Back to Sign In
                </button>
            </div>
        </form>
    );
}
