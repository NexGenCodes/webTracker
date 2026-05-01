"use client";

import { useState, useEffect } from 'react';
import { Settings, Loader2, CheckCircle2, UserCircle2, AlertTriangle, Trash2, Unplug } from 'lucide-react';
import { motion } from 'framer-motion';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { createClient } from '@/lib/supabase/client';
import { disconnectWhatsApp, deleteAccount } from '@/app/actions/setup';
import { logoutAction } from '@/app/actions/auth';
import toast from 'react-hot-toast';

export default function SettingsPage() {
    const { companyId, loading, user } = useMultiTenant();

    const [settingsForm, setSettingsForm] = useState({ name: '', admin_email: '', logo_url: '' });
    const [isSaving, setIsSaving] = useState(false);
    const [saveSuccess, setSaveSuccess] = useState(false);
    const [isDisconnecting, setIsDisconnecting] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deleteConfirmText, setDeleteConfirmText] = useState('');
    const [companyLoading, setCompanyLoading] = useState(true);

    const supabase = createClient();

    useEffect(() => {
        if (!companyId) return;

        const fetchCompanyData = async () => {
            try {
                const { data, error } = await supabase
                    .from('companies')
                    .select('name, admin_email, logo_url')
                    .eq('id', companyId)
                    .single();

                if (data && !error) {
                    setSettingsForm({
                        name: data.name || '',
                        admin_email: data.admin_email || '',
                        logo_url: data.logo_url || ''
                    });
                }
            } catch (err) {
                console.error('Error fetching settings:', err);
            } finally {
                setCompanyLoading(false);
            }
        };

        fetchCompanyData();
    }, [companyId, supabase]);

    const handleSettingsSave = async () => {
        setIsSaving(true);
        try {
            const { error } = await supabase
                .from('companies')
                .update(settingsForm)
                .eq('id', companyId);

            if (error) throw error;
            setSaveSuccess(true);
            setTimeout(() => setSaveSuccess(false), 3000);
        } catch (error) {
            console.error('Error saving settings:', error);
            alert("Failed to save settings");
        } finally {
            setIsSaving(false);
        }
    };

    const handleDisconnectBot = async () => {
        if (!confirm("Are you sure you want to disconnect the WhatsApp bot? This will instantly stop all tracking and delete your session.")) return;

        setIsDisconnecting(true);
        try {
            if (!companyId) return;
            await disconnectWhatsApp(companyId);
            toast.success("Bot disconnected successfully.");
            window.location.reload(); // Refresh to update status globally
        } catch {
            toast.error("Failed to disconnect bot");
        } finally {
            setIsDisconnecting(false);
        }
    };

    const handleDeleteAccount = async () => {
        setIsDeleting(true);
        try {
            if (!companyId) return;
            await deleteAccount(companyId);
            toast.success("Account deleted. Redirecting...");
            
            // Clear session and force full reload to auth
            await logoutAction();
            window.location.href = '/auth';
        } catch {
            toast.error("Failed to delete account. Please try again.");
            setIsDeleting(false);
        }
    };

    if (loading || companyLoading) {
        return (
            <div className="flex-1 flex flex-col items-center justify-center p-12 relative z-10 pt-32">
                <Loader2 className="w-8 h-8 animate-spin text-accent mb-4" />
                <p className="text-text-muted text-sm font-black uppercase tracking-widest animate-pulse">Loading settings...</p>
            </div>
        );
    }

    return (
        <div className="pb-32 md:pb-24 relative bg-background overflow-x-hidden min-h-screen">
            <div className="max-w-4xl mx-auto z-10 relative pt-24 md:pt-32 px-4 sm:px-8">

                {/* Header Section */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="mb-12"
                >
                    <h1 className="text-3xl md:text-5xl font-black uppercase tracking-tighter mb-4 text-text-main">
                        Workspace <span className="text-accent">Settings</span>
                    </h1>
                    <p className="text-text-muted text-sm md:text-base">
                        Manage your company profile, branding, and dangerous operations.
                    </p>
                </motion.div>

                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 }}
                    className="max-w-2xl space-y-8"
                >
                    {/* User Profile */}
                    <div className="glass-panel p-8 md:p-10 border-border/50">
                        <div className="flex items-center gap-4 mb-8 pb-8 border-b border-border/50">
                            <div className="w-12 h-12 rounded-2xl bg-surface border border-border flex items-center justify-center text-text-main">
                                <UserCircle2 size={20} />
                            </div>
                            <div>
                                <h2 className="text-xl font-black text-text-main uppercase tracking-tighter">
                                    User Account
                                </h2>
                                <p className="text-xs font-bold text-text-muted mt-1 uppercase tracking-widest">Personal details & security</p>
                            </div>
                        </div>

                        <div className="space-y-6">
                            <div className="space-y-3">
                                <label className="text-[10px] font-black uppercase tracking-[0.2em] text-text-main ml-1">
                                    Email Address
                                </label>
                                <input
                                    type="email"
                                    value={user?.email || ''}
                                    disabled
                                    className="input-premium w-full !bg-surface-muted opacity-70 cursor-not-allowed"
                                />
                                <p className="text-[10px] text-text-muted ml-1">Your email is managed by your authentication provider.</p>
                            </div>

                            <div className="pt-4 border-t border-border/50 flex justify-between items-center">
                                <div className="text-sm font-medium text-text-muted">Need to change your password?</div>
                                <button
                                    onClick={() => alert('Password reset link has been sent to your email.')}
                                    className="px-5 py-2.5 bg-surface hover:bg-surface-muted text-text-main rounded-xl border border-border text-xs font-black uppercase tracking-widest transition-all active:scale-95"
                                >
                                    Reset Password
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Company Profile Form */}
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

                            <div className="pt-4 border-t border-border/50 flex items-center justify-end gap-4">
                                {saveSuccess && (
                                    <span className="flex items-center gap-2 text-xs font-bold text-success uppercase tracking-widest animate-fade-in">
                                        <CheckCircle2 size={16} /> Saved Successfully
                                    </span>
                                )}
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
                    <div className="glass-panel p-8 md:p-10 border-error/20 bg-error/5 relative overflow-hidden space-y-8">
                        <div className="absolute top-0 left-0 w-1 h-full bg-error" />

                        <h3 className="text-sm font-black text-error uppercase tracking-[0.2em] flex items-center gap-2">
                            <AlertTriangle size={16} /> Danger Zone
                        </h3>

                        {/* Disconnect WhatsApp */}
                        <div className="flex items-start justify-between flex-col sm:flex-row gap-4 pb-8 border-b border-error/10">
                            <div>
                                <h4 className="text-xs font-black text-text-main uppercase tracking-widest mb-1 flex items-center gap-2">
                                    <Unplug size={14} /> Disconnect WhatsApp
                                </h4>
                                <p className="text-sm font-medium text-text-muted leading-relaxed max-w-sm">
                                    Stops all tracking instantly and deletes the WhatsApp session.
                                </p>
                            </div>
                            <button
                                id="settings-disconnect-bot"
                                onClick={handleDisconnectBot}
                                disabled={isDisconnecting}
                                className="px-6 py-3 bg-error/10 hover:bg-error text-error hover:text-white border border-error/20 rounded-xl font-black text-xs uppercase tracking-widest whitespace-nowrap shadow-sm transition-all duration-200 active:scale-95 flex items-center gap-2 shrink-0"
                            >
                                {isDisconnecting ? <Loader2 size={16} className="animate-spin" /> : "Disconnect Bot"}
                            </button>
                        </div>

                        {/* Delete Account */}
                        <div className="space-y-4">
                            <div className="flex items-start justify-between flex-col sm:flex-row gap-4">
                                <div>
                                    <h4 className="text-xs font-black text-error uppercase tracking-widest mb-1 flex items-center gap-2">
                                        <Trash2 size={14} /> Delete Account
                                    </h4>
                                    <p className="text-sm font-medium text-text-muted leading-relaxed max-w-sm">
                                        Permanently delete your company, all shipments, and data. This action is <strong className="text-error">irreversible</strong>.
                                    </p>
                                </div>
                                {!showDeleteConfirm && (
                                    <button
                                        id="settings-delete-account-trigger"
                                        onClick={() => setShowDeleteConfirm(true)}
                                        className="px-6 py-3 bg-error/10 hover:bg-error text-error hover:text-white border border-error/20 rounded-xl font-black text-xs uppercase tracking-widest whitespace-nowrap shadow-sm transition-all duration-200 active:scale-95 flex items-center gap-2 shrink-0"
                                    >
                                        <Trash2 size={14} /> Delete Account
                                    </button>
                                )}
                            </div>

                            {showDeleteConfirm && (
                                <div className="p-6 bg-error/10 border border-error/20 rounded-2xl space-y-4">
                                    <p className="text-sm font-bold text-text-main">
                                        Type <span className="text-error font-black">{settingsForm.name || 'your company name'}</span> to confirm:
                                    </p>
                                    <input
                                        id="delete-account-confirm-input"
                                        type="text"
                                        value={deleteConfirmText}
                                        onChange={(e) => setDeleteConfirmText(e.target.value)}
                                        placeholder="Type company name here..."
                                        className="input-premium w-full !bg-surface/50 focus:!bg-surface !border-error/30 focus:!border-error"
                                    />
                                    <div className="flex items-center gap-3">
                                        <button
                                            id="delete-account-confirm"
                                            onClick={handleDeleteAccount}
                                            disabled={isDeleting || deleteConfirmText !== settingsForm.name}
                                            className="px-6 py-3 bg-error hover:bg-error/90 text-white rounded-xl font-black text-xs uppercase tracking-widest shadow-lg shadow-error/20 transition-all duration-200 active:scale-95 flex items-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed"
                                        >
                                            {isDeleting ? (
                                                <><Loader2 size={14} className="animate-spin" /> Deleting…</>
                                            ) : (
                                                <><Trash2 size={14} /> Permanently Delete</>
                                            )}
                                        </button>
                                        <button
                                            id="delete-account-cancel"
                                            onClick={() => { setShowDeleteConfirm(false); setDeleteConfirmText(''); }}
                                            disabled={isDeleting}
                                            className="px-6 py-3 bg-surface hover:bg-surface-muted text-text-main border border-border rounded-xl font-black text-xs uppercase tracking-widest transition-all duration-200 active:scale-95"
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </motion.div>
            </div>
        </div>
    );
}
