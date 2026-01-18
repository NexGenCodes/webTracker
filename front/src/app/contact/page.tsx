"use client";

import { useI18n } from "@/components/I18nContext";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import { PremiumInput, PremiumTextarea } from "@/components/PremiumInput";
import { IconInfoItem } from "@/components/IconInfoItem";
import { Mail, Phone, MapPin, Send, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { APP_NAME, CONTACT_EMAIL, CONTACT_PHONE, CONTACT_HQ } from "@/lib/constants";
import { useState, useRef } from "react";
import { submitContactMessage } from "@/app/actions/contact";
import toast, { Toaster } from 'react-hot-toast';

export default function ContactPage() {
    const { dict } = useI18n();
    const [loading, setLoading] = useState(false);
    const [submitted, setSubmitted] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const formRef = useRef<HTMLFormElement>(null);

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setLoading(true);

        const formData = new FormData(e.currentTarget);
        const data = {
            name: formData.get('name') as string,
            email: formData.get('email') as string,
            message: formData.get('message') as string,
        };

        const result = await submitContactMessage(data);

        if (result.success) {
            setSubmitted(true);
            setError(null);
            toast.success(dict.contact.success + '! ' + dict.contact.successDesc, {
                duration: 5000,
                position: 'top-center',
                style: {
                    background: 'hsl(var(--success))',
                    color: 'white',
                    fontWeight: 'bold',
                    padding: '16px 24px',
                    borderRadius: '12px',
                },
            });
            // Reset form
            formRef.current?.reset();
        } else {
            setError(result.error || 'Failed to send message');
            toast.error(result.error || 'Failed to send message', {
                duration: 4000,
                position: 'top-center',
                style: {
                    background: '#ef4444',
                    color: 'white',
                    fontWeight: 'bold',
                    padding: '16px 24px',
                    borderRadius: '12px',
                },
            });
        }

        setLoading(false);
    };

    return (
        <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative">
            <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden">
                <div className="absolute inset-0 bg-dot-grid opacity-[0.1]" />
                <div className="bg-stars-layer opacity-[0.3]" />
                <div className="shooting-star" style={{ top: '15%', left: '85%', animationDelay: '4s' }} />
                <div className="shooting-star" style={{ top: '45%', left: '20%', animationDelay: '12s' }} />
                <div className="absolute inset-0 bg-topography opacity-[0.15]" />
            </div>
            <div className="w-full max-w-7xl flex flex-col flex-1 px-6 pt-32 md:pt-40 pb-20 relative z-10">
                <Header showNav={true} />

                <section className="glass-panel overflow-hidden md:flex animate-fade-in shadow-3xl border-border/50">
                    <div className="bg-accent/5 md:w-1/3 p-10 md:p-14 border-b md:border-b-0 md:border-r border-border backdrop-blur-md flex flex-col">
                        <Link href="/" className="inline-flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.3em] text-text-muted hover:text-accent mb-16 transition-all hover:-translate-x-1 group">
                            <ArrowLeft size={14} strokeWidth={3} className="group-hover:-translate-x-1 transition-transform" />
                            {dict.common.home}
                        </Link>

                        <h1 className="text-5xl font-black mb-8 text-gradient leading-[0.85] tracking-tighter uppercase">{dict.contact.title}</h1>
                        <p className="text-text-muted mb-16 font-bold text-lg leading-relaxed border-l-2 border-accent/20 pl-6">{dict.contact.subtitle}</p>

                        <div className="space-y-10 mt-auto">
                            <IconInfoItem icon={Mail} label={dict.contact.emailLabel} value={CONTACT_EMAIL} />
                            <IconInfoItem icon={Phone} label={dict.contact.phoneLabel} value={CONTACT_PHONE} />
                            <IconInfoItem icon={MapPin} label={dict.contact.hqLabel} value={CONTACT_HQ} />
                        </div>
                    </div>

                    <div className="p-10 md:p-16 flex-1 flex flex-col justify-center bg-white/5 backdrop-blur-sm">
                        {submitted ? (
                            <div className="h-full min-h-[500px] flex flex-col items-center justify-center text-center space-y-8 animate-scale-in">
                                <div className="relative">
                                    <div className="absolute inset-0 bg-success blur-3xl opacity-20 animate-pulse" />
                                    <div className="relative w-24 h-24 bg-success/10 rounded-[2.5rem] flex items-center justify-center text-success shadow-inner rotate-3 transition-transform hover:rotate-0 duration-500">
                                        <Send size={40} strokeWidth={2.5} />
                                    </div>
                                </div>
                                <div className="max-w-md">
                                    <h3 className="text-4xl font-black text-text-main mb-4 tracking-tighter uppercase">{dict.contact.success}</h3>
                                    <p className="text-text-muted font-bold text-lg leading-relaxed">{dict.contact.successDesc}</p>
                                </div>
                            </div>
                        ) : (
                            <form onSubmit={handleSubmit} className="space-y-10">
                                {error && (
                                    <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
                                        <p className="text-red-500 text-sm font-semibold">{error}</p>
                                    </div>
                                )}
                                <PremiumInput name="name" label={dict.contact.name} type="text" required />
                                <PremiumInput name="email" label={dict.contact.email} type="email" required />
                                <PremiumTextarea name="message" label={dict.contact.message} required />
                                <button
                                    type="submit"
                                    disabled={loading}
                                    className="btn-primary w-full flex items-center justify-center gap-4 py-6 text-lg shadow-2xl shadow-accent/20 group hover:scale-[1.01] active:scale-[0.99] transition-all border-none! disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    <Send size={20} strokeWidth={3} className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
                                    <span className="font-black uppercase tracking-[0.2em]">
                                        {loading ? 'Sending...' : dict.contact.send}
                                    </span>
                                </button>
                            </form>
                        )}
                    </div>
                </section>

                <Footer minimal={true} />
            </div>
        </main>
    );
}
