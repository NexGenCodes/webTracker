"use client";

import { useState } from "react";
import {
    AlertOctagon,
    Trash2,
    Calendar,
    Ban,
    RefreshCcw,
    PowerOff,
} from "lucide-react";
import { getApiUrl } from "@/lib/utils";
import { toast } from "react-hot-toast";

interface Company {
    id: string;
    name: string | null;
    admin_email: string;
    plan_type: string | null;
    subscription_status: string | null;
}

interface TenantActionsProps {
    company: Company;
    onActionComplete: () => void;
}

export default function TenantActions({ company, onActionComplete }: TenantActionsProps) {
    const [loading, setLoading] = useState<string | null>(null);
    const [confirmDelete, setConfirmDelete] = useState("");

    const handleAction = async (action: string, method: string, path: string, body?: Record<string, unknown>) => {
        setLoading(action);
        try {
            const res = await fetch(`${getApiUrl()}${path}`, {
                method,
                headers: { "Content-Type": "application/json" },
                body: body ? JSON.stringify(body) : undefined,
                credentials: "include",
            });

            if (!res.ok) {
                const data = await res.json();
                throw new Error(data.error || "Action failed");
            }

            toast.success(`Action "${action}" completed successfully`);
            onActionComplete();
        } catch (err) {
            const message = err instanceof Error ? err.message : "Action failed";
            toast.error(message);
        } finally {
            setLoading(null);
        }
    };

    const extendSubscription = () => {
        const nextMonth = new Date();
        nextMonth.setMonth(nextMonth.getMonth() + 1);
        handleAction("Extend Subscription", "PUT", `/api/super-admin/companies/${company.id}/subscription`, {
            status: "active",
            expiry: nextMonth.toISOString(),
        });
    };

    const toggleSuspension = () => {
        const isSuspended = company.subscription_status === "suspended";
        handleAction(isSuspended ? "Reactivate" : "Suspend", "PUT", `/api/super-admin/companies/${company.id}/subscription`, {
            status: isSuspended ? "active" : "suspended",
        });
    };

    const changePlan = (plan: string) => {
        handleAction(`Change Plan to ${plan}`, "PUT", `/api/super-admin/companies/${company.id}/plan`, {
            plan_type: plan,
        });
    };

    const disconnectBot = () => {
        handleAction("Disconnect Bot", "POST", `/api/super-admin/companies/${company.id}/disconnect-bot`);
    };

    const deleteTenant = () => {
        if (confirmDelete !== (company.name || company.admin_email)) {
            toast.error("Confirmation name does not match");
            return;
        }
        handleAction("Delete Tenant", "DELETE", `/api/super-admin/companies/${company.id}`);
    };

    return (
        <div className="mt-8 pt-8 border-t border-border/50">
            <h3 className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-4 flex items-center gap-2">
                <AlertOctagon size={12} className="text-error" />
                Administrative Actions
            </h3>

            <div className="grid grid-cols-2 gap-3">
                {/* Subscription Actions */}
                <button
                    disabled={!!loading}
                    onClick={extendSubscription}
                    className="flex items-center justify-center gap-2 py-2.5 rounded-xl border border-border/50 bg-surface hover:border-border transition-all active:scale-95 text-xs font-bold text-text-main disabled:opacity-50"
                >
                    <Calendar size={14} />
                    <span>+30 Days Sub</span>
                </button>

                <button
                    disabled={!!loading}
                    onClick={toggleSuspension}
                    className={`flex items-center justify-center gap-2 py-2.5 rounded-xl border border-border/50 bg-surface hover:border-border transition-all active:scale-95 text-xs font-bold text-text-main disabled:opacity-50 ${company.subscription_status === "suspended" ? "text-success bg-success/5 border-success/20" : "text-warning bg-warning/5 border-warning/20"}`}
                >
                    {company.subscription_status === "suspended" ? <RefreshCcw size={14} /> : <Ban size={14} />}
                    <span>{company.subscription_status === "suspended" ? "Reactivate" : "Suspend"}</span>
                </button>

                {/* Plan Changes */}
                <div className="col-span-2 flex items-center gap-2 mt-2">
                    {["trial", "starter", "pro", "enterprise"].map((p) => (
                        <button
                            key={p}
                            disabled={!!loading || company.plan_type === p}
                            onClick={() => changePlan(p)}
                            className={`flex-1 py-2 rounded-lg border text-[10px] font-black uppercase tracking-widest transition-all ${company.plan_type === p
                                    ? "bg-accent text-white border-accent shadow-lg shadow-accent/20"
                                    : "bg-surface border-border/50 text-text-muted hover:border-border hover:text-text-main"
                                }`}
                        >
                            {p}
                        </button>
                    ))}
                </div>

                {/* Bot Control */}
                <button
                    disabled={!!loading}
                    onClick={disconnectBot}
                    className="flex items-center justify-center gap-2 py-2.5 rounded-xl border border-border/50 bg-surface hover:border-border transition-all active:scale-95 text-xs font-bold text-text-main disabled:opacity-50 col-span-2 mt-2"
                >
                    <PowerOff size={14} />
                    <span>Force Disconnect WhatsApp</span>
                </button>

                {/* Delete Section */}
                <div className="col-span-2 mt-6 p-4 rounded-xl bg-error/5 border border-error/20">
                    <p className="text-[10px] font-black uppercase tracking-widest text-error mb-3">Destructive: Delete Tenant</p>
                    <div className="flex gap-2">
                        <input
                            type="text"
                            placeholder="Type company name to confirm"
                            value={confirmDelete}
                            onChange={(e) => setConfirmDelete(e.target.value)}
                            className="flex-1 bg-surface border border-error/20 rounded-lg px-3 py-2 text-xs focus:ring-1 focus:ring-error outline-none"
                        />
                        <button
                            disabled={!!loading || confirmDelete !== (company.name || company.admin_email)}
                            onClick={deleteTenant}
                            className="bg-error hover:bg-error-deep disabled:opacity-50 text-white px-4 py-2 rounded-lg transition-all active:scale-95 flex items-center gap-2 text-xs font-bold"
                        >
                            <Trash2 size={14} />
                            Delete
                        </button>
                    </div>
                </div>
            </div>


        </div>
    );
}
