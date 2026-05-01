"use client";

import { Suspense, useEffect, useState } from 'react';
import { MessageSquare, Cpu, Palette } from 'lucide-react';
import { useI18n } from '@/components/providers/I18nContext';
import { FeatureCard } from '@/components/landing/FeatureCard';
import { HeroSection } from '@/components/landing/HeroSection';
import { TrustMetricsSection } from '@/components/landing/TrustMetricsSection';
import { HowItWorksSection } from '@/components/landing/HowItWorksSection';

import { GetStartedSection } from '@/components/landing/GetStartedSection';

function HomeContent() {
  const { dict } = useI18n();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;

  return (
    <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative">
      <div className="w-full flex flex-col flex-1 relative z-10">

        {/* 1. Hero — Search bar + animated tracking demo */}
        <div className="pt-28 md:pt-36">
          <HeroSection />
        </div>

        {/* 2. Trust Metrics — Animated counters */}
        <TrustMetricsSection />

        {/* 3. Feature Cards */}
        <section className="py-16 md:py-24 w-full">
          <div className="max-w-7xl mx-auto px-4 md:px-6">
            <div className="text-center mb-16">
              <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-[10px] font-black uppercase tracking-[0.3em] mb-6">
                {dict.marketing?.features?.badge || 'Features'}
              </div>
              <h2 className="text-3xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-gradient">
                {dict.marketing?.features?.title || 'Built for Modern Logistics'}
              </h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              <FeatureCard
                icon={<MessageSquare />}
                title={dict.marketing?.features?.f2?.title || "Automated WhatsApp Bot"}
                description={dict.marketing?.features?.f2?.desc || "Instantly notify customers via WhatsApp at every milestone. Eliminate 'where is my package' support calls."}
              />
              <FeatureCard
                icon={<Cpu />}
                title={dict.marketing?.features?.f3?.title || "AI-Powered Parsing"}
                description={dict.marketing?.features?.f3?.desc || "Paste raw shipping manifests or forward emails. Our AI automatically extracts tracking numbers and customer details."}
              />
              <FeatureCard
                icon={<Palette />}
                title={dict.marketing?.features?.f4?.title || "Custom Branding"}
                description={dict.marketing?.features?.f4?.desc || "Your logo, your colors. Maintain brand consistency while providing a premium, real-time tracking experience to your clients."}
              />
            </div>
          </div>
        </section>

        {/* 4. How It Works — 3-step pipeline */}
        <HowItWorksSection />



        {/* 6. CTA — Get Started */}
        <GetStartedSection />

      </div>
    </main>
  );
}

export default function Home() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent" />
      </div>
    }>
      <HomeContent />
    </Suspense>
  );
}
