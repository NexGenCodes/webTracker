"use client";

import { useEffect, useState, useCallback } from "react";
import { 
    History,
    ChevronLeft,
    ChevronRight,
    RefreshCcw,
    User
} from "lucide-react";
import { getApiUrl } from "@/lib/utils";
import { motion } from "framer-motion";

interface AuditLog {
    id: number;
    actor_email: string;
    action: string;
    target_company_id: string | null;
    details: Record<string, unknown> | null;
    created_at: string;
}

export default function AuditLogTab() {
    const [logs, setLogs] = useState<AuditLog[]>([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(0);
    const limit = 20;

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetch(`${getApiUrl()}/api/super-admin/audit-log?limit=${limit}&offset=${page * limit}`, {
                credentials: "include",
            });
            const json = await res.json();
            setLogs(json || []);
        } catch (err) {
            console.error("Failed to fetch audit logs", err);
        } finally {
            setLoading(false);
        }
    }, [page, limit]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-xs font-black uppercase tracking-[0.2em] text-text-muted flex items-center gap-2">
                    <History size={14} className="text-accent" />
                    Security & Action Logs
                </h2>
                <div className="flex items-center gap-2">
                    <button
                        disabled={page === 0 || loading}
                        onClick={() => setPage(p => p - 1)}
                        className="p-2 rounded-lg bg-surface border border-border/50 hover:border-border disabled:opacity-50 transition-all"
                    >
                        <ChevronLeft size={16} />
                    </button>
                    <span className="text-[10px] font-black uppercase tracking-widest text-text-muted px-2">Page {page + 1}</span>
                    <button
                        disabled={logs.length < limit || loading}
                        onClick={() => setPage(p => p + 1)}
                        className="p-2 rounded-lg bg-surface border border-border/50 hover:border-border disabled:opacity-50 transition-all"
                    >
                        <ChevronRight size={16} />
                    </button>
                    <button
                        onClick={fetchLogs}
                        className="ml-4 p-2 rounded-lg bg-surface border border-border/50 hover:border-border transition-all active:rotate-180 duration-500"
                    >
                        <RefreshCcw size={16} className={loading ? "animate-spin" : ""} />
                    </button>
                </div>
            </div>

            <div className="glass-panel overflow-hidden">
                {/* Desktop Table */}
                <div className="hidden md:block overflow-x-auto">
                    <table className="w-full text-left">
                        <thead>
                            <tr className="border-b border-border/50">
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Timestamp</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Actor</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Action</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Details</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border/20">
                            {loading && logs.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="text-center py-20">
                                        <RefreshCcw className="animate-spin text-accent mx-auto mb-4" size={24} />
                                        <p className="text-[10px] font-black uppercase tracking-widest text-text-muted">Syncing logs…</p>
                                    </td>
                                </tr>
                            ) : logs.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="text-center py-20 text-text-muted text-sm font-medium">
                                        No audit entries found.
                                    </td>
                                </tr>
                            ) : (
                                logs.map((log, i) => (
                                    <motion.tr 
                                        key={log.id}
                                        initial={{ opacity: 0, x: -10 }}
                                        animate={{ opacity: 1, x: 0 }}
                                        transition={{ delay: i * 0.02 }}
                                        className="hover:bg-surface-muted/30 transition-colors"
                                    >
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-[10px] font-bold text-text-muted">
                                                {new Date(log.created_at).toLocaleDateString()}
                                            </div>
                                            <div className="text-[10px] font-black text-text-main uppercase">
                                                {new Date(log.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="flex items-center gap-2">
                                                <div className="w-6 h-6 rounded-full bg-accent/10 border border-accent/20 flex items-center justify-center text-accent">
                                                    <User size={12} />
                                                </div>
                                                <span className="text-xs font-bold text-text-main">{log.actor_email}</span>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-tighter border ${getActionStyles(log.action)}`}>
                                                {log.action.replace(/_/g, ' ')}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="max-w-xs truncate text-[9px] font-mono text-text-muted bg-surface-muted/50 p-2 rounded-md border border-border/10 overflow-hidden">
                                                {JSON.stringify(log.details, null, 1)}
                                            </div>
                                        </td>
                                    </motion.tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Mobile Card View */}
                <div className="md:hidden divide-y divide-border/20">
                    {loading && logs.length === 0 ? (
                        <div className="text-center py-20">
                            <RefreshCcw className="animate-spin text-accent mx-auto mb-4" size={24} />
                            <p className="text-[10px] font-black uppercase tracking-widest text-text-muted">Syncing logs…</p>
                        </div>
                    ) : logs.length === 0 ? (
                        <div className="text-center py-20 text-text-muted text-sm font-medium">
                            No audit entries found.
                        </div>
                    ) : (
                        logs.map((log, i) => (
                            <motion.div 
                                key={log.id}
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: i * 0.02 }}
                                className="p-4 space-y-3"
                            >
                                <div className="flex items-center justify-between">
                                    <div className="text-[10px] font-black text-text-muted uppercase tracking-widest">
                                        {new Date(log.created_at).toLocaleDateString()} @ {new Date(log.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                    </div>
                                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[9px] font-black uppercase tracking-tighter border ${getActionStyles(log.action)}`}>
                                        {log.action.replace(/_/g, ' ')}
                                    </span>
                                </div>
                                
                                <div className="flex items-center gap-2">
                                    <div className="w-5 h-5 rounded-full bg-accent/10 border border-accent/20 flex items-center justify-center text-accent">
                                        <User size={10} />
                                    </div>
                                    <span className="text-[11px] font-bold text-text-main truncate">{log.actor_email}</span>
                                </div>

                                {log.details && (
                                    <div className="text-[9px] font-mono text-text-muted bg-surface-muted/50 p-2 rounded-md border border-border/10 overflow-hidden break-all whitespace-pre-wrap">
                                        {JSON.stringify(log.details, null, 2)}
                                    </div>
                                )}
                            </motion.div>
                        ))
                    )}
                </div>
            </div>
        </div>
    );
}

function getActionStyles(action: string) {
    if (action.includes('delete')) return "bg-error/5 text-error border-error/20";
    if (action.includes('update') || action.includes('change')) return "bg-accent/5 text-accent border-accent/20";
    if (action.includes('force')) return "bg-warning/5 text-warning border-warning/20";
    return "bg-text-muted/5 text-text-muted border-text-muted/20";
}
