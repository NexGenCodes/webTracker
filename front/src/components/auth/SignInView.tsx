import React from 'react';
import Link from 'next/link';
import { Package, Lock, Mail, Loader2, ChevronLeft } from 'lucide-react';
import { motion } from 'framer-motion';
import { AuthMode } from '@/lib/validations/auth';
import { useSignIn } from '@/hooks/auth/useSignIn';
import { AuthInput } from './ui/AuthInput';
import { AuthBanner } from './ui/AuthBanner';

interface SignInViewProps {
    switchMode: (mode: AuthMode) => void;
    error: string | null;
    setError: (error: string | null) => void;
    successMessage: string | null;
}

    export function SignInView({ switchMode, error, setError, successMessage }: SignInViewProps) {
    const { form, onSubmit, loading, handleGoogleSignIn } = useSignIn(setError);

    return (
        <motion.form
            onSubmit={form.handleSubmit(onSubmit)}
            className="space-y-6 w-full"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0, ...(error ? { x: [0, -8, 8, -4, 4, 0] } : {}) }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            key="signin"
        >
            <div className="flex flex-col items-center mb-10 text-center">
                <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6">
                    <Package className="text-white" size={32} />
                </div>
                <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">
                    Welcome Back
                </h1>
                <p className="text-text-muted font-bold text-sm uppercase tracking-widest opacity-70">
                    Sign in to your account
                </p>
            </div>
            
            <AuthBanner error={error} successMessage={successMessage} />

            {/* Google OAuth */}
            <button
                type="button"
                onClick={handleGoogleSignIn}
                disabled={loading}
                className="w-full flex items-center justify-center gap-3 py-3.5 px-6 rounded-2xl border-2 border-border bg-surface hover:bg-surface-muted transition-all duration-200 active:scale-[0.98] group disabled:opacity-50"
            >
                <svg className="w-5 h-5" viewBox="0 0 24 24">
                    <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
                    <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                    <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                    <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                </svg>
                <span className="text-xs font-black uppercase tracking-widest text-text-main">
                    Continue with Google
                </span>
            </button>

            <div className="flex items-center gap-4 w-full">
                <div className="flex-1 h-px bg-border" />
                <span className="text-[10px] font-black uppercase tracking-[0.3em] text-text-muted/50">or</span>
                <div className="flex-1 h-px bg-border" />
            </div>

            <AuthInput
                label="Email Address"
                icon={Mail}
                type="email"
                placeholder="you@company.com"
                registration={form.register('email')}
                error={form.formState.errors.email?.message}
            />

            <AuthInput
                label="Password"
                icon={Lock}
                type="password"
                placeholder="••••••••"
                registration={form.register('password')}
                error={form.formState.errors.password?.message}
                actionLabel="Forgot?"
                onActionClick={() => switchMode('forgot-password')}
            />

            <button type="submit" disabled={loading} className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 transition-all duration-200">
                {loading ? <Loader2 className="animate-spin" size={20} /> : "Sign In"}
            </button>

            {/* Terms & Privacy */}
            <p className="text-center text-[10px] text-text-muted/60 font-medium leading-relaxed">
                By signing in, you agree to our{' '}
                <Link href="/terms" className="text-accent hover:underline font-bold">Terms of Service</Link>
                {' '}and{' '}
                <Link href="/privacy" className="text-accent hover:underline font-bold">Privacy Policy</Link>
            </p>

            <div className="mt-6 pt-6 border-t border-border flex flex-col items-center gap-4">
                <button type="button" onClick={() => switchMode('register')} className="text-accent hover:text-accent/80 text-xs font-black uppercase tracking-widest transition-colors">
                    Don&apos;t have an account? Register
                </button>
                <Link href="/" className="flex items-center gap-2 text-text-muted hover:text-accent transition-colors text-xs font-black uppercase tracking-widest">
                    <ChevronLeft size={16} /> Back to Home
                </Link>
            </div>
        </motion.form>
    );
}
