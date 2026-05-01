import React from 'react';
import { Smartphone } from 'lucide-react';

interface WhatsAppTabProps {
    whatsappConnected: boolean;
    onConnectClick: () => void;
}

export function WhatsAppTab({ whatsappConnected, onConnectClick }: WhatsAppTabProps) {
    return (
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
                        onClick={onConnectClick}
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
    );
}
