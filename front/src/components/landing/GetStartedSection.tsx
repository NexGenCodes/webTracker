"use client";

import { ArrowRight, ShieldCheck, Sparkles } from 'lucide-react';
import Link from 'next/link';
import { useI18n } from '@/components/providers/I18nContext';

export function GetStartedSection() {
  const { dict } = useI18n();
  return (
    <section className="py-24 md:py-32 relative z-10 w-full">
      <div className="max-w-5xl mx-auto px-4 md:px-6">
        <div className="glass-panel p-10 md:p-20 border-accent/20 relative overflow-hidden text-center rounded-[2.5rem]">
          {/* Ambient glow */}
          <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-accent/10 rounded-full -mr-48 -mt-48 blur-[100px] pointer-events-none" />
          <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-primary/10 rounded-full -ml-48 -mb-48 blur-[100px] pointer-events-none" />
          <div className="absolute inset-0 bg-gradient-to-br from-accent/5 via-transparent to-accent/5 pointer-events-none" />

          <div className="relative z-10">
            <div className="inline-flex items-center justify-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-[10px] font-black uppercase tracking-[0.3em] mb-8">
              <Sparkles size={12} />
              {dict.marketing?.getStarted?.badge || 'Start Tracking Today'}
            </div>

            <h2 className="text-3xl md:text-5xl lg:text-6xl font-black mb-6 tracking-tighter uppercase text-gradient leading-[0.95]" dangerouslySetInnerHTML={{ __html: dict.marketing?.getStarted?.title || 'Ready to Upgrade<br />Your Logistics?' }}>
            </h2>

            <p className="text-text-muted text-base md:text-lg max-w-xl mx-auto mb-12 font-bold">
              {dict.marketing?.getStarted?.subtitle || 'Transform your operations with AI-powered tracking, instant WhatsApp alerts, and real-time global visibility.'}
            </p>

            <div className="flex items-center justify-center max-w-lg mx-auto">
              <Link
                href="/auth"
                className="w-full sm:w-auto px-10 py-5 bg-accent text-white rounded-xl font-black uppercase tracking-widest text-sm md:text-base transition-all hover:bg-accent/90 shadow-xl shadow-accent/20 active:scale-95 flex items-center justify-center gap-3"
              >
                {dict.marketing?.getStarted?.ctaAccount || 'Create Free Account'} <ArrowRight size={18} />
              </Link>
            </div>

            {/* Trust line */}
            <div className="mt-10 flex items-center justify-center gap-2 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted/50">
              <ShieldCheck size={12} />
              {dict.marketing?.getStarted?.trustLine || 'No credit card required · 7-day free trial · Cancel anytime'}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
