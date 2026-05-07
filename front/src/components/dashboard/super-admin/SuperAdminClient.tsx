"use client";

import { useState, useMemo, useEffect } from "react";
import {
    Shield,
    Building2,
    Users,
    Wifi,
    WifiOff,
    Search,
    ChevronDown,
    ChevronUp,
    CheckCircle2,
    XCircle,
    AlertTriangle,
    CreditCard,
    Clock,
    Eye,
    X,
    TrendingUp,
    History,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import type { SessionUser } from "@/lib/auth";
import TenantActions from "./TenantActions";
import AnalyticsTab from "./AnalyticsTab";
import AuditLogTab from "./AuditLogTab";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getApiUrl } from "@/lib/utils";
import { createClient } from "@/lib/supabase/client";

/* ------------------------------------------------------------------ */
/*  Types                                                              */
/* ------------------------------------------------------------------ */

interface Company {
    id: string;
    name: string | null;
    admin_email: string;
    whatsapp_phone: string | null;
    auth_status: string | null;
    subscription_status: string | null;
    subscription_expiry: string | null;
    plan_type: string | null;
    brand_color: string | null;
    logo_url: string | null;
    tracking_prefix: string | null;
    created_at: string | null;
}

interface SuperAdminClientProps {
    user: SessionUser;
    initialCompanies: Company[];
    jwt?: string;
}

/* ------------------------------------------------------------------ */
/*  Constants                                                          */
/* ------------------------------------------------------------------ */

type SortKey = "name" | "plan_type" | "created_at" | "subscription_status";

const STATUS_CONFIG: Record<string, { icon: typeof CheckCircle2; color: string; bg: string; label: string }> = {
    active: { icon: CheckCircle2, color: "text-success", bg: "bg-success/10 border-success/20", label: "Active" },
    trialing: { icon: Clock, color: "text-warning", bg: "bg-warning/10 border-warning/20", label: "Trial" },
    pending_linking: { icon: AlertTriangle, color: "text-warning", bg: "bg-warning/10 border-warning/20", label: "Pending" },
    expired: { icon: XCircle, color: "text-error", bg: "bg-error/10 border-error/20", label: "Expired" },
    suspended: { icon: XCircle, color: "text-error", bg: "bg-error/10 border-error/20", label: "Suspended" },
};

const PLAN_LABELS: Record<string, string> = {
    trial: "Trial",
    starter: "Starter",
    pro: "Professional",
    enterprise: "Enterprise",
};

/* ------------------------------------------------------------------ */
/*  Micro-components                                                   */
/* ------------------------------------------------------------------ */

function StatusPill({ status }: { status: string | null }) {
    const s = status || "pending_linking";
    const cfg = STATUS_CONFIG[s] || STATUS_CONFIG.pending_linking;
    const Icon = cfg.icon;
    return (
        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest ${cfg.color} ${cfg.bg}`}>
            <Icon size={10} />
            {cfg.label}
        </span>
    );
}

function BotDot({ connected }: { connected: boolean }) {
    return (
        <span className={`inline-flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest ${connected ? "text-success" : "text-text-muted"}`}>
            {connected ? <Wifi size={12} /> : <WifiOff size={12} />}
            {connected ? "Online" : "Offline"}
        </span>
    );
}

interface StatCardProps {
    icon: React.ComponentType<{ size?: number }>;
    label: string;
    value: string | number;
    accent?: string;
}

function StatCard({ icon: Icon, label, value, accent }: StatCardProps) {
    return (
        <div className="glass-panel p-6 border-border/40 hover:border-border hover:shadow-xl hover:shadow-black/5 hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
            <div className={`absolute top-0 right-0 w-24 h-24 rounded-full blur-2xl transition-all ${accent || "bg-primary/5 group-hover:bg-primary/10"}`} />
            <div className="flex items-center justify-between mb-4">
                <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">{label}</p>
                <div className="w-8 h-8 rounded-full bg-surface border border-border flex items-center justify-center text-text-main">
                    <Icon size={14} />
                </div>
            </div>
            <p className="text-3xl font-black text-text-main drop-shadow-sm">{value}</p>
        </div>
    );
}

/* ------------------------------------------------------------------ */
/*  Detail Modal                                                       */
/* ------------------------------------------------------------------ */

function CompanyDetailModal({ company, onClose, onRefresh }: { company: Company; onClose: () => void; onRefresh: () => void }) {
    const created = company.created_at ? new Date(company.created_at).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" }) : "—";
    const expiry = company.subscription_expiry ? new Date(company.subscription_expiry).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" }) : "—";
    const planLabel = PLAN_LABELS[(company.plan_type || "trial").toLowerCase()] || company.plan_type || "Trial";
    const botConnected = company.auth_status === "active";
    const initials = (company.name || company.admin_email).substring(0, 2).toUpperCase();

    return (
        <AnimatePresence>
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="fixed inset-0 z-[10000] flex items-center justify-center p-4"
                onClick={onClose}
            >
                {/* Backdrop */}
                <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

                <motion.div
                    initial={{ opacity: 0, scale: 0.95, y: 20 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95, y: 20 }}
                    transition={{ type: "spring", stiffness: 400, damping: 30 }}
                    onClick={(e) => e.stopPropagation()}
                    className="relative w-full max-w-lg glass-panel p-0 overflow-hidden"
                >
                    {/* Header Gradient Bar */}
                    <div className="h-1.5 bg-gradient-to-r from-accent via-accent-deep to-primary" />

                    <div className="p-8">
                        {/* Close */}
                        <button
                            onClick={onClose}
                            className="absolute top-6 right-6 w-10 h-10 rounded-xl bg-surface-muted hover:bg-error/10 hover:text-error flex items-center justify-center transition-all active:scale-95"
                            aria-label="Close detail modal"
                        >
                            <X size={18} />
                        </button>

                        {/* Avatar + Name */}
                        <div className="flex items-center gap-5 mb-8">
                            <div
                                className="w-16 h-16 rounded-2xl flex items-center justify-center shadow-lg border border-white/10 shrink-0"
                                style={{ background: `linear-gradient(135deg, ${company.brand_color || "var(--color-accent)"}, var(--color-primary))` }}
                            >
                                <span className="text-xl font-black text-white">{initials}</span>
                            </div>
                            <div className="min-w-0">
                                <h2 className="text-2xl font-black text-text-main uppercase tracking-tight truncate">
                                    {company.name || "Unnamed Company"}
                                </h2>
                                <p className="text-xs font-bold text-text-muted truncate">{company.admin_email}</p>
                            </div>
                        </div>

                        {/* Detail Grid */}
                        <div className="grid grid-cols-2 gap-4">
                            <DetailField label="Plan" value={planLabel} />
                            <DetailField label="Subscription" value={<StatusPill status={company.subscription_status} />} />
                            <DetailField label="Bot Status" value={<BotDot connected={botConnected} />} />
                            <DetailField label="WhatsApp" value={company.whatsapp_phone || "Not linked"} />
                            <DetailField label="Tracking Prefix" value={company.tracking_prefix || "—"} />
                            <DetailField label="Created" value={created} />
                            <DetailField label="Expiry" value={expiry} />
                            <DetailField label="Company ID" value={<span className="text-[10px] font-mono break-all">{company.id}</span>} />
                        </div>

                        {/* Tenant Lifecycle Actions */}
                        <TenantActions
                            company={company}
                            onActionComplete={() => {
                                onRefresh();
                                onClose();
                            }}
                        />
                    </div>
                </motion.div>
            </motion.div>
        </AnimatePresence>
    );
}

function DetailField({ label, value }: { label: string; value: React.ReactNode }) {
    return (
        <div className="space-y-1">
            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">{label}</p>
            <div className="text-sm font-bold text-text-main">{value}</div>
        </div>
    );
}

/* ------------------------------------------------------------------ */
/*  Main Component                                                     */
/* ------------------------------------------------------------------ */

export default function SuperAdminClient({ user, initialCompanies, jwt }: SuperAdminClientProps) {
    const router = useRouter();
    const queryClient = useQueryClient();

    const { data: companies = initialCompanies } = useQuery({
        queryKey: ["super-admin-companies"],
        queryFn: async () => {
            const res = await fetch(`${getApiUrl()}/api/super-admin/companies`, {
                credentials: "include",
            });
            if (!res.ok) throw new Error("Failed to fetch companies");
            return res.json() as Promise<Company[]>;
        },
        initialData: initialCompanies,
        staleTime: 60_000,
    });

    // --- REALTIME OVERSEER SUBSCRIPTION ---
    const supabase = useMemo(() => createClient(jwt), [jwt]);

    useEffect(() => {
        const channel = supabase
            .channel('super-admin-global')
            .on('postgres_changes', { event: '*', schema: 'public', table: 'companies' }, () => {
                queryClient.invalidateQueries({ queryKey: ["super-admin-companies"] });
            })
            .subscribe();

        return () => {
            supabase.removeChannel(channel);
        };
    }, [supabase, queryClient]);

    const [searchQuery, setSearchQuery] = useState("");
    const [sortKey, setSortKey] = useState<SortKey>("created_at");
    const [sortAsc, setSortAsc] = useState(false);
    const [selectedCompany, setSelectedCompany] = useState<Company | null>(null);
    const [filterStatus, setFilterStatus] = useState<string>("all");
    const [activeTab, setActiveTab] = useState<"tenants" | "intelligence" | "audit">("tenants");

    /* ---------- Derived data ---------- */

    const filtered = useMemo(() => {
        let list = [...companies];

        // Text search
        if (searchQuery.trim()) {
            const q = searchQuery.toLowerCase();
            list = list.filter(
                (c) =>
                    c.name?.toLowerCase().includes(q) ||
                    c.admin_email.toLowerCase().includes(q) ||
                    c.tracking_prefix?.toLowerCase().includes(q)
            );
        }

        // Status filter
        if (filterStatus !== "all") {
            if (filterStatus === "bot_online") {
                list = list.filter((c) => c.auth_status === "active");
            } else if (filterStatus === "bot_offline") {
                list = list.filter((c) => c.auth_status !== "active");
            } else {
                list = list.filter((c) => (c.subscription_status || "active") === filterStatus);
            }
        }

        // Sort
        list.sort((a, b) => {
            const aVal = (a[sortKey] || "").toString().toLowerCase();
            const bVal = (b[sortKey] || "").toString().toLowerCase();
            return sortAsc ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
        });

        return list;
    }, [companies, searchQuery, sortKey, sortAsc, filterStatus]);

    /* ---------- Aggregate stats ---------- */

    const stats = useMemo(() => {
        const total = companies.length;
        const active = companies.filter((c) => c.subscription_status === "active" || c.subscription_status === "trialing").length;
        const botsOnline = companies.filter((c) => c.auth_status === "active").length;
        const expired = companies.filter((c) => {
            if (!c.subscription_expiry) return false;
            return new Date(c.subscription_expiry).getTime() < Date.now();
        }).length;
        return { total, active, botsOnline, expired };
    }, [companies]);

    /* ---------- Handlers ---------- */

    function handleSort(key: SortKey) {
        if (sortKey === key) {
            setSortAsc(!sortAsc);
        } else {
            setSortKey(key);
            setSortAsc(true);
        }
    }

    const SortIcon = sortAsc ? ChevronUp : ChevronDown;

    /* ---------- Render ---------- */

    return (
        <div className="pb-32 md:pb-24 relative bg-background overflow-x-hidden">
            <div className="max-w-7xl mx-auto z-10 relative pt-24 md:pt-32 px-4 sm:px-8">
                {/* ---- Header ---- */}
                <motion.header
                    initial={{ y: -20, opacity: 0 }}
                    animate={{ y: 0, opacity: 1 }}
                    className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12"
                >
                    <div className="flex items-center gap-5">
                        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-accent to-accent-deep flex items-center justify-center shadow-lg shadow-accent/20 border border-white/10 shrink-0">
                            <Shield className="text-white" size={28} />
                        </div>
                        <div className="min-w-0">
                            <h1 className="text-3xl sm:text-4xl font-black text-text-main uppercase tracking-tighter drop-shadow-sm">
                                Super Admin
                            </h1>
                            <p className="text-xs font-bold text-text-muted mt-1 uppercase tracking-widest flex items-center gap-2">
                                <span className="w-2 h-2 rounded-full bg-accent animate-pulse inline-block" />
                                Platform Control Center
                            </p>
                        </div>
                    </div>
                    <span className="text-xs font-bold text-text-muted uppercase tracking-widest">
                        {user.email}
                    </span>
                </motion.header>

                {/* ---- Stats Grid ---- */}
                <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 }}
                    className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-12"
                >
                    <StatCard icon={Building2} label="Total Tenants" value={stats.total} />
                    <StatCard icon={Users} label="Active Subscriptions" value={stats.active} accent="bg-success/5 group-hover:bg-success/10" />
                    <StatCard icon={Wifi} label="Bots Online" value={stats.botsOnline} accent="bg-accent/5 group-hover:bg-accent/10" />
                    <StatCard icon={AlertTriangle} label="Expired" value={stats.expired} accent="bg-error/5 group-hover:bg-error/10" />
                </motion.div>

                {/* ---- Tabs ---- */}
                <div className="flex items-center gap-1 p-1 bg-surface-muted/50 rounded-2xl border border-border/40 mb-8 w-fit">
                    <TabBtn active={activeTab === 'tenants'} onClick={() => setActiveTab('tenants')} icon={Building2} label="Tenants" />
                    <TabBtn active={activeTab === 'intelligence'} onClick={() => setActiveTab('intelligence')} icon={TrendingUp} label="Intelligence" />
                    <TabBtn active={activeTab === 'audit'} onClick={() => setActiveTab('audit')} icon={History} label="Audit Logs" />
                </div>

                <AnimatePresence mode="wait">
                    {activeTab === 'tenants' && (
                        <motion.div
                            key="tenants"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                        >
                            {/* ---- Toolbar ---- */}
                            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 mb-8">
                                {/* Search */}
                                <div className="relative flex-1">
                                    <Search size={16} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none" />
                                    <input
                                        type="text"
                                        value={searchQuery}
                                        onChange={(e) => setSearchQuery(e.target.value)}
                                        placeholder="Search by name, email, or prefix…"
                                        className="input-premium w-full pl-11 pr-4 !py-3 !rounded-xl text-sm"
                                    />
                                </div>

                                {/* Status Filter */}
                                <select
                                    value={filterStatus}
                                    onChange={(e) => setFilterStatus(e.target.value)}
                                    className="input-premium !py-3 !rounded-xl text-sm max-w-[200px] cursor-pointer"
                                >
                                    <option value="all">All Statuses</option>
                                    <option value="active">Active</option>
                                    <option value="trialing">Trialing</option>
                                    <option value="expired">Expired</option>
                                    <option value="suspended">Suspended</option>
                                    <option value="bot_online">Bot Online</option>
                                    <option value="bot_offline">Bot Offline</option>
                                </select>
                            </div>

                            {/* ---- Companies Table ---- */}
                            <div className="glass-panel overflow-hidden">
                                {/* Desktop Table */}
                                <div className="hidden md:block overflow-x-auto">
                                    <table className="w-full text-left">
                                        <thead>
                                            <tr className="border-b border-border/50">
                                                {([
                                                    ["name", "Company"],
                                                    ["plan_type", "Plan"],
                                                    ["subscription_status", "Status"],
                                                    ["created_at", "Created"],
                                                ] as [SortKey, string][]).map(([key, label]) => (
                                                    <th
                                                        key={key}
                                                        onClick={() => handleSort(key)}
                                                        className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted cursor-pointer select-none hover:text-text-main transition-colors"
                                                    >
                                                        <span className="inline-flex items-center gap-1.5">
                                                            {label}
                                                            {sortKey === key && <SortIcon size={12} />}
                                                        </span>
                                                    </th>
                                                ))}
                                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Bot</th>
                                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted text-right">Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {filtered.length === 0 ? (
                                                <tr>
                                                    <td colSpan={6} className="text-center py-16 text-text-muted text-sm font-medium">
                                                        No companies match your search.
                                                    </td>
                                                </tr>
                                            ) : (
                                                filtered.map((company, i) => {
                                                    const initials = (company.name || company.admin_email).substring(0, 2).toUpperCase();
                                                    const botConnected = company.auth_status === "active";
                                                    const planLabel = PLAN_LABELS[(company.plan_type || "trial").toLowerCase()] || company.plan_type || "Trial";
                                                    const created = company.created_at
                                                        ? new Date(company.created_at).toLocaleDateString("en-US", { month: "short", day: "numeric" })
                                                        : "—";

                                                    return (
                                                        <motion.tr
                                                            key={company.id}
                                                            initial={{ opacity: 0, y: 8 }}
                                                            animate={{ opacity: 1, y: 0 }}
                                                            transition={{ delay: 0.03 * i }}
                                                            className="border-b border-border/30 last:border-none hover:bg-surface-muted/50 transition-colors"
                                                        >
                                                            {/* Company */}
                                                            <td className="px-6 py-4">
                                                                <div className="flex items-center gap-3">
                                                                    <div
                                                                        className="w-9 h-9 rounded-xl flex items-center justify-center text-white text-xs font-black shrink-0 shadow-sm"
                                                                        style={{ background: company.brand_color || "var(--color-accent)" }}
                                                                    >
                                                                        {initials}
                                                                    </div>
                                                                    <div className="min-w-0">
                                                                        <p className="text-sm font-bold text-text-main truncate max-w-[200px]">
                                                                            {company.name || "Unnamed"}
                                                                        </p>
                                                                        <p className="text-[11px] text-text-muted truncate max-w-[200px]">
                                                                            {company.admin_email}
                                                                        </p>
                                                                    </div>
                                                                </div>
                                                            </td>

                                                            {/* Plan */}
                                                            <td className="px-6 py-4">
                                                                <span className="inline-flex items-center gap-1.5 text-xs font-bold text-text-main">
                                                                    <CreditCard size={12} className="text-text-muted" />
                                                                    {planLabel}
                                                                </span>
                                                            </td>

                                                            {/* Status */}
                                                            <td className="px-6 py-4">
                                                                <StatusPill status={company.subscription_status} />
                                                            </td>

                                                            {/* Created */}
                                                            <td className="px-6 py-4 text-xs font-medium text-text-muted">{created}</td>

                                                            {/* Bot */}
                                                            <td className="px-6 py-4">
                                                                <BotDot connected={botConnected} />
                                                            </td>

                                                            {/* Actions */}
                                                            <td className="px-6 py-4">
                                                                <div className="flex items-center justify-end gap-2">
                                                                    <button
                                                                        onClick={() => setSelectedCompany(company)}
                                                                        className="w-9 h-9 rounded-xl bg-surface-muted hover:bg-accent/10 hover:text-accent flex items-center justify-center transition-all active:scale-95"
                                                                        aria-label={`View details for ${company.name}`}
                                                                    >
                                                                        <Eye size={16} />
                                                                    </button>
                                                                </div>
                                                            </td>
                                                        </motion.tr>
                                                    );
                                                })
                                            )}
                                        </tbody>
                                    </table>
                                </div>

                                {/* Mobile Card List */}
                                <div className="md:hidden divide-y divide-border/30">
                                    {filtered.length === 0 ? (
                                        <div className="text-center py-16 text-text-muted text-sm font-medium">
                                            No companies match your search.
                                        </div>
                                    ) : (
                                        filtered.map((company, i) => {
                                            const initials = (company.name || company.admin_email).substring(0, 2).toUpperCase();
                                            const botConnected = company.auth_status === "active";
                                            const planLabel = PLAN_LABELS[(company.plan_type || "trial").toLowerCase()] || company.plan_type || "Trial";

                                            return (
                                                <motion.div
                                                    key={company.id}
                                                    initial={{ opacity: 0, y: 8 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ delay: 0.03 * i }}
                                                    className="p-5 flex items-center gap-4 active:bg-surface-muted/50 transition-colors cursor-pointer"
                                                    onClick={() => setSelectedCompany(company)}
                                                >
                                                    <div
                                                        className="w-11 h-11 rounded-xl flex items-center justify-center text-white text-xs font-black shrink-0 shadow-sm"
                                                        style={{ background: company.brand_color || "var(--color-accent)" }}
                                                    >
                                                        {initials}
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <p className="text-sm font-bold text-text-main truncate">{company.name || "Unnamed"}</p>
                                                        <div className="flex items-center gap-3 mt-1">
                                                            <StatusPill status={company.subscription_status} />
                                                            <BotDot connected={botConnected} />
                                                        </div>
                                                    </div>
                                                    <span className="text-[10px] font-black uppercase tracking-widest text-text-muted shrink-0">{planLabel}</span>
                                                </motion.div>
                                            );
                                        })
                                    )}
                                </div>

                                {/* Footer */}
                                <div className="px-6 py-4 border-t border-border/30 flex items-center justify-between">
                                    <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">
                                        {filtered.length} of {companies.length} tenants
                                    </p>
                                </div>
                            </div>
                        </motion.div>
                    )}

                    {activeTab === 'intelligence' && (
                        <motion.div
                            key="intelligence"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                        >
                            <AnalyticsTab />
                        </motion.div>
                    )}

                    {activeTab === 'audit' && (
                        <motion.div
                            key="audit"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                        >
                            <AuditLogTab />
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>

            {/* Detail Modal */}
            {selectedCompany && (
                <CompanyDetailModal
                    company={selectedCompany}
                    onClose={() => setSelectedCompany(null)}
                    onRefresh={() => {
                        queryClient.invalidateQueries({ queryKey: ["super-admin-companies"] });
                        router.refresh();
                    }}
                />
            )}
        </div>
    );
}

interface TabBtnProps {
    active: boolean;
    onClick: () => void;
    icon: React.ComponentType<{ size?: number; className?: string }>;
    label: string;
}

function TabBtn({ active, onClick, icon: Icon, label }: TabBtnProps) {
    return (
        <button
            onClick={onClick}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all ${active
                    ? "bg-surface text-text-main shadow-sm border border-border/50"
                    : "text-text-muted hover:text-text-main"
                }`}
        >
            <Icon size={14} className={active ? "text-accent" : ""} />
            {label}
        </button>
    );
}
