"use client";

import { Package, MapPin, CheckCircle2 } from 'lucide-react';

import { useI18n } from '@/components/providers/I18nContext';

export function HowItWorksSection() {
  const { dict } = useI18n();

  const steps = [
    {
      number: "01",
      icon: Package,
      title: dict.marketing?.howItWorks?.step1?.title || "Create",
      description: dict.marketing?.howItWorks?.step1?.description || "Register shipments via admin portal, WhatsApp, or bulk CSV upload with AI-powered parsing.",
    },
    {
      number: "02",
      icon: MapPin,
      title: dict.marketing?.howItWorks?.step2?.title || "Track",
      description: dict.marketing?.howItWorks?.step2?.description || "Customers get a unique tracking link with real-time map visualization and live status updates.",
    },
    {
      number: "03",
      icon: CheckCircle2,
      title: dict.marketing?.howItWorks?.step3?.title || "Deliver",
      description: dict.marketing?.howItWorks?.step3?.description || "Automated WhatsApp & email notifications at every milestone. Instant proof of delivery.",
    },
  ];
  return (
    <section className="py-24 md:py-32 relative z-10 w-full">
      <div className="max-w-7xl mx-auto px-4 md:px-6">
        <div className="text-center mb-20">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-[10px] font-black uppercase tracking-[0.3em] mb-6">
            {dict.marketing?.howItWorks?.badge || 'How It Works'}
          </div>
          <h2 className="text-3xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-gradient">
            {dict.marketing?.howItWorks?.title || 'Three Steps to Global Control'}
          </h2>
          <p className="text-text-muted text-lg max-w-xl mx-auto font-bold">
            {dict.marketing?.howItWorks?.subtitle || 'From creation to delivery — a seamless pipeline.'}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 relative">
          {/* Connecting line (desktop only) */}
          <div className="hidden md:block absolute top-[72px] left-[16.67%] right-[16.67%] h-[2px]">
            <div className="w-full h-full bg-gradient-to-r from-accent/0 via-accent/40 to-accent/0" />
          </div>

          {steps.map((step, i) => {
            const Icon = step.icon;
            return (
              <div
                key={i}
                className="glass-panel p-8 md:p-10 relative overflow-hidden group hover:border-accent/30 hover:-translate-y-2 transition-all duration-500"
              >
                {/* Step number glow */}
                <div className="absolute -top-4 -right-4 w-24 h-24 bg-accent/5 rounded-full blur-2xl group-hover:bg-accent/10 transition-all duration-500 pointer-events-none" />

                <div className="relative z-10">
                  {/* Number + Icon row */}
                  <div className="flex items-center gap-4 mb-8">
                    <span className="text-5xl font-black text-accent/20 group-hover:text-accent/40 transition-colors duration-500 leading-none">
                      {step.number}
                    </span>
                    <div className="w-14 h-14 bg-accent/10 rounded-2xl flex items-center justify-center text-accent group-hover:bg-accent group-hover:text-white transition-all duration-500 ring-1 ring-accent/10 group-hover:ring-accent/30 shadow-lg shadow-transparent group-hover:shadow-accent/10">
                      <Icon size={24} />
                    </div>
                  </div>

                  <h3 className="text-2xl font-black mb-3 tracking-tight uppercase">
                    {step.title}
                  </h3>
                  <p className="text-text-muted text-sm leading-relaxed font-semibold opacity-80 group-hover:opacity-100 transition-opacity">
                    {step.description}
                  </p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
