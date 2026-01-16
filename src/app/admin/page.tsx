"use client";

import React, { useState } from 'react';
import { createShipment } from '../actions/shipment';
import { parseEmail } from '@/lib/email-parser';
import { Copy, Check, ChevronLeft, Package } from 'lucide-react';
import { useI18n } from '@/components/I18nContext';
import { ThemeToggle } from '@/components/ThemeToggle';
import { LanguageToggle } from '@/components/LanguageToggle';
import Link from 'next/link';
import { APP_NAME } from '@/lib/constants';

export default function AdminPage() {
    const { dict } = useI18n();
    const [emailText, setEmailText] = useState('');
    const [trackingId, setTrackingId] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [copied, setCopied] = useState(false);
    const [password, setPassword] = useState('');
    const [isAuthenticated, setIsAuthenticated] = useState(false);

    const handleLogin = (e: React.FormEvent) => {
        e.preventDefault();
        if (password === 'admin123') {
            setIsAuthenticated(true);
            setError(null);
        } else {
            setError(dict.admin.wrongPassword);
        }
    };

    const handleGenerate = async () => {
        setError(null);
        setLoading(true);
        try {
            const dto = parseEmail(emailText);
            const result = await createShipment(dto);
            if (result.success) {
                setTrackingId(result.trackingNumber ?? null);
            } else {
                setError(result.error ?? dict.admin.failedCreate);
            }
        } catch (err: any) {
            setError(err.message || dict.admin.invalidEmail);
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = async () => {
        if (trackingId) {
            await navigator.clipboard.writeText(trackingId);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    const handleBack = () => {
        setTrackingId(null);
        setEmailText('');
        setCopied(false);
    };

    if (!isAuthenticated) {
        return (
            <div className="min-h-screen flex flex-col items-center p-4">
                <header className="w-full max-w-5xl py-6 flex justify-between items-center mb-20">
                    <Link href="/" className="flex items-center gap-3 font-extrabold text-2xl tracking-tighter">
                        <div className="bg-accent p-2 rounded-xl">
                            <Package className="text-white" size={24} />
                        </div>
                        <span className="text-gradient uppercase">{APP_NAME}</span>
                    </Link>
                    <div className="hidden md:flex items-center gap-8 mr-12 text-sm font-medium text-gray-400">
                        <Link href="/about" className="hover:text-accent transition-colors">{dict.common.about}</Link>
                        <Link href="/contact" className="hover:text-accent transition-colors">{dict.common.contact}</Link>
                    </div>
                    <div className="flex items-center gap-4">
                        <LanguageToggle />
                        <ThemeToggle />
                    </div>
                </header>

                <div className="max-w-md w-full p-8 glass-panel animate-fade-in mt-12">
                    <h1 className="text-2xl font-black text-text-main mb-6 text-center">{dict.admin.loginTitle}</h1>
                    <form onSubmit={handleLogin} className="space-y-4">
                        <input
                            type="password"
                            placeholder={dict.admin.loginPlaceholder}
                            className="w-full bg-surface-muted border border-border p-3 rounded-xl text-text-main focus:border-accent outline-none transition-all"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                        />
                        {error && <p className="text-error text-sm text-center font-bold tracking-tight">{error}</p>}
                        <button type="submit" className="btn-primary w-full py-4 text-lg uppercase tracking-widest">
                            {dict.admin.loginButton}
                        </button>
                    </form>
                </div>
            </div>
        );
    }

    if (trackingId) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center animate-fade-in space-y-8">
                <div className="space-y-2">
                    <h2 className="text-3xl font-black text-success tracking-tight uppercase">{dict.admin.success}</h2>
                    <p className="text-text-muted font-medium">{dict.admin.successDesc}</p>
                </div>

                <div className="w-full max-w-md bg-surface border border-border rounded-2xl p-8 shadow-xl flex flex-col items-center gap-4 relative overflow-hidden">
                    <div className="absolute top-0 right-0 w-32 h-32 bg-success/5 rounded-full -mr-16 -mt-16 blur-2xl" />
                    <span className="text-xs text-text-muted uppercase tracking-widest font-black">{dict.shipment.trackingId}</span>
                    <span className="text-4xl font-mono text-text-main font-black tracking-widest break-all">
                        {trackingId}
                    </span>
                </div>

                <div className="flex flex-col w-full max-w-sm gap-4">
                    <button
                        onClick={handleCopy}
                        className="btn-primary flex items-center justify-center gap-2 py-4 text-lg w-full"
                    >
                        {copied ? <Check /> : <Copy />}
                        {copied ? dict.admin.copied : dict.admin.copy}
                    </button>
                    <button
                        onClick={handleBack}
                        className="flex items-center justify-center gap-2 text-gray-400 hover:text-white py-2 transition-colors"
                    >
                        <ChevronLeft size={20} />
                        {dict.admin.createAnother}
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-xl mx-auto p-4 flex flex-col gap-6 py-12">
            <header className="flex justify-between items-center mb-8">
                <Link href="/" className="flex items-center gap-3 font-extrabold text-2xl tracking-tighter">
                    <div className="bg-accent p-2 rounded-xl shadow-lg shadow-accent/20">
                        <Package className="text-white" size={24} />
                    </div>
                    <span className="text-gradient uppercase">{APP_NAME}</span>
                </Link>
                <div className="hidden md:flex items-center gap-8 mr-12 text-sm font-medium text-gray-400">
                    <Link href="/about" className="hover:text-accent transition-colors">{dict.common.about}</Link>
                    <Link href="/contact" className="hover:text-accent transition-colors">{dict.common.contact}</Link>
                </div>
                <div className="flex items-center gap-4">
                    <LanguageToggle />
                    <ThemeToggle />
                </div>
            </header>

            <div className="space-y-4">
                <div className="bg-surface-muted p-6 rounded-2xl border border-border text-sm text-text-muted font-medium">
                    <p className="font-black mb-2 text-text-main uppercase tracking-widest text-xs">{dict.admin.instructions}</p>
                    <ul className="list-disc pl-4 space-y-2">
                        <li>{dict.admin.step1}</li>
                        <li>{dict.admin.step2}</li>
                        <li>{dict.admin.step3}</li>
                    </ul>
                </div>

                <textarea
                    className="w-full h-64 bg-surface text-text-main p-6 rounded-2xl border border-border focus:border-accent outline-none resize-none font-mono text-sm transition-all shadow-inner"
                    placeholder={dict.admin.placeholder}
                    value={emailText}
                    onChange={(e) => setEmailText(e.target.value)}
                />

                {error && (
                    <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-300 text-sm">
                        {error}
                    </div>
                )}

                <button
                    disabled={loading || !emailText.trim()}
                    onClick={handleGenerate}
                    className="btn-primary w-full py-4 text-lg flex items-center justify-center gap-2"
                >
                    {loading ? dict.common.loading : dict.admin.generate}
                </button>

                <Link href="/" className="flex items-center justify-center gap-2 text-gray-400 hover:text-white py-2 transition-colors">
                    <ChevronLeft size={20} />
                    {dict.common.home}
                </Link>
            </div>
        </div>
    );
}
