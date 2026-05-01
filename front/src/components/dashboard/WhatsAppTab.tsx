'use client';

import React, { useState, useTransition } from 'react';
import { Smartphone, Unplug, Loader2, AlertTriangle, X } from 'lucide-react';
import { disconnectWhatsApp } from '@/app/actions/setup';
import { useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';

interface WhatsAppTabProps {
    whatsappConnected: boolean;
    whatsappPhone?: string;
    companyId: string;
    onConnectClick: () => void;
}

export function WhatsAppTab({ whatsappConnected, whatsappPhone, companyId, onConnectClick }: WhatsAppTabProps) {
    const queryClient = useQueryClient();
    const [showConfirm, setShowConfirm] = useState(false);
    const [isPending, startTransition] = useTransition();

    const handleDisconnect = () => {
        startTransition(async () => {
            try {
                await disconnectWhatsApp(companyId);
                toast.success('WhatsApp bot disconnected successfully.');
                queryClient.invalidateQueries({ queryKey: ['company', companyId] });
                setShowConfirm(false);
            } catch {
                toast.error('Failed to disconnect bot. Please try again.');
            }
        });
    };

    return (
        <div className="max-w-2xl mx-auto space-y-6">
            {/* Main Status Card */}
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
                        onClick={onConnectClick}
                        className="btn-primary px-10 py-4 text-sm flex items-center justify-center gap-3 mx-auto shadow-lg shadow-accent/20 hover:shadow-accent/40 hover:-translate-y-1 transition-all duration-300"
                    >
                        Connect Number
                    </button>
                ) : (
                    <div className="space-y-6">
                        <div className="inline-flex items-center gap-3 px-6 py-3 bg-success/10 border border-success/20 rounded-full text-success shadow-inner">
                            <span className="relative flex h-3 w-3">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-3 w-3 bg-success"></span>
                            </span>
                            <span className="text-sm font-black uppercase tracking-widest">Online &amp; Monitoring</span>
                        </div>

                        {whatsappPhone && (
                            <p className="text-xs font-bold text-text-muted uppercase tracking-widest">
                                Connected to: <span className="text-text-main">{whatsappPhone}</span>
                            </p>
                        )}
                    </div>
                )}
            </div>

            {/* Disconnect Section — only visible when connected */}
            {whatsappConnected && (
                <div className="glass-panel p-6 md:p-8 border-error/20 bg-error/5 relative overflow-hidden">
                    <div className="absolute top-0 left-0 w-1 h-full bg-error" />

                    {!showConfirm ? (
                        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                            <div className="pl-2">
                                <h3 className="text-xs font-black text-error uppercase tracking-[0.2em] mb-1 flex items-center gap-2">
                                    <Unplug size={14} /> Disconnect Bot
                                </h3>
                                <p className="text-sm font-medium text-text-muted leading-relaxed max-w-sm">
                                    Remove the WhatsApp connection. All automated tracking will stop immediately.
                                </p>
                            </div>
                            <button
                                id="disconnect-whatsapp-trigger"
                                onClick={() => setShowConfirm(true)}
                                className="px-6 py-3 bg-error/10 hover:bg-error text-error hover:text-white border border-error/20 rounded-xl font-black text-xs uppercase tracking-widest whitespace-nowrap shadow-sm transition-all duration-200 active:scale-95 flex items-center gap-2 shrink-0"
                            >
                                <Unplug size={14} />
                                Disconnect
                            </button>
                        </div>
                    ) : (
                        <div className="space-y-4 pl-2">
                            <div className="flex items-start gap-3">
                                <div className="w-10 h-10 rounded-xl bg-error/20 border border-error/30 flex items-center justify-center shrink-0 mt-0.5">
                                    <AlertTriangle size={18} className="text-error" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-black text-error uppercase tracking-widest mb-1">
                                        Are you sure?
                                    </h3>
                                    <p className="text-sm font-medium text-text-muted leading-relaxed">
                                        This will <strong className="text-text-main">immediately stop</strong> all automated tracking
                                        and delete the WhatsApp session. You can reconnect later with a new pairing code.
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center gap-3 pt-2">
                                <button
                                    id="disconnect-whatsapp-confirm"
                                    onClick={handleDisconnect}
                                    disabled={isPending}
                                    className="px-6 py-3 bg-error hover:bg-error/90 text-white rounded-xl font-black text-xs uppercase tracking-widest shadow-lg shadow-error/20 transition-all duration-200 active:scale-95 flex items-center gap-2 min-w-[160px] justify-center"
                                >
                                    {isPending ? (
                                        <>
                                            <Loader2 size={14} className="animate-spin" />
                                            Disconnecting…
                                        </>
                                    ) : (
                                        <>
                                            <Unplug size={14} />
                                            Yes, Disconnect
                                        </>
                                    )}
                                </button>
                                <button
                                    id="disconnect-whatsapp-cancel"
                                    onClick={() => setShowConfirm(false)}
                                    disabled={isPending}
                                    className="px-6 py-3 bg-surface hover:bg-surface-muted text-text-main border border-border rounded-xl font-black text-xs uppercase tracking-widest transition-all duration-200 active:scale-95 flex items-center gap-2"
                                >
                                    <X size={14} />
                                    Cancel
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
