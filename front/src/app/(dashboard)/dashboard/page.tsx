"use client";

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import {
    Package, Smartphone, CreditCard, Settings, LogOut,
    CheckCircle2, AlertTriangle, XCircle, Loader2,
    ChevronRight, Wifi, WifiOff, UserCircle2, Activity
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { getApiUrl } from '@/lib/utils';
import { createClient } from '@/lib/supabase/client';
import WhatsAppConnectModal from '@/components/dashboard/WhatsAppConnectModal';
type Tab = 'overview' | 'whatsapp' | 'billing' | 'settings';

interface CompanyData {
    name: string;
    admin_email: string;
    subscription_status: string;
    subscription_expiry: string;
    plan_type: string;
    auth_status: string;
    whatsapp_phone: string;
    brand_color: string;
    logo_url: string;
}

export default function DashboardPage() {
    const router = useRouter();
    const { user, companyId, loading, refreshAuth } = useMultiTenant();
    const [activeTab, setActiveTab] = useState<Tab>('overview');
    const [signingOut, setSigningOut] = useState(false);
    const [companyData, setCompanyData] = useState<CompanyData | null>(null);
    const [companyLoading, setCompanyLoading] = useState(true);

    const [settingsForm, setSettingsForm] = useState({ name: '', admin_email: '', logo_url: '' });
    const [isSaving, setIsSaving] = useState(false);

    const [shipmentStats, setShipmentStats] = useState({ total: 0, active: 0, delivered: 0 });
    const [fetchingStats, setFetchingStats] = useState(false);

    // Single Supabase client instance for all subscriptions in this component
    const supabase = createClient();

    // Connect WhatsApp Modal State
    const [isConnectModalOpen, setIsConnectModalOpen] = useState(false);

    const fetchCompanyData = useCallback(async () => {
        if (!companyId) {
            setCompanyLoading(false);
            return;
        }
        try {
            const { data, error } = await supabase
                .from('companies')
                .select('name, admin_email, subscription_status, subscription_expiry, plan_type, auth_status, whatsapp_phone, brand_color, logo_url, tracking_prefix')
                .eq('id', companyId)
                .single();

            if (data && !error) {
                setCompanyData(data);
                
                // Only populate the form if we are loading the company data for the first time
                // This prevents Realtime background refreshes from clobbering what the user is typing
                setSettingsForm(prev => {
                    if (!prev.name && !prev.admin_email) {
                        return {
                            name: data.name || '',
                            admin_email: data.admin_email || '',
                            logo_url: data.logo_url || '',
                        };
                    }
                    return prev;
                });
            } else if (error && error.code === 'PGRST116') {
                // If not found, they might need to re-authenticate
                refreshAuth();
            }
        } catch {
            // Silently fail
        } finally {
            setCompanyLoading(false);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [companyId, refreshAuth]);

    useEffect(() => {
        if (!loading && companyId) {
            fetchCompanyData();
        } else if (!loading) {
            setCompanyLoading(false);
        }
    }, [loading, companyId, fetchCompanyData]);

    // Supabase Realtime: global listener for company row changes
    useEffect(() => {
        if (!companyId) return;

        const channel = supabase
            .channel(`company-global-${companyId}`)
            .on(
                'postgres_changes',
                {
                    event: 'UPDATE',
                    schema: 'public',
                    table: 'companies',
                    filter: `id=eq.${companyId}`,
                },
                () => {
                    // Refresh company data on any backend change (auth_status, subscription, etc.)
                    fetchCompanyData();
                }
            )
            .subscribe();

        return () => {
            supabase.removeChannel(channel);
        };
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [companyId, fetchCompanyData]);


    useEffect(() => {
        if (!companyId) return;
        const fetchStats = async () => {
            setFetchingStats(true);
            try {
                const { data, error } = await supabase
                    .from('shipment')
                    .select('status')
                    .eq('company_id', companyId);

                if (data && !error) {
                    const active = data.filter(s => s.status !== 'DELIVERED' && s.status !== 'CANCELED').length;
                    const delivered = data.filter(s => s.status === 'DELIVERED').length;
                    setShipmentStats({ total: data.length, active, delivered });
                }
            } catch (err) {
                console.error("Error fetching stats:", err);
            } finally {
                setFetchingStats(false);
            }
        };

        fetchStats();

        // Subscribe to real-time changes
        const channel = supabase
            .channel(`shipment-stats-${companyId}`)
            .on(
                'postgres_changes',
                {
                    event: '*',
                    schema: 'public',
                    table: 'shipment',
                    filter: `company_id=eq.${companyId}`,
                },
                () => {
                    // Re-fetch stats quietly on any change
                    fetchStats();
                }
            )
            .subscribe();

        return () => {
            supabase.removeChannel(channel);
        };
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [companyId, companyData?.auth_status]);

    const handleSignOut = async () => {
        setSigningOut(true);
        await fetch(`${getApiUrl()}/api/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        await refreshAuth();
        router.push('/auth');
        router.refresh();
    };

    const handleSettingsSave = async () => {
        setIsSaving(true);
        try {
            const supabase = createClient();
            const { error } = await supabase
                .from('companies')
                .update(settingsForm)
                .eq('id', companyId);

            if (!error) {
                fetchCompanyData();
            } else if (error.code === '42501') { // RLS violation / auth issue
                refreshAuth();
            }
        } finally {
            setIsSaving(false);
        }
    };

    if (loading || companyLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center p-4 bg-background">
                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="flex flex-col items-center gap-6"
                >
                    <div className="bg-accent/20 p-4 rounded-3xl relative">
                        <div className="absolute inset-0 bg-accent/30 rounded-3xl blur-xl animate-pulse" />
                        <Package className="text-accent relative z-10 animate-bounce" size={40} />
                    </div>
                    <p className="text-accent font-black uppercase tracking-[0.3em] text-sm animate-pulse">
                        Loading Workspace...
                    </p>
                </motion.div>
            </div>
        );
    }

    const companyName = companyData?.name || user?.company_name || 'CARGOHIVE';
    const subscriptionStatus = companyData?.subscription_status || 'active';
    const whatsappConnected = companyData?.auth_status === 'active';
    const overallStatus = whatsappConnected ? subscriptionStatus : 'pending';

    const planType = (companyData?.plan_type || 'trial').toLowerCase();
    const planDetails: Record<string, { name: string; price: string }> = {
        trial: { name: '7 Days Free Trial', price: '₦0' },
        basic: { name: 'Basic Plan', price: '₦10,000' },
        pro: { name: 'Professional Plan', price: '₦25,000' },
        enterprise: { name: 'Enterprise Plan', price: '₦50,000' },
    };
    const currentPlan = planDetails[planType] || planDetails['trial'];

    const expiryDate = companyData?.subscription_expiry
        ? new Date(companyData.subscription_expiry).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
        : '—';

    const tabs: { id: Tab; icon: typeof Package; label: string }[] = [
        { id: 'overview', icon: Activity, label: 'Overview' },
        { id: 'whatsapp', icon: Smartphone, label: 'WhatsApp' },
        { id: 'billing', icon: CreditCard, label: 'Billing' },
        { id: 'settings', icon: Settings, label: 'Settings' },
    ];

    const StatusBadge = ({ status }: { status: string }) => {
        const config: Record<string, { icon: typeof CheckCircle2; color: string; label: string }> = {
            active: { icon: CheckCircle2, color: 'text-success', label: 'Active' },
            pending: { icon: AlertTriangle, color: 'text-warning', label: 'Pending Setup' },
            suspended: { icon: XCircle, color: 'text-error', label: 'Suspended' },
        };
        const c = config[status] || config.pending;
        return (
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full bg-surface border border-border/50 shadow-sm ${c.color}`}>
                <c.icon size={12} className={status === 'active' ? 'animate-pulse' : ''} />
                <span className="text-[10px] font-black uppercase tracking-widest">{c.label}</span>
            </div>
        );
    };

    return (
        <div className="pb-24 relative bg-background overflow-x-hidden">

            {/* Main Content Container */}
            <div className="max-w-6xl mx-auto z-10 relative pt-24 md:pt-32 px-4 sm:px-8">

                {/* Premium Header */}
                <motion.header
                    initial={{ y: -20, opacity: 0 }}
                    animate={{ y: 0, opacity: 1 }}
                    className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12"
                >
                    <div className="flex items-center gap-6">
                        {settingsForm.logo_url ? (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={settingsForm.logo_url} alt="Logo" className="w-16 h-16 rounded-2xl object-cover shadow-lg border border-border" />
                        ) : (
                            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-accent to-primary flex items-center justify-center shadow-lg shadow-accent/20 border border-white/10">
                                <span className="text-2xl font-black text-white">{companyName.substring(0, 2).toUpperCase()}</span>
                            </div>
                        )}
                        <div>
                            <h1 className="text-3xl sm:text-4xl font-black text-text-main uppercase tracking-tighter drop-shadow-sm">
                                {companyName}
                            </h1>
                            <div className="flex items-center gap-3 mt-3">
                                <StatusBadge status={overallStatus} />
                                <span className="w-1 h-1 rounded-full bg-border" />
                                <span className="text-xs font-bold text-text-muted uppercase tracking-widest flex items-center gap-2">
                                    <UserCircle2 size={14} /> {user?.email}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        {whatsappConnected && (
                            <button
                                onClick={handleSignOut}
                                disabled={signingOut}
                                className="flex items-center gap-2 px-5 py-2.5 bg-surface border border-border rounded-xl text-text-muted hover:text-error hover:border-error/30 hover:bg-error/5 transition-all text-xs font-black uppercase tracking-widest active:scale-95 shadow-sm"
                            >
                                {signingOut ? <Loader2 size={16} className="animate-spin" /> : <LogOut size={16} />}
                                Sign Out
                            </button>
                        )}
                    </div>
                </motion.header>

                {/* Animated Tabs */}
                <div className="flex flex-wrap gap-2 sm:gap-4 border-b border-border/50 mb-10 pb-2 relative">
                    {tabs.map((tab) => {
                        const isActive = activeTab === tab.id;
                        return (
                            <button
                                key={tab.id}
                                onClick={() => setActiveTab(tab.id)}
                                className={`relative flex items-center gap-2 px-4 py-3 font-black text-xs uppercase tracking-widest transition-colors whitespace-nowrap rounded-xl ${isActive ? 'text-text-main' : 'text-text-muted hover:text-text-main hover:bg-surface'
                                    }`}
                            >
                                <tab.icon size={16} className={isActive ? 'text-accent' : ''} />
                                {tab.label}
                                {isActive && (
                                    <motion.div
                                        layoutId="activeTabIndicator"
                                        className="absolute bottom-[-9px] left-0 right-0 h-[3px] bg-accent rounded-t-full shadow-[0_-2px_10px_rgba(var(--color-accent),0.5)]"
                                        initial={false}
                                        transition={{ type: "spring", stiffness: 500, damping: 30 }}
                                    />
                                )}
                            </button>
                        );
                    })}
                </div>

                {/* Tab Content with Exit/Enter Animations */}
                <div className="relative min-h-[400px]">
                    <AnimatePresence mode="wait">
                        <motion.div
                            key={activeTab}
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            transition={{ duration: 0.2 }}
                        >
                            {/* OVERVIEW TAB */}
                            {activeTab === 'overview' && (
                                <div className="space-y-8">
                                    {/* Action Cards */}
                                    {!whatsappConnected && (
                                        <div className="relative overflow-hidden glass-panel p-6 border-warning/40 bg-warning/5 flex flex-col sm:flex-row items-start sm:items-center gap-6 group hover:shadow-lg hover:shadow-warning/10 transition-all duration-300">
                                            <div className="absolute top-0 left-0 w-1 h-full bg-warning" />
                                            <div className="w-14 h-14 bg-warning/20 rounded-2xl flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform duration-300">
                                                <AlertTriangle className="text-warning" size={28} />
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="text-base font-black text-text-main uppercase tracking-widest mb-1">
                                                    Action Required: Connect WhatsApp
                                                </h3>
                                                <p className="text-sm font-medium text-text-muted">
                                                    Your tracking bot is currently offline. Connect your WhatsApp Business number to enable automated real-time tracking for your customers.
                                                </p>
                                            </div>
                                            <button
                                                onClick={() => {
                                                    setIsConnectModalOpen(true);
                                                }}
                                                className="btn-primary px-6 py-3.5 text-xs flex items-center gap-2 active:scale-95 shrink-0 shadow-lg shadow-warning/20 !bg-warning hover:!bg-warning/90 !text-black"
                                            >
                                                Connect Now <ChevronRight size={16} />
                                            </button>
                                        </div>
                                    )}

                                    {/* Quick Stats Grid */}
                                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                                        <div className="glass-panel p-6 border-border/40 hover:border-border hover:shadow-xl hover:shadow-black/5 hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
                                            <div className="absolute top-0 right-0 w-24 h-24 bg-primary/5 rounded-full blur-2xl group-hover:bg-primary/10 transition-all" />
                                            <div className="flex items-center justify-between mb-4">
                                                <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Current Plan</p>
                                                <div className="w-8 h-8 rounded-full bg-surface border border-border flex items-center justify-center text-text-main">
                                                    <CreditCard size={14} />
                                                </div>
                                            </div>
                                            <p className="text-2xl font-black text-text-main drop-shadow-sm">{currentPlan.name}</p>
                                            <p className="text-xs font-bold text-text-muted mt-2">
                                                {planType === 'trial' ? 'Free Trial Period' : `${currentPlan.price} / month`}
                                            </p>
                                        </div>

                                        <div className="glass-panel p-6 border-border/40 hover:border-border hover:shadow-xl hover:shadow-black/5 hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
                                            <div className="absolute top-0 right-0 w-24 h-24 bg-accent/5 rounded-full blur-2xl group-hover:bg-accent/10 transition-all" />
                                            <div className="flex items-center justify-between mb-4">
                                                <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Bot Status</p>
                                                <div className={`w-8 h-8 rounded-full border flex items-center justify-center ${whatsappConnected ? 'bg-success/10 border-success/30 text-success' : 'bg-error/10 border-error/30 text-error'}`}>
                                                    {whatsappConnected ? <Wifi size={14} /> : <WifiOff size={14} />}
                                                </div>
                                            </div>
                                            <p className="text-2xl font-black text-text-main drop-shadow-sm">
                                                {whatsappConnected ? 'Connected' : 'Offline'}
                                            </p>
                                            <p className="text-xs font-bold text-text-muted mt-2">
                                                {whatsappConnected ? 'Ready for tracking' : 'Action required'}
                                            </p>
                                        </div>

                                        <div className="glass-panel p-6 border-border/40 hover:border-border hover:shadow-xl hover:shadow-black/5 hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
                                            <div className="flex items-center justify-between mb-4">
                                                <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Next Billing</p>
                                                <div className="w-8 h-8 rounded-full bg-surface border border-border flex items-center justify-center text-text-main">
                                                    <Package size={14} />
                                                </div>
                                            </div>
                                            <p className="text-2xl font-black text-text-main drop-shadow-sm">{expiryDate}</p>
                                            <p className={`text-xs font-bold mt-2 ${planType === 'trial' ? 'text-warning' : 'text-success'}`}>
                                                {planType === 'trial' ? 'Trial expires soon' : 'Auto-renewal active'}
                                            </p>
                                        </div>
                                    </div>

                                    {/* Shipment Overview Data */}
                                    {whatsappConnected && (
                                        <div className="mt-12 space-y-6">
                                            <div className="flex items-center justify-between">
                                                <h3 className="text-sm font-black text-text-main uppercase tracking-[0.15em]">
                                                    Live Operations
                                                </h3>
                                                {fetchingStats && (
                                                    <span className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-text-muted">
                                                        <Loader2 size={12} className="animate-spin" /> Syncing
                                                    </span>
                                                )}
                                            </div>

                                            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                                                <div className="glass-panel p-6 border-border/50 bg-gradient-to-br from-surface to-surface/50 hover:shadow-xl hover:-translate-y-1 transition-all duration-300">
                                                    <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-4">Total Shipments</p>
                                                    <p className="text-4xl font-black text-text-main">{shipmentStats.total}</p>
                                                </div>
                                                <div className="glass-panel p-6 border-accent/20 bg-gradient-to-br from-accent/10 to-transparent hover:shadow-xl hover:shadow-accent/5 hover:-translate-y-1 transition-all duration-300">
                                                    <p className="text-[10px] font-black uppercase tracking-[0.2em] text-accent mb-4 flex items-center gap-2">
                                                        <span className="w-2 h-2 rounded-full bg-accent animate-pulse" /> Active Shipments
                                                    </p>
                                                    <p className="text-4xl font-black text-accent drop-shadow-sm">{shipmentStats.active}</p>
                                                </div>
                                                <div className="glass-panel p-6 border-success/20 bg-gradient-to-br from-success/10 to-transparent hover:shadow-xl hover:shadow-success/5 hover:-translate-y-1 transition-all duration-300">
                                                    <p className="text-[10px] font-black uppercase tracking-[0.2em] text-success mb-4 flex items-center gap-2">
                                                        <CheckCircle2 size={14} /> Delivered
                                                    </p>
                                                    <p className="text-4xl font-black text-success drop-shadow-sm">{shipmentStats.delivered}</p>
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* WHATSAPP TAB */}
                            {activeTab === 'whatsapp' && (
                                <div className="max-w-2xl mx-auto">
                                    <div className="glass-panel p-10 border-border/50 text-center relative overflow-hidden group">
                                        <div className="absolute inset-0 bg-gradient-to-b from-accent/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                                        <div className="mx-auto w-20 h-20 bg-accent/10 border border-accent/20 rounded-3xl flex items-center justify-center mb-8 shadow-inner shadow-accent/20">
                                            <Smartphone className="text-accent" size={36} />
                                        </div>
                                        <h2 className="text-2xl font-black text-text-main uppercase tracking-tighter mb-4">
                                            {whatsappConnected ? 'Bot is Operational' : 'Activate Tracking Bot'}
                                        </h2>
                                        <p className="text-base font-medium text-text-muted mb-10 max-w-md mx-auto leading-relaxed">
                                            {whatsappConnected
                                                ? 'Your WhatsApp tracking bot is actively monitoring messages and updating customers in real-time.'
                                                : 'Link your WhatsApp Business number to automate status updates and let customers track packages instantly.'}
                                        </p>
                                        {!whatsappConnected ? (
                                            <button
                                                onClick={() => {
                                                    setIsConnectModalOpen(true);
                                                }}
                                                className="btn-primary px-10 py-4 text-sm flex items-center justify-center gap-3 mx-auto shadow-lg shadow-accent/20 hover:shadow-accent/40 hover:-translate-y-1 transition-all duration-300"
                                            >
                                                Connect Number
                                            </button>
                                        ) : (
                                            <div className="inline-flex items-center gap-3 px-6 py-3 bg-success/10 border border-success/20 rounded-full text-success shadow-inner">
                                                <span className="relative flex h-3 w-3">
                                                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75"></span>
                                                    <span className="relative inline-flex rounded-full h-3 w-3 bg-success"></span>
                                                </span>
                                                <span className="text-sm font-black uppercase tracking-widest">Online & Monitoring</span>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            )}

                            {/* BILLING TAB */}
                            {activeTab === 'billing' && (
                                <div className="max-w-2xl mx-auto">
                                    <div className="glass-panel p-8 md:p-10 border-border/50">
                                        <div className="flex items-center gap-4 mb-8 pb-8 border-b border-border/50">
                                            <div className="w-12 h-12 rounded-2xl bg-surface border border-border flex items-center justify-center text-text-main">
                                                <CreditCard size={20} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-black text-text-main uppercase tracking-tighter">
                                                    Billing & Plans
                                                </h2>
                                                <p className="text-xs font-bold text-text-muted mt-1 uppercase tracking-widest">Manage your subscription</p>
                                            </div>
                                        </div>

                                        <div className="space-y-4">
                                            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center p-5 bg-surface/40 hover:bg-surface/60 transition-colors rounded-2xl border border-border/40 group">
                                                <div className="mb-4 sm:mb-0">
                                                    <p className="text-sm font-black text-text-main flex items-center gap-2">
                                                        {currentPlan.name} <span className="px-2 py-0.5 bg-accent/10 text-accent text-[9px] rounded-full uppercase tracking-widest">Current</span>
                                                    </p>
                                                    <p className="text-xs font-medium text-text-muted mt-1">
                                                        {planType === 'trial' ? 'Full access to all features' : 'Billed monthly'}
                                                    </p>
                                                </div>
                                                <p className="text-2xl font-black text-text-main">
                                                    {planType === 'trial' ? 'Free' : currentPlan.price}
                                                </p>
                                            </div>

                                            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center p-5 bg-surface/40 hover:bg-surface/60 transition-colors rounded-2xl border border-border/40">
                                                <div className="mb-4 sm:mb-0">
                                                    <p className="text-sm font-black text-text-main">{planType === 'trial' ? 'Trial Expiry' : 'Next Payment Cycle'}</p>
                                                    <p className={`text-xs font-medium mt-1 ${planType === 'trial' ? 'text-warning' : 'text-success'}`}>
                                                        {planType === 'trial' ? 'Upgrade to keep tracking' : 'Auto-renewal enabled'}
                                                    </p>
                                                </div>
                                                <p className="text-lg font-black text-text-main">{expiryDate}</p>
                                            </div>

                                            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center p-5 bg-surface/40 hover:bg-surface/60 transition-colors rounded-2xl border border-border/40">
                                                <div className="mb-4 sm:mb-0">
                                                    <p className="text-sm font-black text-text-main">Payment Method</p>
                                                    <p className="text-xs font-medium text-text-muted mt-1">Processed securely via Paystack</p>
                                                </div>
                                                <button disabled title="Coming soon" className="px-5 py-2.5 rounded-xl border border-border font-black text-xs uppercase tracking-widest bg-surface cursor-not-allowed opacity-50">
                                                    Update Card
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* SETTINGS TAB */}
                            {activeTab === 'settings' && (
                                <div className="max-w-2xl mx-auto space-y-8">
                                    <div className="glass-panel p-8 md:p-10 border-border/50">
                                        <div className="flex items-center gap-4 mb-8 pb-8 border-b border-border/50">
                                            <div className="w-12 h-12 rounded-2xl bg-surface border border-border flex items-center justify-center text-text-main">
                                                <Settings size={20} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-black text-text-main uppercase tracking-tighter">
                                                    Company Profile
                                                </h2>
                                                <p className="text-xs font-bold text-text-muted mt-1 uppercase tracking-widest">Update workspace details</p>
                                            </div>
                                        </div>

                                        <div className="space-y-8">
                                            {/* Logo URL */}
                                            <div className="space-y-3">
                                                <label className="text-[10px] font-black uppercase tracking-[0.2em] text-text-main ml-1 flex items-center gap-2">
                                                    Company Logo URL
                                                </label>
                                                <input
                                                    type="url"
                                                    value={settingsForm.logo_url}
                                                    onChange={(e) => setSettingsForm({ ...settingsForm, logo_url: e.target.value })}
                                                    className="input-premium w-full !bg-surface/50 focus:!bg-surface"
                                                    placeholder="https://example.com/logo.png"
                                                />
                                                {settingsForm.logo_url && (
                                                    <div className="mt-4 p-6 bg-surface/30 rounded-2xl border border-border/50 flex justify-center relative overflow-hidden">
                                                        <div className="absolute inset-0 bg-dot-grid opacity-[0.05]" />
                                                        {/* eslint-disable-next-line @next/next/no-img-element */}
                                                        <img
                                                            src={settingsForm.logo_url}
                                                            alt="Preview"
                                                            className="max-h-20 object-contain relative z-10 drop-shadow-md rounded-xl"
                                                            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                                                            onLoad={(e) => { (e.target as HTMLImageElement).style.display = 'block'; }}
                                                        />
                                                    </div>
                                                )}
                                            </div>

                                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                                                <div className="space-y-3">
                                                    <label className="text-[10px] font-black uppercase tracking-[0.2em] text-text-main ml-1">
                                                        Company Name
                                                    </label>
                                                    <input
                                                        type="text"
                                                        value={settingsForm.name}
                                                        onChange={(e) => setSettingsForm({ ...settingsForm, name: e.target.value })}
                                                        className="input-premium w-full !bg-surface/50 focus:!bg-surface"
                                                        placeholder="Enter company name"
                                                    />
                                                </div>
                                                <div className="space-y-3">
                                                    <label className="text-[10px] font-black uppercase tracking-[0.2em] text-text-main ml-1">
                                                        Admin Email
                                                    </label>
                                                    <input
                                                        type="email"
                                                        value={settingsForm.admin_email}
                                                        onChange={(e) => setSettingsForm({ ...settingsForm, admin_email: e.target.value })}
                                                        className="input-premium w-full !bg-surface/50 focus:!bg-surface"
                                                        placeholder="admin@company.com"
                                                    />
                                                </div>
                                            </div>

                                            <div className="pt-4 border-t border-border/50 flex justify-end">
                                                <button
                                                    onClick={handleSettingsSave}
                                                    disabled={isSaving}
                                                    className="btn-primary px-8 py-3.5 text-sm flex items-center justify-center gap-3 active:scale-95 transition-all shadow-lg shadow-accent/20"
                                                >
                                                    {isSaving ? <Loader2 className="animate-spin" size={16} /> : "Save Changes"}
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    {/* Danger Zone */}
                                    <div className="glass-panel p-8 md:p-10 border-error/20 bg-error/5 relative overflow-hidden group">
                                        <div className="absolute top-0 left-0 w-1 h-full bg-error" />
                                        <div className="flex items-start justify-between flex-col sm:flex-row gap-6">
                                            <div>
                                                <h3 className="text-sm font-black text-error uppercase tracking-[0.2em] mb-2 flex items-center gap-2">
                                                    <AlertTriangle size={16} /> Danger Zone
                                                </h3>
                                                <p className="text-sm font-medium text-text-muted leading-relaxed max-w-sm">
                                                    Disconnecting your WhatsApp bot stops all tracking instantly. You can also securely sign out of your account here.
                                                </p>
                                            </div>
                                            <div className="flex flex-col gap-3">
                                                <button
                                                    disabled
                                                    title="Coming soon"
                                                    className="px-6 py-3 bg-error/10 text-error/50 border border-error/10 rounded-xl font-black text-xs uppercase tracking-widest whitespace-nowrap shadow-sm cursor-not-allowed opacity-60"
                                                >
                                                    Disconnect Bot
                                                </button>
                                                <button
                                                    onClick={handleSignOut}
                                                    disabled={signingOut}
                                                    className="px-6 py-3 bg-surface hover:bg-surface-hover text-text-main border border-border rounded-xl font-black text-xs uppercase tracking-widest transition-all active:scale-95 whitespace-nowrap shadow-sm flex items-center justify-center gap-2"
                                                >
                                                    {signingOut ? <Loader2 size={16} className="animate-spin" /> : <LogOut size={16} />}
                                                    Sign Out
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </motion.div>
                    </AnimatePresence>
                </div>
            </div>

            {/* Connect WhatsApp Modal Component */}
            <WhatsAppConnectModal
                isOpen={isConnectModalOpen}
                onClose={() => setIsConnectModalOpen(false)}
                companyId={companyId || ''}
                companyData={companyData}
                onSuccess={() => fetchCompanyData()}
            />
        </div>
    );
}
