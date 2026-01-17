"use client";

import React, { useState, Suspense } from 'react';
import { signIn } from "next-auth/react";
import { useSearchParams, useRouter } from 'next/navigation';
import { Package, Lock, User, AlertCircle, ChevronLeft, Loader2 } from 'lucide-react';
import Link from 'next/link';
import { APP_NAME } from '@/lib/constants';
import { useI18n } from '@/components/I18nContext';
import { ThemeToggle } from '@/components/ThemeToggle';
import { Header } from '@/components/Header';

function SignInForm() {
    const { dict } = useI18n();
    const router = useRouter();
    const searchParams = useSearchParams();
    const callbackUrl = searchParams.get('callbackUrl') || '/admin';
    const errorParam = searchParams.get('error');

    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(errorParam ? "Invalid credentials. Please try again." : null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        try {
            const result = await signIn("credentials", {
                username,
                password,
                redirect: false,
                callbackUrl,
            });

            if (result?.error) {
                setError("Invalid username or password. Please check your credentials.");
            } else {
                router.push(callbackUrl);
            }
        } catch (err) {
            setError("Something went wrong. Please try again later.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="w-full max-w-md animate-scale-in">
            <div className="glass-panel p-8 md:p-10 shadow-3xl border-border/50 relative overflow-hidden">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent/10 rounded-full -mr-16 -mt-16 blur-2xl pointer-events-none" />

                <div className="flex flex-col items-center mb-10 text-center">
                    <div className="bg-accent p-3 rounded-2xl shadow-lg shadow-accent/20 mb-6">
                        <Package className="text-white" size={32} />
                    </div>
                    <h1 className="text-3xl font-black text-text-main tracking-tighter uppercase mb-2">
                        {dict.admin.loginTitle || "Admin Access"}
                    </h1>
                    <p className="text-text-muted font-bold text-sm uppercase tracking-widest opacity-70">
                        Secure Logistics Portal
                    </p>
                </div>

                {error && (
                    <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-3 text-error text-sm animate-fade-in">
                        <AlertCircle size={18} className="shrink-0" />
                        <p className="font-bold">{error}</p>
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                    <div className="space-y-2">
                        <label className="text-[10px] font-black uppercase tracking-[0.2em] text-accent/80 ml-1">
                            Username
                        </label>
                        <div className="relative group">
                            <User className="absolute left-5 top-1/2 -translate-y-1/2 text-text-muted group-focus-within:text-accent transition-colors" size={20} />
                            <input
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                className="input-premium pl-12!"
                                placeholder="Enter admin username"
                                required
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-[10px] font-black uppercase tracking-[0.2em] text-accent/80 ml-1">
                            Password
                        </label>
                        <div className="relative group">
                            <Lock className="absolute left-5 top-1/2 -translate-y-1/2 text-text-muted group-focus-within:text-accent transition-colors" size={20} />
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="input-premium pl-12!"
                                placeholder="••••••••"
                                required
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="btn-primary w-full py-4 text-base flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50"
                    >
                        {loading ? (
                            <Loader2 className="animate-spin" size={20} />
                        ) : (
                            "Authenticate"
                        )}
                    </button>
                </form>

                <div className="mt-8 pt-8 border-t border-border flex justify-center">
                    <Link href="/" className="flex items-center gap-2 text-text-muted hover:text-accent transition-colors text-xs font-black uppercase tracking-widest">
                        <ChevronLeft size={16} />
                        Back to Home
                    </Link>
                </div>
            </div>

            <p className="mt-8 text-center text-[10px] font-black text-text-muted uppercase tracking-[0.3em] opacity-40">
                &copy; {new Date().getFullYear()} {APP_NAME} Internal Systems
            </p>
        </div>
    );
}

export default function SignIn() {
    return (
        <main className="min-h-screen flex flex-col items-center justify-center p-6 relative overflow-hidden">
            {/* Background Flair - Reusing from Home */}
            <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden">
                <div className="absolute inset-0 bg-dot-grid opacity-[0.1]" />
                <div className="bg-stars-layer opacity-[0.4]" />
                <div className="absolute inset-0 bg-topography opacity-[0.2]" />
                <div className="shooting-star" style={{ top: '15%', left: '70%', animationDelay: '3s' }} />
                <div className="shooting-star" style={{ top: '60%', left: '20%', animationDelay: '8s' }} />
                <div className="absolute top-0 right-0 w-[600px] h-[600px] bg-accent/5 blur-[120px] rounded-full" />
                <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-primary/5 blur-[100px] rounded-full" />
            </div>

            <Header showNav={false} />

            <div className="relative z-10 w-full flex justify-center mt-32 md:mt-24">
                <Suspense fallback={
                    <div className="glass-panel p-10 w-full max-w-md flex flex-col items-center gap-4 animate-pulse">
                        <div className="w-16 h-16 bg-surface-muted rounded-2xl" />
                        <div className="h-8 w-3/4 bg-surface-muted rounded-lg" />
                    </div>
                }>
                    <SignInForm />
                </Suspense>
            </div>
        </main>
    );
}
