import React, { useState } from 'react';
import Link from 'next/link';
import { useI18n } from '@/components/providers/I18nContext';
import { Package, Lock, Mail, Building2, Loader2, ChevronLeft, ArrowRight, ShieldCheck, CheckCircle2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { AuthMode, RegisterStep } from '@/lib/validations/auth';
import { useRegister } from '@/hooks/auth/useRegister';
import { AuthInput } from './ui/AuthInput';
import { OTPInput } from './ui/OTPInput';
import { AuthBanner } from './ui/AuthBanner';
import { PasswordStrength } from './ui/PasswordStrength';

interface RegisterViewProps {
    switchMode: (mode: AuthMode) => void;
    registerStep: RegisterStep;
    setRegisterStep: (step: RegisterStep) => void;
    emailCache: string;
    setEmailCache: (email: string) => void;
    error: string | null;
    setError: (error: string | null) => void;
    successMessage: string | null;
}

    export function RegisterView({ 
    switchMode, 
    registerStep, 
    setRegisterStep, 
    emailCache, 
    setEmailCache, 
    error, 
    setError, 
    successMessage 
}: RegisterViewProps) {
    const { form, onRegisterIntent, verifyOtp, loading, otpTimer, handleGoogleSignIn } = useRegister(setError, setEmailCache, setRegisterStep, registerStep);
    const [otp, setOtp] = useState(['', '', '', '', '', '']);
    const { dict } = useI18n();

    const formatTime = (seconds: number) => {
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${s < 10 ? '0' : ''}${s}`;
    };

    const handleVerifyRegisterOTP = async (e: React.FormEvent) => {
        e.preventDefault();
        await verifyOtp(otp.join(''));
    };

    const stepsInfo: { key: RegisterStep; label: string }[] = [
        { key: 'credentials', label: dict.auth?.stepAccount || 'Account' },
        { key: 'otp', label: dict.auth?.stepVerify || 'Verify' },
    ];
    const currentIdx = stepsInfo.findIndex((s) => s.key === registerStep);

    const watchedPassword = form.watch('password');

    return (
        <motion.div
            className="w-full"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            key="register"
        >
            {/* Step Progress */}
            <div className="flex items-center justify-center gap-2 mb-8">
                {stepsInfo.map((s, i) => (
                    <React.Fragment key={s.key}>
                        <div className="flex flex-col items-center gap-1">
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-black transition-all duration-300 ${
                                i < currentIdx ? 'bg-success text-white' : i === currentIdx ? 'bg-accent text-white shadow-lg shadow-accent/30' : 'bg-surface-muted text-text-muted'
                            }`}>
                                {i < currentIdx ? <CheckCircle2 size={14} /> : i + 1}
                            </div>
                            <span className={`text-[9px] font-black uppercase tracking-widest ${i === currentIdx ? 'text-accent' : 'text-text-muted opacity-50'}`}>
                                {s.label}
                            </span>
                        </div>
                        {i < stepsInfo.length - 1 && (
                            <div className={`w-12 h-0.5 rounded-full transition-all duration-300 mb-4 ${i < currentIdx ? 'bg-success' : 'bg-border'}`} />
                        )}
                    </React.Fragment>
                ))}
            </div>

            <AuthBanner error={error} successMessage={successMessage} />

            {/* Step 1: Credentials */}
            {registerStep === 'credentials' && (
                <form onSubmit={form.handleSubmit(onRegisterIntent)} className="space-y-5">
                    <div className="flex flex-col items-center mb-8 text-center">
                        <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6">
                            <Package className="text-white" size={32} />
                        </div>
                        <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">{dict.auth?.getStarted || 'Get Started'}</h1>
                        <p className="text-text-muted font-bold text-sm uppercase tracking-widest opacity-70">{dict.auth?.createAccount || 'Create your account'}</p>
                    </div>

                    {/* Google OAuth */}
                    <button
                        type="button"
                        onClick={handleGoogleSignIn}
                        disabled={loading}
                        className="w-full flex items-center justify-center gap-3 py-3.5 px-6 rounded-2xl border-2 border-border bg-surface hover:bg-surface-muted transition-all duration-200 active:scale-[0.98] disabled:opacity-50"
                    >
                        <svg className="w-5 h-5" viewBox="0 0 24 24">
                            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
                            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                        </svg>
                        <span className="text-xs font-black uppercase tracking-widest text-text-main">
                            {dict.auth?.continueGoogle || 'Continue with Google'}
                        </span>
                    </button>

                    <div className="flex items-center gap-4 w-full">
                        <div className="flex-1 h-px bg-border" />
                        <span className="text-[10px] font-black uppercase tracking-[0.3em] text-text-muted/50">{dict.common?.or || 'or'}</span>
                        <div className="flex-1 h-px bg-border" />
                    </div>

                    <AuthInput
                        label={dict.auth?.companyLabel || 'Company Name'}
                        icon={Building2}
                        type="text"
                        placeholder={dict.auth?.companyPlaceholder || 'Fast Track Logistics'}
                        registration={form.register('companyName')}
                        error={form.formState.errors.companyName?.message}
                    />

                    <AuthInput
                        label={dict.auth?.emailLabel || 'Email Address'}
                        icon={Mail}
                        type="email"
                        placeholder="you@company.com"
                        registration={form.register('email')}
                        error={form.formState.errors.email?.message}
                    />

                    <AuthInput
                        label={dict.auth?.passwordLabel || 'Password'}
                        icon={Lock}
                        type="password"
                        placeholder={dict.auth?.minChars || 'Min 8 characters'}
                        registration={form.register('password')}
                        error={form.formState.errors.password?.message}
                    />

                    <PasswordStrength password={watchedPassword} />

                    <AuthInput
                        label={dict.auth?.confirmPasswordLabel || 'Confirm Password'}
                        icon={Lock}
                        type="password"
                        placeholder="••••••••"
                        registration={form.register('confirmPassword')}
                        error={form.formState.errors.confirmPassword?.message}
                    />

                    <button type="submit" disabled={loading} className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 transition-all duration-200 mt-6">
                        {loading ? <Loader2 className="animate-spin" size={20} /> : <>{dict.auth?.continue || 'Continue'} <ArrowRight size={18} /></>}
                    </button>

                    {/* Terms & Privacy */}
                    <p className="text-center text-[10px] text-text-muted/60 font-medium leading-relaxed">
                        {dict.auth?.termsAgreeRegister || 'By creating an account, you agree to our'}{' '}
                        <Link href="/terms" className="text-accent hover:underline font-bold">{dict.common?.terms || 'Terms of Service'}</Link>
                        {' '}{dict.auth?.and || 'and'}{' '}
                        <Link href="/privacy" className="text-accent hover:underline font-bold">{dict.common?.privacy || 'Privacy Policy'}</Link>
                    </p>
                </form>
            )}

            {/* Step 2: OTP */}
            {registerStep === 'otp' && (
                <form onSubmit={handleVerifyRegisterOTP} className="space-y-8">
                    <div className="flex flex-col items-center mb-8 text-center">
                        <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6"><ShieldCheck className="text-white" size={32} /></div>
                        <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">{dict.auth?.verifyEmail || 'Verify Email'}</h1>
                        <p className="text-text-muted font-bold text-sm tracking-wide opacity-70">{dict.auth?.otpSent || 'We sent a 6-digit code to'}</p>
                        <p className="text-accent font-black text-sm mt-1">{emailCache}</p>
                    </div>

                    <OTPInput otp={otp} setOtp={setOtp} />

                    <button type="submit" disabled={loading || otp.join('').length < 6} className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 transition-all duration-200">
                        {loading ? <Loader2 className="animate-spin" size={20} /> : <>{dict.auth?.verifyContinue || 'Verify & Continue'} <ArrowRight size={18} /></>}
                    </button>

                    <p className="text-xs text-text-muted text-center flex flex-col gap-2 items-center">
                        {otpTimer > 0 ? (
                            <span>{dict.auth?.codeExpires || 'Code expires in'} <span className="font-bold text-accent">{formatTime(otpTimer)}</span></span>
                        ) : (
                            <span className="text-error font-bold">{dict.auth?.codeExpired || 'Code expired'}</span>
                        )}
                        <span>
                            {dict.auth?.noCode || "Didn't get the code?"}{' '}
                            <button 
                                type="button" 
                                className="text-accent font-bold hover:underline disabled:opacity-50" 
                                onClick={() => { 
                                    setOtp(['', '', '', '', '', '']); 
                                    setError(null); 
                                    const values = form.getValues();
                                    if (values.email && values.password && values.companyName) {
                                        onRegisterIntent(values);
                                    }
                                }}
                            >
                                {dict.auth?.resend || 'Resend'}
                            </button>
                        </span>
                    </p>
                </form>
            )}

            <div className="mt-8 pt-8 border-t border-border flex flex-col items-center gap-4">
                {registerStep === 'credentials' && (
                    <button type="button" onClick={() => switchMode('signin')} className="text-accent hover:text-accent/80 text-xs font-black uppercase tracking-widest transition-colors">
                        {dict.auth?.hasAccount || 'Already have an account? Sign In'}
                    </button>
                )}
                <Link href="/" className="flex items-center gap-2 text-text-muted hover:text-accent transition-colors text-xs font-black uppercase tracking-widest">
                    <ChevronLeft size={16} /> {dict.common?.backToHome || 'Back to Home'}
                </Link>
            </div>
        </motion.div>
    );
}
