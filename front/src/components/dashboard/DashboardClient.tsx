"use client";

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
    Package, Smartphone,
    CheckCircle2, AlertTriangle, XCircle,
    UserCircle2, Activity
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { createClient } from '@/lib/supabase/client';
import WhatsAppConnectModal from '@/components/dashboard/WhatsAppConnectModal';
import { OverviewTab } from './OverviewTab';
import { WhatsAppTab } from './WhatsAppTab';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import Image from 'next/image';

type Tab = 'overview' | 'whatsapp';

export interface CompanyData {
    name: string;
    admin_email: string;
    subscription_status: string;
    subscription_expiry: string;
    plan_type: string;
    auth_status: string;
    whatsapp_phone: string;
    brand_color: string;
    logo_url: string;
    tracking_prefix?: string;
}

interface DashboardClientProps {
    initialCompanyData: CompanyData | null;
    initialStats: { total: number; active: number; delivered: number };
    user: { email: string; company_name: string; plan_type: string } | null;
    companyId: string;
}

// --- CONSTANTS & EXTRACTED COMPONENTS (Phase 1 Fixes) ---
const TABS: { id: Tab; icon: typeof Package; label: string }[] = [
    { id: 'overview', icon: Activity, label: 'Overview' },
    { id: 'whatsapp', icon: Smartphone, label: 'WhatsApp' },
];

const PLAN_DETAILS: Record<string, { name: string; price: string }> = {
    trial: { name: '7 Days Free Trial', price: '₦0' },
    basic: { name: 'Basic Plan', price: '₦10,000' },
    pro: { name: 'Professional Plan', price: '₦25,000' },
    enterprise: { name: 'Enterprise Plan', price: '₦50,000' },
};

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

export default function DashboardClient({ initialCompanyData, initialStats, user, companyId }: DashboardClientProps) {
    const router = useRouter();
    const queryClient = useQueryClient();
    const [activeTab, setActiveTab] = useState<Tab>('overview');
    const [isConnectModalOpen, setIsConnectModalOpen] = useState(false);

    // Stable supabase browser client (createBrowserClient is a singleton internally)
    const supabase = createClient();

    // --- PHASE 2: REACT QUERY FETCHING ---
    const { data: companyData, isError: companyError } = useQuery({
        queryKey: ['company', companyId],
        queryFn: async () => {
            if (!companyId) return null;
            const { data, error } = await supabase
                .from('companies')
                .select('name, admin_email, subscription_status, subscription_expiry, plan_type, auth_status, whatsapp_phone, brand_color, logo_url, tracking_prefix')
                .eq('id', companyId)
                .single();

            if (error) throw error;
            return data as CompanyData;
        },
        initialData: initialCompanyData,
        staleTime: 1000 * 60 * 5, // 5 minutes
    });

    const { data: shipmentStats, isFetching: fetchingStats } = useQuery({
        queryKey: ['shipments', companyId],
        queryFn: async () => {
            if (!companyId) return { total: 0, active: 0, delivered: 0 };
            const { data, error } = await supabase
                .from('shipment')
                .select('status')
                .eq('company_id', companyId);

            if (error) throw error;

            const active = data.filter(s => s.status !== 'DELIVERED' && s.status !== 'CANCELED').length;
            const delivered = data.filter(s => s.status === 'DELIVERED').length;
            return { total: data.length, active, delivered };
        },
        initialData: initialStats,
        staleTime: 1000 * 60 * 5, // 5 minutes
    });

    // Handle Authorization Errors
    useEffect(() => {
        if (companyError) {
            router.push('/auth');
        }
    }, [companyError, router]);

    // --- REALTIME SUBSCRIPTIONS (Invalidating React Query) ---
    useEffect(() => {
        if (!companyId) return;

        const companyChannel = supabase
            .channel(`company-global-${companyId}`)
            .on('postgres_changes', { event: 'UPDATE', schema: 'public', table: 'companies', filter: `id=eq.${companyId}` }, () => {
                queryClient.invalidateQueries({ queryKey: ['company', companyId] });
            });

        const statsChannel = supabase
            .channel(`shipment-stats-${companyId}`)
            .on('postgres_changes', { event: '*', schema: 'public', table: 'shipment', filter: `company_id=eq.${companyId}` }, () => {
                queryClient.invalidateQueries({ queryKey: ['shipments', companyId] });
            });

        companyChannel.subscribe();
        statsChannel.subscribe();

        return () => {
            supabase.removeChannel(companyChannel);
            supabase.removeChannel(statsChannel);
        };
    }, [companyId, supabase, queryClient]);


    // --- DERIVED STATE ---
    const companyName = companyData?.name || user?.company_name || 'CARGOHIVE';
    const subscriptionStatus = companyData?.subscription_status || 'active';
    const whatsappConnected = companyData?.auth_status === 'active';
    const overallStatus = whatsappConnected ? subscriptionStatus : 'pending';

    const planType = (companyData?.plan_type || 'trial').toLowerCase();
    const currentPlan = PLAN_DETAILS[planType] || PLAN_DETAILS['trial'];

    const expiryDate = companyData?.subscription_expiry
        ? new Date(companyData.subscription_expiry).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
        : '—';

    return (
        <div className="pb-32 md:pb-24 relative bg-background overflow-x-hidden">
            <div className="max-w-6xl mx-auto z-10 relative pt-24 md:pt-32 px-4 sm:px-8">
                <motion.header
                    initial={{ y: -20, opacity: 0 }}
                    animate={{ y: 0, opacity: 1 }}
                    className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12"
                >
                    <div className="flex items-center gap-6">
                        {companyData?.logo_url ? (
                            <div className="relative w-16 h-16 rounded-2xl shadow-lg border border-border overflow-hidden">
                                <Image src={companyData.logo_url} alt="Logo" fill className="object-cover" />
                            </div>
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
                </motion.header>

                {/* Phase 3: A11y Tab Navigation */}
                <div
                    role="tablist"
                    aria-label="Dashboard Navigation"
                    className="flex flex-wrap gap-2 sm:gap-4 border-b border-border/50 mb-10 pb-2 relative"
                >
                    {TABS.map((tab) => {
                        const isActive = activeTab === tab.id;
                        return (
                            <button
                                key={tab.id}
                                role="tab"
                                aria-selected={isActive}
                                aria-controls={`panel-${tab.id}`}
                                tabIndex={isActive ? 0 : -1}
                                onClick={() => setActiveTab(tab.id)}
                                onKeyDown={(e) => {
                                    if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
                                        const currentIndex = TABS.findIndex(t => t.id === activeTab);
                                        const nextIndex = e.key === 'ArrowRight'
                                            ? (currentIndex + 1) % TABS.length
                                            : (currentIndex - 1 + TABS.length) % TABS.length;
                                        setActiveTab(TABS[nextIndex].id);
                                    }
                                }}
                                className={`relative flex items-center gap-2 px-4 py-3 font-black text-xs uppercase tracking-widest transition-colors whitespace-nowrap rounded-xl focus:outline-none focus-visible:ring-2 focus-visible:ring-accent ${isActive ? 'text-text-main' : 'text-text-muted hover:text-text-main hover:bg-surface'
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

                <div className="relative min-h-[400px]">
                    <AnimatePresence mode="wait">
                        <motion.div
                            key={activeTab}
                            role="tabpanel"
                            id={`panel-${activeTab}`}
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            transition={{ duration: 0.2 }}
                        >
                            {activeTab === 'overview' && (
                                <OverviewTab
                                    companyData={companyData || null}
                                    shipmentStats={shipmentStats}
                                    fetchingStats={fetchingStats}
                                    whatsappConnected={whatsappConnected}
                                    planType={planType}
                                    currentPlan={currentPlan}
                                    expiryDate={expiryDate}
                                    onConnectClick={() => setIsConnectModalOpen(true)}
                                />
                            )}

                            {activeTab === 'whatsapp' && (
                                <WhatsAppTab
                                    whatsappConnected={whatsappConnected}
                                    whatsappPhone={companyData?.whatsapp_phone}
                                    companyId={companyId}
                                    onConnectClick={() => setIsConnectModalOpen(true)}
                                />
                            )}
                        </motion.div>
                    </AnimatePresence>
                </div>
            </div>

            <WhatsAppConnectModal
                isOpen={isConnectModalOpen}
                onClose={() => setIsConnectModalOpen(false)}
                companyId={companyId || ''}
                companyData={companyData || null}
                onSuccess={() => queryClient.invalidateQueries({ queryKey: ['company', companyId] })}
            />
        </div>
    );
}
