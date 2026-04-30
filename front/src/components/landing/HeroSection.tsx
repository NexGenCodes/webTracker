"use client";

import { ArrowRight, ShieldCheck } from 'lucide-react';
import Link from 'next/link';
import { useI18n } from '@/components/providers/I18nContext';

export function HeroSection() {
  const { dict } = useI18n();

  return (
    <section className="py-16 md:py-32 lg:py-40 relative z-10 w-full">
      <div className="max-w-5xl mx-auto px-4 md:px-6 text-center">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-[10px] font-black uppercase tracking-[0.3em] mb-8 animate-fade-in">
          <ShieldCheck size={14} />
          {dict.common.safeLogistics}
        </div>

        {/* Headline */}
        <h1 className="text-4xl sm:text-5xl md:text-7xl lg:text-8xl font-black mb-6 tracking-tighter uppercase text-gradient leading-[0.9] animate-fade-in">
          {dict.marketing?.hero?.headline || 'Next Generation'}
          <br />
          {dict.marketing?.hero?.headlineLine2 || 'Global Logistics'}
        </h1>

        {/* Subtitle */}
        <p className="text-text-muted text-base md:text-lg lg:text-xl max-w-2xl mx-auto font-bold mb-12 animate-fade-in opacity-80">
          {dict.marketing?.hero?.subtitle || 'Track shipments in real-time across borders. AI-powered parsing, WhatsApp alerts, and enterprise-grade security.'}
        </p>

        {/* CTA Buttons */}
        <div className="max-w-2xl mx-auto mb-12 flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in">
          <Link
            href="/track"
            className="px-8 py-4 bg-accent text-white rounded-[14px] font-black uppercase tracking-widest text-sm md:text-base transition-all hover:bg-accent/90 active:scale-95 shadow-lg shadow-accent/20 flex items-center justify-center gap-2 w-full sm:w-auto"
          >
            {dict.marketing?.hero?.ctaTrack || 'Track a Shipment'}
            <ArrowRight size={18} />
          </Link>
          <Link
            href="/pricing"
            className="px-8 py-4 bg-surface-muted text-text-main border border-border rounded-[14px] font-black uppercase tracking-widest text-sm md:text-base transition-all hover:bg-surface active:scale-95 flex items-center justify-center gap-2 w-full sm:w-auto"
          >
            {dict.marketing?.hero?.ctaPricing || 'View Pricing'}
          </Link>
        </div>

        {/* Secondary CTA */}
        <div className="mt-8 flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in">
          <Link
            href="/auth"
            className="text-accent hover:text-accent/80 text-xs font-black uppercase tracking-[0.2em] transition-colors flex items-center gap-2"
          >
            {dict.marketing?.hero?.ctaCreate || 'Create Free Account'} <ArrowRight size={14} />
          </Link>
        </div>
      </div>
    </section>
  );
}
