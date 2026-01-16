"use client";

import { useI18n } from "@/components/I18nContext";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import { FeatureCard } from "@/components/FeatureCard";
import { Shield, Activity, ArrowLeft, ArrowRight } from "lucide-react";
import Link from "next/link";

export default function AboutPage() {
    const { dict } = useI18n();

    return (
        <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative">
            <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden">
                <div className="absolute inset-0 bg-dot-grid opacity-[0.1]" />
                <div className="bg-stars-layer opacity-[0.3]" />
                <div className="shooting-star" style={{ top: '20%', left: '70%', animationDelay: '1s' }} />
                <div className="shooting-star" style={{ top: '60%', left: '30%', animationDelay: '9s' }} />
                <div className="absolute inset-0 bg-topography opacity-[0.15]" />
            </div>
            <div className="w-full max-w-7xl flex flex-col flex-1 px-6 pt-32 md:pt-40 relative z-10">

                <Header />

                <section className="animate-fade-in relative pb-10 md:pb-24">
                    <div className="absolute -top-20 -left-20 w-80 h-80 bg-accent/10 rounded-full blur-3xl pointer-events-none animate-pulse" />
                    <div className="absolute top-40 -right-20 w-80 h-80 bg-accent/5 rounded-full blur-3xl pointer-events-none" />

                    <div className="max-w-4xl relative z-10 mb-20">
                        <Link href="/" className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-surface-muted text-[10px] font-black uppercase tracking-widest text-accent mb-12 hover:bg-accent/10 transition-all hover:-translate-x-1 shadow-sm border border-border/50">
                            <ArrowLeft size={14} strokeWidth={3} />
                            {dict.common.home}
                        </Link>

                        <h1 className="text-5xl md:text-8xl font-black mb-6 tracking-tighter text-gradient leading-tight lg:leading-[0.85] uppercase">
                            {dict.about.title}
                        </h1>
                        <p className="text-lg md:text-xl text-accent font-black mb-10 uppercase tracking-[0.3em] flex items-center gap-3">
                            <span className="h-px w-12 bg-accent opacity-30" />
                            {dict.about.subtitle}
                        </p>
                        <p className="text-xl md:text-2xl text-text-main font-bold leading-relaxed max-w-2xl border-l-4 border-accent pl-8 py-2">
                            {dict.about.description}
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-10 mb-32">
                        <FeatureCard
                            icon={<Shield size={32} strokeWidth={2.5} />}
                            title={dict.about.mission}
                            description={dict.about.missionDesc}
                            className="bg-accent/5"
                        />
                        <FeatureCard
                            icon={<Activity size={32} strokeWidth={2.5} />}
                            title={dict.about.transparency}
                            description={dict.about.transparencyDesc}
                            className="bg-success/5"
                        />
                    </div>

                    <div className="relative overflow-hidden rounded-[48px] border border-border group shadow-3xl">
                        <div className="absolute inset-0 bg-primary" />
                        <div className="absolute inset-0 bg-linear-to-br from-accent/40 to-transparent opacity-50 transition-opacity group-hover:opacity-70" />
                        <div className="absolute -bottom-24 -right-24 w-96 h-96 bg-accent/20 rounded-full blur-3xl pointer-events-none group-hover:scale-110 transition-transform duration-700" />

                        <div className="relative z-10 p-12 md:p-20 flex flex-col md:flex-row justify-between items-center gap-12">
                            <div className="max-w-2xl text-center md:text-left">
                                <h2 className="text-5xl md:text-7xl font-black mb-8 tracking-tighter uppercase text-surface leading-[0.9]">
                                    {dict.about.ctaTitle}
                                </h2>
                                <p className="text-surface/80 text-xl font-bold leading-relaxed mb-0">
                                    {dict.about.ctaDesc}
                                </p>
                            </div>
                            <Link href="/contact" className="btn-primary flex items-center gap-4 hover:scale-105! shadow-2xl shadow-white/10 group/btn border-none! py-6 px-12">
                                <span className="font-black uppercase tracking-[0.2em] text-sm">{dict.about.ctaButton}</span>
                                <ArrowRight size={24} className="group-hover/btn:translate-x-1 transition-transform" />
                            </Link>
                        </div>
                    </div>
                </section>

                <Footer minimal={true} />
            </div>
        </main>
    );
}
