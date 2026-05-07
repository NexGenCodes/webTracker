"use client";

import React, { useEffect, useState } from "react";
import {
    BarChart3,
    PieChart,
    TrendingUp,
    ShipWheel,
    RefreshCcw
} from "lucide-react";
import { getApiUrl } from "@/lib/utils";
import { motion } from "framer-motion";

interface AnalyticsData {
    total_tenants: number;
    new_tenants_this_month: number;
    total_shipments: number;
    shipments_today: number;
    plan_distribution: Record<string, number>;
    subscription_distribution: Record<string, number>;
}

export default function AnalyticsTab() {
    const [data, setData] = useState<AnalyticsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchAnalytics = async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await fetch(`${getApiUrl()}/api/super-admin/analytics`, {
                credentials: "include",
            });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const json = await res.json();
            setData(json);
        } catch (err) {
            console.error("Failed to fetch analytics", err);
            setError(err instanceof Error ? err.message : "Failed to load analytics");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchAnalytics();
    }, []);

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
                <RefreshCcw className="animate-spin text-accent" size={32} />
                <p className="text-xs font-black uppercase tracking-widest text-text-muted">Loading Intelligence…</p>
            </div>
        );
    }

    if (!data) {
        return (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
                <BarChart3 className="text-error" size={32} />
                <p className="text-xs font-black uppercase tracking-widest text-error">
                    {error || "No data available"}
                </p>
                <button
                    onClick={fetchAnalytics}
                    className="px-4 py-2 bg-surface border border-border rounded-xl text-xs font-black uppercase tracking-widest hover:bg-surface-muted transition-all active:scale-95"
                >
                    Retry
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            <div className="flex items-center justify-between">
                <h2 className="text-xs font-black uppercase tracking-[0.2em] text-text-muted flex items-center gap-2">
                    <BarChart3 size={14} className="text-accent" />
                    Platform Intelligence
                </h2>
                <button
                    onClick={fetchAnalytics}
                    disabled={loading}
                    className="p-2 rounded-lg bg-surface border border-border/50 hover:border-border transition-all active:rotate-180 duration-500 disabled:opacity-50"
                >
                    <RefreshCcw size={16} className={loading ? "animate-spin" : ""} />
                </button>
            </div>

            {/* Stat Cards */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
                <AnalyticStat
                    label="Growth (Monthly)"
                    value={`+${data.new_tenants_this_month}`}
                    sub="New companies"
                    icon={TrendingUp}
                    color="text-success"
                />
                <AnalyticStat
                    label="Total Shipments"
                    value={data.total_shipments.toLocaleString()}
                    sub="Lifetime volume"
                    icon={ShipWheel}
                    color="text-accent"
                />
                <AnalyticStat
                    label="Shipments Today"
                    value={data.shipments_today}
                    sub="Active operations"
                    icon={TrendingUp}
                    color="text-primary"
                />
                <AnalyticStat
                    label="Est. Monthly Revenue"
                    value={calculateRevenue(data.plan_distribution)}
                    sub="Gross estimation"
                    icon={BarChart3}
                    color="text-warning"
                />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Plan Distribution */}
                <div className="glass-panel p-8">
                    <h3 className="text-xs font-black uppercase tracking-widest text-text-muted mb-8 flex items-center gap-2">
                        <PieChart size={14} className="text-accent" />
                        Plan Distribution
                    </h3>
                    <div className="space-y-4">
                        {Object.entries(data.plan_distribution || {}).map(([plan, count]) => (
                            <DistributionBar
                                key={plan}
                                label={plan}
                                value={count as number}
                                total={data.total_tenants}
                                color={getPlanColor(plan)}
                            />
                        ))}
                    </div>
                </div>

                {/* Subscription Status */}
                <div className="glass-panel p-8">
                    <h3 className="text-xs font-black uppercase tracking-widest text-text-muted mb-8 flex items-center gap-2">
                        <BarChart3 size={14} className="text-success" />
                        Subscription Health
                    </h3>
                    <div className="space-y-4">
                        {Object.entries(data.subscription_distribution || {}).map(([status, count]) => (
                            <DistributionBar
                                key={status}
                                label={status}
                                value={count as number}
                                total={data.total_tenants}
                                color={status === 'active' ? 'bg-success' : status === 'trialing' ? 'bg-warning' : 'bg-error'}
                            />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}

interface AnalyticStatProps {
    label: string;
    value: string | number;
    sub: string;
    icon: React.ComponentType<{ size?: number; className?: string }>;
    color: string;
    formatValue?: (value: string | number) => string;
}

function AnalyticStat({ label, value, sub, icon: Icon, color, formatValue }: AnalyticStatProps) {
    const displayValue = formatValue ? formatValue(value) : value;
    return (
        <div className="glass-panel p-6 relative overflow-hidden group">
            <div className="flex items-center justify-between mb-4">
                <p className="text-[10px] font-black uppercase tracking-widest text-text-muted">{label}</p>
                <div className={`p-2 rounded-lg bg-surface border border-border/50 ${color}`}>
                    <Icon size={14} />
                </div>
            </div>
            <div className="text-2xl font-black text-text-main">{displayValue}</div>
            <p className="text-[10px] font-bold text-text-muted mt-1 uppercase tracking-tight">{sub}</p>
        </div>
    );
}

function DistributionBar({ label, value, total, color }: { label: string; value: number; total: number; color: string }) {
    const percent = total > 0 ? (value / total) * 100 : 0;
    return (
        <div className="space-y-2">
            <div className="flex items-center justify-between text-[10px] font-black uppercase tracking-widest">
                <span className="text-text-main">{label}</span>
                <span className="text-text-muted">{value} ({percent.toFixed(1)}%)</span>
            </div>
            <div className="h-2 w-full bg-surface-muted rounded-full overflow-hidden border border-border/20">
                <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${percent}%` }}
                    className={`h-full ${color}`}
                />
            </div>
        </div>
    );
}

function calculateRevenue(plans: Record<string, number> = {}) {
    const starter = (plans.starter || 0) * 12000;
    const pro = (plans.pro || 0) * 30000;
    const enterprise = (plans.enterprise || 0) * 85000;
    const total = (starter + pro + enterprise);
    return `₦${(total).toLocaleString()}`;
}

function getPlanColor(plan: string) {
    const colors: Record<string, string> = {
        trial: "bg-text-muted/30",
        starter: "bg-primary",
        pro: "bg-accent",
        enterprise: "bg-accent-deep",
    };
    return colors[plan.toLowerCase()] || "bg-text-muted/20";
}
