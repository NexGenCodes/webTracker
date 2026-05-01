"use client";

import { useState } from 'react';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { CheckCircle2, Zap, ArrowRight, ShieldCheck, HelpCircle } from 'lucide-react';
import { subscribeAction } from '@/app/actions/billing';

import { BILLING_PLANS } from '@/constants';
import { useI18n } from '@/components/providers/I18nContext';
import toast from 'react-hot-toast';

export default function BillingPage() {
    const { user } = useMultiTenant();
    const { dict } = useI18n();
    const [billingCycle, setBillingCycle] = useState<'monthly' | 'annually'>('monthly');
    const [loadingPlan, setLoadingPlan] = useState<string | null>(null);

    const handleSubscribe = async (planId: string) => {
        setLoadingPlan(planId);
        try {
            const data = await subscribeAction(planId, window.location.href);
            if (data.authorization_url) {
                window.location.href = data.authorization_url;
            } else {
                toast.error(data.error || 'Failed to start subscription');
            }
        } catch {
            toast.error('Network error. Please try again.');
        } finally {
            setLoadingPlan(null);
        }
    };

    return (
        <main className="min-h-screen bg-transparent selection:bg-accent/20">
            <div className="pt-12 pb-24 px-6 relative z-10 max-w-7xl mx-auto">
                {/* Header Section */}
                <div className="text-center max-w-3xl mx-auto mb-16 animate-fade-in">
                    <h1 className="text-3xl md:text-5xl font-black uppercase tracking-tighter mb-4 text-text-main">
                        Simple, <span className="text-accent">Transparent</span> Pricing
                    </h1>
                    <p className="text-text-muted text-sm md:text-base mb-8">
                        No hidden fees. No surprise charges. Upgrade, downgrade, or cancel at any time.
                    </p>

                    {/* Toggle */}
                    <div className="inline-flex items-center bg-surface p-1 rounded-2xl border border-border">
                        <button
                            onClick={() => setBillingCycle('monthly')}
                            className={`px-6 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all ${billingCycle === 'monthly' ? 'bg-accent text-white shadow-lg shadow-accent/20' : 'text-text-muted hover:text-text-main'}`}
                        >
                            Monthly
                        </button>
                        <button
                            onClick={() => setBillingCycle('annually')}
                            className={`px-6 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all ${billingCycle === 'annually' ? 'bg-accent text-white shadow-lg shadow-accent/20' : 'text-text-muted hover:text-text-main'}`}
                        >
                            Annually <span className="text-[9px] bg-green-500/20 text-green-500 px-2 py-0.5 rounded-full ml-1">-15%</span>
                        </button>
                    </div>
                </div>

                {/* Current Plan Banner (If Logged In) */}
                {user && (
                    <div className="max-w-4xl mx-auto mb-12 bg-surface/50 border border-accent/20 rounded-2xl p-6 flex flex-col sm:flex-row items-center justify-between gap-4 animate-fade-in" style={{ animationDelay: '0.1s' }}>
                        <div>
                            <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-1">Current Plan</p>
                            <div className="flex items-center gap-3">
                                <h3 className="text-xl font-black text-text-main uppercase tracking-wider">{user.plan_type || 'Trial'}</h3>
                                <span className="px-3 py-1 bg-accent/10 text-accent text-[10px] font-black uppercase tracking-widest rounded-full border border-accent/20">
                                    Active
                                </span>
                            </div>
                        </div>
                        <div className="text-center sm:text-right">
                            <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-1">Billing Cycle</p>
                            <p className="text-sm font-bold text-text-main">Renews automatically</p>
                        </div>
                    </div>
                )}

                {/* Pricing Cards */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
                    {BILLING_PLANS.map((plan, index) => {
                        const plansDict = (dict.marketing?.pricing?.plans as Record<string, string>) || {};
                        return (
                            <div
                                key={plan.id}
                                className={`relative bg-surface rounded-3xl border transition-all duration-300 hover:scale-[1.02] flex flex-col ${plan.popular ? 'border-accent shadow-2xl shadow-accent/10' : 'border-border'}`}
                                style={{ animationDelay: `${0.2 + (index * 0.1)}s` }}
                            >
                                {plan.popular && (
                                    <div className="absolute -top-4 left-0 right-0 flex justify-center">
                                        <div className="bg-accent text-white text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-full flex items-center gap-1 shadow-lg shadow-accent/30">
                                            <Zap size={12} className="fill-white" /> {dict.marketing?.pricing?.mostPopular || 'Most Popular'}
                                        </div>
                                    </div>
                                )}

                                <div className="p-8 border-b border-border">
                                    <h3 className="text-lg font-black uppercase tracking-widest text-text-main mb-2">{plansDict[plan.nameKey]}</h3>
                                    <p className="text-text-muted text-sm min-h-[40px] mb-6">{plansDict[plan.descKey]}</p>

                                    <div className="mb-2">
                                        <span className="text-4xl font-black text-text-main tracking-tighter">
                                            {billingCycle === 'annually' && plan.price !== 'Custom'
                                                ? `₦${(parseInt(plan.price.replace(/\D/g, '')) * 0.85).toLocaleString()}`
                                                : plan.price}
                                        </span>
                                        <span className="text-text-muted text-sm font-bold ml-1">{plansDict[plan.intervalKey]}</span>
                                    </div>

                                    {plan.trialKey ? (
                                        <p className="text-accent text-xs font-black uppercase tracking-widest mb-6">+{plansDict[plan.trialKey]}</p>
                                    ) : (
                                        <div className="h-[24px] mb-6"></div> // Spacer
                                    )}

                                    <button
                                        onClick={() => handleSubscribe(plan.id)}
                                        disabled={loadingPlan === plan.id}
                                        className={`w-full py-4 rounded-xl text-xs font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2 ${plan.popular ? 'bg-accent text-white hover:bg-accent-hover shadow-lg shadow-accent/20' : 'bg-surface-muted text-text-main hover:bg-border border border-border'} disabled:opacity-50`}
                                    >
                                        {loadingPlan === plan.id ? 'Please wait...' : plansDict[plan.btnKey]}
                                        {loadingPlan !== plan.id && <ArrowRight size={14} />}
                                    </button>
                                </div>

                                <div className="p-8 flex-1 bg-surface/50 rounded-b-3xl">
                                    <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-6">What&apos;s included</p>
                                    <ul className="space-y-4">
                                        {plan.features.map((feature: string, i: number) => (
                                            <li key={i} className="flex items-start gap-3">
                                                <CheckCircle2 size={16} className="text-accent shrink-0 mt-0.5" />
                                                <span className="text-sm text-text-main font-medium">{plansDict[feature]}</span>
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            </div>
                        )
                    })}
                </div>

                {/* FAQ or Trust Section */}
                <div className="mt-24 max-w-3xl mx-auto text-center border-t border-border pt-16">
                    <ShieldCheck size={32} className="text-text-muted mx-auto mb-4" />
                    <h3 className="text-lg font-black uppercase tracking-widest text-text-main mb-2">Secure & Reliable</h3>
                    <p className="text-text-muted text-sm mb-6">
                        All payments are securely processed by Paystack. You can cancel your subscription at any time from your dashboard without penalty.
                    </p>
                    <a href="/contact" className="inline-flex items-center gap-2 text-accent text-sm font-bold hover:underline">
                        <HelpCircle size={16} /> Have pricing questions? Contact us.
                    </a>
                </div>
            </div>
        </main>
    );
}
