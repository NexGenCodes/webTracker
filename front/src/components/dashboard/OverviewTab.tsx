import React from 'react';
import { Package, Wifi, WifiOff, CreditCard, AlertTriangle, ChevronRight, CheckCircle2, Loader2 } from 'lucide-react';
import { CompanyData } from './DashboardClient';

interface OverviewTabProps {
    companyData: CompanyData | null;
    shipmentStats: { total: number; active: number; delivered: number } | undefined;
    fetchingStats: boolean;
    whatsappConnected: boolean;
    planType: string;
    currentPlan: { name: string; price: string };
    expiryDate: string;
    daysRemaining: number;
    isExpired: boolean;
    onConnectClick: () => void;
}

export function OverviewTab({
    shipmentStats,
    fetchingStats,
    whatsappConnected,
    planType,
    currentPlan,
    expiryDate,
    daysRemaining,
    isExpired,
    onConnectClick
}: OverviewTabProps) {
    return (
        <div className="space-y-8">
            {isExpired && (
                <div className="relative overflow-hidden glass-panel p-6 border-error/40 bg-error/5 flex flex-col sm:flex-row items-start sm:items-center gap-6 group hover:shadow-lg hover:shadow-error/10 transition-all duration-300">
                    <div className="absolute top-0 left-0 w-1 h-full bg-error" />
                    <div className="w-14 h-14 bg-error/20 rounded-2xl flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform duration-300">
                        <AlertTriangle className="text-error" size={28} />
                    </div>
                    <div className="flex-1">
                        <h3 className="text-base font-black text-text-main uppercase tracking-widest mb-1">
                            Subscription Expired
                        </h3>
                        <p className="text-sm font-medium text-text-muted">
                            Your access has been suspended because your trial or subscription has ended. Please upgrade your plan to continue using the service.
                        </p>
                    </div>
                    <a
                        href="/dashboard/billing"
                        className="btn-primary px-6 py-3.5 text-xs flex items-center gap-2 active:scale-95 shrink-0 shadow-lg shadow-error/20 !bg-error hover:!bg-error/90 !text-black"
                    >
                        Upgrade Now <ChevronRight size={16} />
                    </a>
                </div>
            )}

            {!whatsappConnected && !isExpired && (
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
                        onClick={onConnectClick}
                        className="btn-primary px-6 py-3.5 text-xs flex items-center gap-2 active:scale-95 shrink-0 shadow-lg shadow-warning/20 !bg-warning hover:!bg-warning/90 !text-black"
                    >
                        Connect Now <ChevronRight size={16} />
                    </button>
                </div>
            )}

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
                        <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">
                            {planType === 'trial' ? 'Trial Time Left' : 'Next Billing'}
                        </p>
                        <div className="w-8 h-8 rounded-full bg-surface border border-border flex items-center justify-center text-text-main">
                            <Package size={14} />
                        </div>
                    </div>
                    <p className={`text-2xl font-black drop-shadow-sm ${isExpired ? 'text-error' : 'text-text-main'}`}>
                        {isExpired ? 'Expired' : (planType === 'trial' ? `${daysRemaining} Days` : expiryDate)}
                    </p>
                    <p className={`text-xs font-bold mt-2 ${planType === 'trial' || isExpired ? (isExpired ? 'text-error' : 'text-warning') : 'text-success'}`}>
                        {isExpired ? 'Renew immediately' : (planType === 'trial' ? 'Upgrade to keep access' : 'Auto-renewal active')}
                    </p>
                </div>
            </div>

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
                            <p className="text-4xl font-black text-text-main">{shipmentStats?.total}</p>
                        </div>
                        <div className="glass-panel p-6 border-accent/20 bg-gradient-to-br from-accent/10 to-transparent hover:shadow-xl hover:shadow-accent/5 hover:-translate-y-1 transition-all duration-300">
                            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-accent mb-4 flex items-center gap-2">
                                <span className="w-2 h-2 rounded-full bg-accent animate-pulse" /> Active Shipments
                            </p>
                            <p className="text-4xl font-black text-accent drop-shadow-sm">{shipmentStats?.active}</p>
                        </div>
                        <div className="glass-panel p-6 border-success/20 bg-gradient-to-br from-success/10 to-transparent hover:shadow-xl hover:shadow-success/5 hover:-translate-y-1 transition-all duration-300">
                            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-success mb-4 flex items-center gap-2">
                                <CheckCircle2 size={14} /> Delivered
                            </p>
                            <p className="text-4xl font-black text-success drop-shadow-sm">{shipmentStats?.delivered}</p>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
