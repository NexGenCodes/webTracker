'use client';

import React, { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { adminLoginAction } from '@/app/actions/auth';
import { ShieldCheck, Lock, Mail, Loader2, ArrowRight } from 'lucide-react';
import { motion } from 'framer-motion';

export default function AdminLoginPage() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [isPending, startTransition] = useTransition();
    const router = useRouter();

    const handleLogin = (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        startTransition(async () => {
            const result = await adminLoginAction({ email, password });
            if (result.success) {
                router.push('/super-admin');
            } else {
                setError(result.error || 'Invalid credentials');
            }
        });
    };

    return (
        <div className="min-h-screen bg-surface flex items-center justify-center p-4">
            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="w-full max-w-md bg-surface border border-border rounded-3xl p-8 shadow-xl"
            >
                <div className="flex flex-col items-center mb-8">
                    <div className="bg-accent/10 p-4 rounded-full mb-4">
                        <ShieldCheck className="w-10 h-10 text-accent" />
                    </div>
                    <h1 className="text-2xl font-black uppercase tracking-widest text-text-main">System Admin</h1>
                    <p className="text-sm text-text-muted mt-2 font-medium">Authorized personnel only</p>
                </div>

                {error && (
                    <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-xl text-error text-sm font-bold text-center">
                        {error}
                    </div>
                )}

                <form onSubmit={handleLogin} className="space-y-5">
                    <div>
                        <label className="block text-xs font-black uppercase tracking-widest text-text-muted mb-2">
                            Admin Email
                        </label>
                        <div className="relative">
                            <Mail className="absolute left-4 top-1/2 -translate-y-1/2 text-text-muted w-5 h-5" />
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="w-full bg-surface-muted border-2 border-border focus:border-accent rounded-xl py-3.5 pl-12 pr-4 text-text-main font-medium outline-none transition-all"
                                placeholder="admin@cargohive.com"
                                required
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-xs font-black uppercase tracking-widest text-text-muted mb-2">
                            Passphrase
                        </label>
                        <div className="relative">
                            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-text-muted w-5 h-5" />
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full bg-surface-muted border-2 border-border focus:border-accent rounded-xl py-3.5 pl-12 pr-4 text-text-main font-medium outline-none transition-all"
                                placeholder="••••••••••••"
                                required
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isPending}
                        className="w-full mt-6 bg-accent hover:bg-accent-hover text-white font-black uppercase tracking-widest py-4 rounded-xl flex items-center justify-center gap-2 transition-all active:scale-[0.98] disabled:opacity-50"
                    >
                        {isPending ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                Authenticate <ArrowRight className="w-5 h-5" />
                            </>
                        )}
                    </button>
                </form>
            </motion.div>
        </div>
    );
}
