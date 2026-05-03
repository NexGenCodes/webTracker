"use client";

import { useState, useTransition, useEffect } from 'react';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { CheckCircle2, Zap, ArrowRight, HelpCircle, Loader2, Clock, ExternalLink, ShieldCheck } from 'lucide-react';
import { subscribeAction, PlanData, checkPaymentStatusAction, getPaymentHistoryAction, PaymentData } from '@/app/actions/billing';
import { useSearchParams, useRouter } from 'next/navigation';
import { createClient } from '@/lib/supabase/client';

import { formatPrice } from '@/lib/utils';
import { useI18n } from '@/components/providers/I18nContext';
import toast from 'react-hot-toast';

interface BillingClientProps {
    initialPlans: PlanData[];
    initialPayments: PaymentData[];
    companyData: {
        subscription_expiry: string | null;
        plan_type: string | null;
        subscription_status: string | null;
    } | null;
}

export default function BillingClient({ initialPlans, initialPayments, companyData }: BillingClientProps) {
    const { user, companyId, refreshAuth } = useMultiTenant();
    const { dict } = useI18n();
    const searchParams = useSearchParams();
    const router = useRouter();

    const [billingCycle, setBillingCycle] = useState<'monthly' | 'annually'>('monthly');
    const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
    const [isVerifying, setIsVerifying] = useState(false);
    const [activeTab, setActiveTab] = useState<'plans' | 'subscription' | 'history'>('plans');
    const [, startTransition] = useTransition();

    const [plans] = useState<PlanData[]>(initialPlans);
    const [payments, setPayments] = useState<PaymentData[]>(initialPayments);
    const [loadingPayments] = useState(false);

    const reference = searchParams.get('reference');

    const handleVerifyPayment = async () => {
        if (!companyId || !reference) return;
        setIsVerifying(true);

        // Wait a few seconds to let webhook process
        await new Promise(resolve => setTimeout(resolve, 3000));

        const result = await checkPaymentStatusAction();

        if (result.success && result.data?.status === 'active') {
            toast.success('Subscription activated successfully!');
            await refreshAuth();
            router.refresh(); // Force Server Component to re-fetch
        } else {
            toast.success('Payment received. Dashboard will update automatically once processed.');
        }

        setIsVerifying(false);
        router.replace('/dashboard/billing');
    };

    // Realtime subscription for payments
    useEffect(() => {
        if (!companyId) return;

        const supabase = createClient();

        const channel = supabase
            .channel(`payments-${companyId}`)
            .on(
                'postgres_changes',
                {
                    event: '*',
                    schema: 'public',
                    table: 'payments',
                    filter: `company_id=eq.${companyId}`
                },
                () => {
                    // Re-fetch payment history when any change occurs
                    // if the page is a Server Component!
                    router.refresh();

                    // Fallback for immediate UI update
                    const fetchPayments = async () => {
                        const res = await getPaymentHistoryAction();
                        if (res.success && res.data) {
                            setPayments(res.data);
                        }
                    };
                    fetchPayments();
                }
            )
            .subscribe();

        return () => {
            supabase.removeChannel(channel);
        };
    }, [companyId, router]);

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
                </div>

                {/* Tab Navigation */}
                <div className="max-w-6xl mx-auto mb-12 flex items-center justify-center border-b border-border">
                    <div className="flex gap-8">
                        {[
                            { id: 'plans', label: 'Pricing Plans' },
                            { id: 'subscription', label: 'Subscription' },
                            { id: 'history', label: 'Transaction History' }
                        ].map((tab) => (
                            <button
                                key={tab.id}
                                onClick={() => setActiveTab(tab.id as 'plans' | 'subscription' | 'history')}
                                className={`pb-4 text-xs font-black uppercase tracking-widest transition-all relative ${activeTab === tab.id
                                    ? 'text-accent'
                                    : 'text-text-muted hover:text-text-main'
                                    }`}
                            >
                                {tab.label}
                                {activeTab === tab.id && (
                                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent animate-in fade-in slide-in-from-bottom-1 duration-300" />
                                )}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Tab Content */}
                <div className="min-h-[400px]">
                    {activeTab === 'subscription' && user && (
                        <div className="max-w-4xl mx-auto mb-12 bg-surface/50 border border-accent/20 rounded-2xl p-6 flex flex-col sm:flex-row items-center justify-between gap-4 animate-fade-in">
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
                                <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-1">Next Billing Cycle</p>
                                <p className="text-sm font-bold text-text-main">
                                    {companyData?.subscription_expiry
                                        ? new Date(companyData.subscription_expiry).toLocaleDateString(undefined, {
                                            year: 'numeric',
                                            month: 'long',
                                            day: 'numeric'
                                        })
                                        : 'Renews automatically'}
                                </p>
                            </div>
                        </div>
                    )}

                    {activeTab === 'plans' && (
                        <div className="animate-fade-in">
                            {/* Toggle */}
                            <div className="flex justify-center mb-12">
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

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
                                {plans.map((plan) => {
                                    const plansDict = (dict.marketing?.pricing?.plans as Record<string, string>) || {};
                                    const isCustom = plan.id === 'enterprise';
                                    const isCurrentPlan = companyData?.plan_type?.toLowerCase() === plan.id.toLowerCase();

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
                                            className={`relative bg-surface rounded-3xl border transition-all duration-300 flex flex-col ${isCurrentPlan
                                                ? 'border-border opacity-60 grayscale-[0.5] scale-[0.98]'
                                                : plan.popular
                                                    ? 'border-accent shadow-2xl shadow-accent/10 hover:scale-[1.02]'
                                                    : 'border-border hover:scale-[1.02]'
                                                }`}
                                        >
                                            {isCurrentPlan && (
                                                <div className="absolute -top-4 left-0 right-0 flex justify-center">
                                                    <div className="bg-surface-muted text-text-muted text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-full flex items-center gap-1 border border-border shadow-sm">
                                                        <ShieldCheck size={12} /> Active Plan
                                                    </div>
                                                </div>
                                            )}

                                            {plan.popular && !isCurrentPlan && (
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

                                                {(() => {
                                                    const isTrial = companyData?.plan_type?.toLowerCase() === 'trial' || !companyData?.plan_type;

                                                    // Determine button text
                                                    let btnText = plansDict[plan.btn_key];
                                                    if (isCurrentPlan) {
                                                        btnText = 'Active';
                                                    } else if (plan.id === 'starter' && !isTrial) {
                                                        btnText = 'Get Starter';
                                                    } else if (plan.id === 'starter' && isTrial) {
                                                        btnText = 'Subscribe';
                                                    }

                                                    return (
                                                        <button
                                                            onClick={() => handleSubscribe(plan.id)}
                                                            disabled={loadingPlan === plan.id || isCurrentPlan}
                                                            className={`w-full py-4 rounded-xl text-xs font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2 ${isCurrentPlan
                                                                ? 'bg-surface-muted text-text-muted cursor-default border border-border'
                                                                : plan.popular
                                                                    ? 'bg-accent text-white hover:bg-accent-hover shadow-lg shadow-accent/20'
                                                                    : 'bg-surface-muted text-text-main hover:bg-border border border-border'
                                                                } disabled:opacity-50`}
                                                        >
                                                            {loadingPlan === plan.id ? 'Please wait...' : btnText}
                                                            {loadingPlan !== plan.id && !isCurrentPlan && <ArrowRight size={14} />}
                                                        </button>
                                                    );
                                                })()}
                                            </div>

                                            <div className="p-8 flex-1 bg-surface/50 rounded-b-3xl">
                                                <p className="text-[10px] font-black uppercase tracking-widest text-text-muted mb-6">What&apos;s included</p>
                                                <ul className="space-y-4">
                                                    {plan.features?.map((feature: string) => (
                                                        <li key={feature} className="flex items-start gap-3">
                                                            <CheckCircle2 size={16} className="text-accent shrink-0 mt-0.5" />
                                                            <span className="text-sm text-text-main font-medium">{plansDict[feature] || feature}</span>
                                                        </li>
                                                    ))}
                                                </ul>
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    )}

                    {activeTab === 'history' && (
                        <div className="max-w-5xl mx-auto animate-fade-in">
                            <div className="bg-surface border border-border rounded-3xl overflow-hidden shadow-xl shadow-black/5">
                                {loadingPayments ? (
                                    <div className="p-20 flex flex-col items-center gap-4 text-text-muted">
                                        <Loader2 className="animate-spin text-accent" size={32} />
                                        <p className="text-xs font-black uppercase tracking-widest">Loading transactions...</p>
                                    </div>
                                ) : payments.length === 0 ? (
                                    <div className="p-20 flex flex-col items-center gap-4 text-center">
                                        <div className="w-16 h-16 rounded-full bg-surface-muted flex items-center justify-center text-text-muted mb-2">
                                            <Clock size={32} />
                                        </div>
                                        <h3 className="text-lg font-black uppercase tracking-widest text-text-main">No Transactions Yet</h3>
                                        <p className="text-text-muted text-sm max-w-xs mx-auto">
                                            When you subscribe to a plan, your payment history will appear here.
                                        </p>
                                    </div>
                                ) : (
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-left border-collapse">
                                            <thead>
                                                <tr className="bg-surface-muted/50 border-b border-border">
                                                    <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-text-muted">Date</th>
                                                    <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-text-muted">Reference</th>
                                                    <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-text-muted">Amount</th>
                                                    <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-text-muted">Status</th>
                                                    <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-text-muted">Actions</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-border">
                                                {payments.map((payment) => (
                                                    <tr key={payment.id} className="hover:bg-surface-muted/30 transition-colors">
                                                        <td className="px-6 py-4">
                                                            <p className="text-sm font-bold text-text-main">
                                                                {new Date(payment.created_at).toLocaleDateString(undefined, {
                                                                    year: 'numeric',
                                                                    month: 'short',
                                                                    day: 'numeric'
                                                                })}
                                                            </p>
                                                            <p className="text-[10px] text-text-muted uppercase">
                                                                {new Date(payment.created_at).toLocaleTimeString(undefined, {
                                                                    hour: '2-digit',
                                                                    minute: '2-digit'
                                                                })}
                                                            </p>
                                                        </td>
                                                        <td className="px-6 py-4">
                                                            <code className="text-[11px] font-mono bg-surface-muted px-2 py-1 rounded text-text-muted border border-border">
                                                                {payment.reference}
                                                            </code>
                                                        </td>
                                                        <td className="px-6 py-4">
                                                            <span className="text-sm font-black text-text-main">
                                                                {formatPrice(payment.amount, 'NGN')}
                                                            </span>
                                                        </td>
                                                        <td className="px-6 py-4">
                                                            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border ${payment.status === 'success' || payment.status === 'active'
                                                                ? 'bg-green-500/10 text-green-500 border-green-500/20'
                                                                : payment.status === 'pending'
                                                                    ? 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20'
                                                                    : 'bg-red-500/10 text-red-500 border-red-500/20'
                                                                }`}>
                                                                <div className={`w-1.5 h-1.5 rounded-full ${payment.status === 'success' || payment.status === 'active' ? 'bg-green-500' : payment.status === 'pending' ? 'bg-yellow-500' : 'bg-red-500'
                                                                    }`} />
                                                                {payment.status}
                                                            </span>
                                                        </td>
                                                        <td className="px-6 py-4">
                                                            <button className="p-2 hover:bg-accent/10 hover:text-accent rounded-lg transition-colors text-text-muted flex items-center gap-2 text-[10px] font-black uppercase tracking-widest">
                                                                <ExternalLink size={14} /> Receipt
                                                            </button>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                )}
                            </div>
                            <p className="mt-4 text-center text-[10px] text-text-muted uppercase tracking-widest font-bold">
                                Secure payments processed by Paystack
                            </p>
                        </div>
                    )}
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
