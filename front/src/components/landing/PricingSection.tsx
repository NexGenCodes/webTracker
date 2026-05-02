"use client";

import { useEffect, useState } from 'react';
import { CheckCircle, ArrowRight, Loader2 } from 'lucide-react';
import { cn, formatPrice } from '@/lib/utils';
import Link from 'next/link';
import { getPlansAction, PlanData } from '@/app/actions/billing';
import { useI18n } from '@/components/providers/I18nContext';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';

export function PricingSection() {
  const { dict } = useI18n();
  const { user } = useMultiTenant();
  const [isYearly, setIsYearly] = useState(false);
  const [plans, setPlans] = useState<PlanData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPlans = async () => {
      const res = await getPlansAction();
      if (res.success && res.data) {
        setPlans(res.data);
      }
      setLoading(false);
    };

    fetchPlans();
  }, []);

  return (
    <section id="pricing" className="py-24 md:py-32 relative z-10 w-full">
      <div className="max-w-7xl mx-auto px-4 md:px-6">
        {/* Header */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-[10px] font-black uppercase tracking-[0.3em] mb-6">
            {dict.marketing?.pricing?.badge || 'Pricing'}
          </div>
          <h2 className="text-3xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-gradient">
            {dict.marketing?.pricing?.title || 'Simple, Transparent Pricing'}
          </h2>
          <p className="text-text-muted text-lg max-w-xl mx-auto font-bold mb-10">
            {dict.marketing?.pricing?.subtitle || 'Choose the perfect plan for your tracking needs. No hidden fees, cancel anytime.'}
          </p>

          {/* Toggle */}
          <div className="inline-flex items-center gap-4 glass-panel p-1.5 rounded-full">
            <button
              onClick={() => setIsYearly(false)}
              className={cn(
                "px-5 py-2 rounded-full text-xs font-black uppercase tracking-widest transition-all duration-300",
                !isYearly
                  ? "bg-accent text-white shadow-lg shadow-accent/20"
                  : "text-text-muted hover:text-text-main"
              )}
            >
              {dict.marketing?.pricing?.monthly || 'Monthly'}
            </button>
            <button
              onClick={() => setIsYearly(true)}
              className={cn(
                "px-5 py-2 rounded-full text-xs font-black uppercase tracking-widest transition-all duration-300 flex items-center gap-2",
                isYearly
                  ? "bg-accent text-white shadow-lg shadow-accent/20"
                  : "text-text-muted hover:text-text-main"
              )}
            >
              {dict.marketing?.pricing?.yearly || 'Yearly'}
              <span className="text-[9px] bg-success/20 text-success px-2 py-0.5 rounded-full font-black">
                {dict.marketing?.pricing?.saveLabel || 'Save 15%'}
              </span>
            </button>
          </div>
        </div>

        {/* Cards */}
        {loading ? (
          <div className="flex justify-center items-center py-20">
            <Loader2 className="animate-spin text-accent" size={40} />
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 items-stretch">
            {plans.map((plan, i) => {
              const isCustom = plan.id === 'enterprise';
              const plansDict = (dict.marketing?.pricing?.plans as Record<string, string>) || {};

              let priceDisplay = '';
              let yearlyBilled = '';
              if (isCustom) {
                  priceDisplay = dict.marketing?.pricing?.custom || 'Custom';
              } else {
                  const basePrice = plan.price;
                  if (isYearly) {
                      const yearlyMonthly = basePrice * 0.85; // 15% discount
                      priceDisplay = formatPrice(yearlyMonthly, plan.currency);
                      yearlyBilled = formatPrice(yearlyMonthly * 12, plan.currency);
                  } else {
                      priceDisplay = formatPrice(basePrice, plan.currency);
                  }
              }

              return (
                <div
                  key={plan.id}
                  className={cn(
                    "glass-panel p-8 md:p-10 relative overflow-hidden transition-all duration-500 hover:-translate-y-2 rounded-3xl flex flex-col",
                    plan.popular
                      ? "border-accent/40 shadow-xl shadow-accent/10 md:scale-105"
                      : "border-border/50 hover:border-accent/20"
                  )}
                >
                  {/* Popular indicator */}
                  {plan.popular && (
                    <>
                      <div className="absolute top-0 inset-x-0 h-1 bg-gradient-to-r from-accent to-accent/50" />
                      <div className="absolute top-4 right-4 bg-accent/20 text-accent text-[9px] font-black uppercase px-3 py-1 rounded-full tracking-[0.2em]">
                        {dict.marketing?.pricing?.mostPopular || 'Most Popular'}
                      </div>
                    </>
                  )}

                  {/* Plan name */}
                  <h3 className="text-xl font-black uppercase tracking-tight mb-2">{plansDict[plan.name_key]}</h3>
                  <p className="text-sm text-text-muted font-bold mb-6 min-h-[48px]">{plansDict[plan.desc_key]}</p>

                  {/* Price */}
                  <div className="mb-2">
                    {isCustom ? (
                      <span className="text-4xl md:text-5xl font-black">{priceDisplay}</span>
                    ) : (
                      <>
                        <span className="text-4xl md:text-5xl font-black">{priceDisplay}</span>
                        <span className="text-text-muted font-bold">{plansDict[plan.interval_key]}</span>
                        {isYearly && (
                          <div className="text-xs text-text-muted font-bold mt-1">
                            {dict.marketing?.pricing?.billed || 'Billed'} {yearlyBilled}{dict.marketing?.pricing?.perYear || '/year'}
                          </div>
                        )}
                      </>
                    )}
                  </div>

                  {plan.trial_key ? (
                    <p className="text-accent text-xs font-black uppercase tracking-widest mb-6">+{plansDict[plan.trial_key]}</p>
                  ) : (
                    <div className="h-[24px] mb-6"></div>
                  )}

                  {/* Features */}
                  <ul className="space-y-3 mb-8 flex-1">
                    {plan.features?.map((feature: string, idx: number) => (
                      <li key={idx} className="flex items-center gap-3">
                        <CheckCircle size={16} className="text-accent shrink-0" />
                        <span className="text-sm font-bold text-text-main/80">{plansDict[feature] || feature}</span>
                      </li>
                    ))}
                  </ul>

                  {/* CTA */}
                  <Link
                    href={isCustom ? "mailto:support@cargohive.com?subject=Enterprise%20Plan%20Inquiry" : (user ? "/dashboard/billing" : "/auth")}
                    className={cn(
                      "w-full py-4 rounded-xl font-black uppercase tracking-widest text-sm transition-all active:scale-95 flex items-center justify-center gap-2 mt-auto",
                      plan.popular
                        ? "bg-accent text-white hover:bg-accent/90 shadow-lg shadow-accent/20"
                        : "bg-surface-muted text-text-main hover:bg-surface border border-border hover:border-accent/30"
                    )}
                  >
                    {isCustom ? plansDict[plan.btn_key] : (user ? (dict.common?.dashboard ? `Go to ${dict.common.dashboard}` : "Go to Dashboard") : plansDict[plan.btn_key])}
                    <ArrowRight size={14} />
                  </Link>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
}
