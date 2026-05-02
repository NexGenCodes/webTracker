"use client";

import { useState, useTransition, useEffect } from 'react';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { CheckCircle2, Zap, ArrowRight, ShieldCheck, HelpCircle, Loader2 } from 'lucide-react';
import { subscribeAction, getPlansAction, PlanData, waitForPaymentAction } from '@/app/actions/billing';
import { useSearchParams, useRouter } from 'next/navigation';

import { formatPrice } from '@/lib/utils';
import { useI18n } from '@/components/providers/I18nContext';
import toast from 'react-hot-toast';

export default function BillingPage() {
    const { user, companyId, refreshAuth } = useMultiTenant();
    const { dict } = useI18n();
    const searchParams = useSearchParams();
    const router = useRouter();
    const [billingCycle, setBillingCycle] = useState<'monthly' | 'annually'>('monthly');
    const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
    const [isVerifying, setIsVerifying] = useState(false);
    const [, startTransition] = useTransition();
    const [plans, setPlans] = useState<PlanData[]>([]);
    const [loadingPlans, setLoadingPlans] = useState(true);

    const reference = searchParams.get('reference');

    const handleVerifyPayment = async () => {
        if (!companyId || !reference) return;
        setIsVerifying(true);

        const result = await waitForPaymentAction(reference);
        
        if (result.success && result.data?.status === 'active') {
            toast.success('Subscription activated successfully!');
            await refreshAuth();
        } else {
            toast.error(result.error || 'Payment verification is taking longer than usual.');
        }

        setIsVerifying(false);
        router.replace('/dashboard/billing');
    };

    useEffect(() => {
        const fetchPlans = async () => {
            const res = await getPlansAction();
            if (res.success && res.data) {
                setPlans(res.data);
            }
            setLoadingPlans(false);
        };

        fetchPlans();
    }, []);

    useEffect(() => {
        if (reference) {
            handleVerifyPayment();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [reference]);

    const handleSubscribe = async (planId: string) => {
        setLoadingPlan(planId);
        startTransition(async () => {
            const result = await subscribeAction(planId, window.location.href);

            if (result.success && result.data?.authorization_url) {
                window.location.href = result.data.authorization_url;
            } else {
                toast.error(result.error || 'Failed to start subscription');
            }
            setLoadingPlan(null);
        });
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
                {loadingPlans ? (
                    <div className="flex justify-center items-center py-20">
                        <Loader2 className="animate-spin text-accent" size={40} />
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
                        {plans.map((plan, index) => {
                            const plansDict = (dict.marketing?.pricing?.plans as Record<string, string>) || {};
                            const isCustom = plan.id === 'enterprise';

                            let priceDisplay = '';
                            if (isCustom) {
                                priceDisplay = dict.marketing?.pricing?.custom || 'Custom';
                            } else {
                                const basePrice = plan.price;
                                if (billingCycle === 'annually') {
                                    const yearlyMonthly = basePrice * 0.85; // 15% discount
                                    priceDisplay = formatPrice(yearlyMonthly, plan.currency);
                                } else {
                                    priceDisplay = formatPrice(basePrice, plan.currency);
                                }
                            }

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
                                        <h3 className="text-lg font-black uppercase tracking-widest text-text-main mb-2">{plansDict[plan.name_key]}</h3>
                                        <p className="text-text-muted text-sm min-h-[40px] mb-6">{plansDict[plan.desc_key]}</p>

                                        <div className="mb-2">
                                            <span className="text-4xl font-black text-text-main tracking-tighter">
                                                {priceDisplay}
                                            </span>
                                            {!isCustom && (
                                                <span className="text-text-muted text-sm font-bold ml-1">{plansDict[plan.interval_key]}</span>
                                            )}
                                        </div>

                                        {plan.trial_key ? (
                                            <p className="text-accent text-xs font-black uppercase tracking-widest mb-6">+{plansDict[plan.trial_key]}</p>
                                        ) : (
                                            <div className="h-[24px] mb-6"></div> // Spacer
                                        )}

                                        <button
                                            onClick={() => handleSubscribe(plan.id)}
                                            disabled={loadingPlan === plan.id}
                                            className={`w-full py-4 rounded-xl text-xs font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2 ${plan.popular ? 'bg-accent text-white hover:bg-accent-hover shadow-lg shadow-accent/20' : 'bg-surface-muted text-text-main hover:bg-border border border-border'} disabled:opacity-50`}
                                        >
                                            {loadingPlan === plan.id ? 'Please wait...' : plansDict[plan.btn_key]}
                                            {loadingPlan !== plan.id && <ArrowRight size={14} />}
                                        </button>
                                    </div>

                                    <div className="p-8 flex-1 bg-surface/50 rounded-b-3xl">
                                        <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-6">What&apos;s included</p>
                                        <ul className="space-y-4">
                                            {plan.features?.map((feature: string, i: number) => (
                                                <li key={i} className="flex items-start gap-3">
                                                    <CheckCircle2 size={16} className="text-accent shrink-0 mt-0.5" />
                                                    <span className="text-sm text-text-main font-medium">{plansDict[feature] || feature}</span>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                </div>
                            )
                        })}
                    </div>
                )}

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

            {/* Verification Overlay */}
            {isVerifying && (
                <div className="fixed inset-0 z-[100] bg-background/80 backdrop-blur-sm flex flex-col items-center justify-center p-6 animate-fade-in">
                    <div className="bg-surface border border-accent/20 rounded-3xl p-8 max-w-sm w-full text-center shadow-2xl shadow-accent/10">
                        <Loader2 className="w-12 h-12 text-accent animate-spin mx-auto mb-6" />
                        <h2 className="text-xl font-black uppercase tracking-widest text-text-main mb-2">Verifying Payment</h2>
                        <p className="text-text-muted text-sm mb-6">
                            We&apos;re confirming your payment with Paystack. This usually takes just a few seconds.
                        </p>
                        <div className="flex items-center justify-center gap-2 text-[10px] font-black uppercase tracking-widest text-accent/60">
                            <ShieldCheck size={14} /> Secured by Paystack
                        </div>
                    </div>
                </div>
            )}
        </main>
    );
}
