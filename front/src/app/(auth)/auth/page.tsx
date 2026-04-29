"use client";

import React, { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { PLATFORM_NAME } from '@/constants';
import { useAuthFlow } from '@/hooks/auth/useAuthFlow';
import { SignInView } from '@/components/auth/SignInView';
import { RegisterView } from '@/components/auth/RegisterView';
import { ForgotPasswordView } from '@/components/auth/ForgotPasswordView';
import { ResetPasswordView } from '@/components/auth/ResetPasswordView';
import { AuthBrandingPanel } from '@/components/auth/AuthBrandingPanel';
import { AnimatePresence, motion } from 'framer-motion';
import { LanguageToggle } from '@/components/shared/LanguageToggle';
import { ThemeToggle } from '@/components/shared/ThemeToggle';

function AuthForms() {
    const searchParams = useSearchParams();
    const errorParam = searchParams.get('error');

    const {
        mode,
        registerStep,
        setRegisterStep,
        emailCache,
        setEmailCache,
        error,
        setError,
        successMessage,
        setSuccessMessage,
        switchMode
    } = useAuthFlow();

    // Set initial error from URL if present and error state hasn't been set yet
    React.useEffect(() => {
        if (errorParam && !error) {
            setError("Invalid credentials. Please try again.");
        }
    }, [errorParam, error, setError]);

    return (
        <div className="w-full max-w-md">
            <div className="glass-panel p-8 md:p-10 shadow-3xl border-border/50 relative overflow-hidden transition-all duration-300">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent/10 rounded-full -mr-16 -mt-16 blur-2xl pointer-events-none" />
                <div className="absolute bottom-0 left-0 w-24 h-24 bg-primary/10 rounded-full -ml-12 -mb-12 blur-2xl pointer-events-none" />

                <AnimatePresence mode="wait">
                    {mode === 'signin' && (
                        <SignInView 
                            switchMode={switchMode}
                            error={error}
                            setError={setError}
                            successMessage={successMessage}
                        />
                    )}
                    {mode === 'register' && (
                        <RegisterView 
                            switchMode={switchMode}
                            registerStep={registerStep}
                            setRegisterStep={setRegisterStep}
                            emailCache={emailCache}
                            setEmailCache={setEmailCache}
                            error={error}
                            setError={setError}
                            successMessage={successMessage}
                        />
                    )}
                    {mode === 'forgot-password' && (
                        <motion.div
                            key="forgot"
                            initial={{ opacity: 0, x: 20 }}
                            animate={{ opacity: 1, x: 0 }}
                            exit={{ opacity: 0, x: -20 }}
                            transition={{ duration: 0.3 }}
                        >
                            <ForgotPasswordView 
                                switchMode={switchMode}
                                emailCache={emailCache}
                                setEmailCache={setEmailCache}
                                error={error}
                                setError={setError}
                                successMessage={successMessage}
                                setSuccessMessage={setSuccessMessage}
                            />
                        </motion.div>
                    )}
                    {mode === 'reset-password' && (
                        <motion.div
                            key="reset"
                            initial={{ opacity: 0, x: 20 }}
                            animate={{ opacity: 1, x: 0 }}
                            exit={{ opacity: 0, x: -20 }}
                            transition={{ duration: 0.3 }}
                        >
                            <ResetPasswordView 
                                switchMode={switchMode}
                                emailCache={emailCache}
                                setEmailCache={setEmailCache}
                                error={error}
                                setError={setError}
                                successMessage={successMessage}
                                setSuccessMessage={setSuccessMessage}
                            />
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>
            <p className="mt-8 text-center text-[10px] font-black text-text-muted uppercase tracking-[0.3em] opacity-40">
                &copy; {new Date().getFullYear()} {PLATFORM_NAME}
            </p>
        </div>
    );
}

export default function AuthPage() {
    return (
        <main className="min-h-screen flex relative overflow-hidden">
            {/* Left Panel — Branding (desktop only) */}
            <AuthBrandingPanel />

            {/* Right Panel — Auth Forms */}
            <div className="flex-1 flex flex-col min-h-screen relative">
                {/* Background effects for right panel */}
                <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden lg:left-1/2">
                    <div className="absolute inset-0 bg-dot-grid opacity-[0.06]" />
                    <div className="bg-stars-layer opacity-[0.3]" />
                    <div className="absolute top-0 right-0 w-[400px] h-[400px] bg-accent/5 blur-[120px] rounded-full" />
                    <div className="absolute bottom-0 left-0 w-[300px] h-[300px] bg-primary/5 blur-[100px] rounded-full" />
                </div>

                {/* Mobile-only background (full screen on small) */}
                <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden lg:hidden">
                    <div className="absolute inset-0 bg-topography opacity-[0.15]" />
                </div>

                {/* Top bar — Theme/Language toggles */}
                <div className="relative z-10 flex items-center justify-end gap-3 p-6">
                    <LanguageToggle />
                    <ThemeToggle />
                </div>

                {/* Centered form */}
                <div className="relative z-10 flex-1 flex items-center justify-center px-6 pb-12">
                    <Suspense fallback={
                        <div className="glass-panel p-10 w-full max-w-md flex flex-col items-center gap-4 animate-pulse">
                            <div className="w-16 h-16 bg-surface-muted rounded-2xl" />
                            <div className="h-8 w-3/4 bg-surface-muted rounded-lg" />
                        </div>
                    }>
                        <AuthForms />
                    </Suspense>
                </div>
            </div>
        </main>
    );
}
