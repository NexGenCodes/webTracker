import React from 'react';
import { Copy, Check, Loader2, RefreshCw } from 'lucide-react';
import { createClient } from '@/lib/supabase/client';

interface PhoneCodeViewProps {
    pairingCode: string;
    pairStatus: 'idle' | 'waiting' | 'connected';
    handleCopyCode: () => void;
    codeCopied: boolean;
    onRegenerate: () => void;
    companyId: string;
    handleConnected: () => void;
}

export function PhoneCodeView({
    pairingCode,
    pairStatus,
    handleCopyCode,
    codeCopied,
    onRegenerate,
    companyId,
    handleConnected
}: PhoneCodeViewProps) {
    return (
        <div className="text-center space-y-8">
            <div className="space-y-2">
                <h4 className="text-lg font-black text-text-main">Enter Code on WhatsApp</h4>
                <p className="text-sm font-medium text-text-muted max-w-sm mx-auto">
                    Open WhatsApp {'>'} Linked Devices {'>'} Link with phone number instead
                </p>
            </div>

            <div className="flex justify-center items-center gap-1 sm:gap-2 md:gap-3 w-full">
                {pairingCode.split('').map((char, i) => (
                    <React.Fragment key={i}>
                        {i === 4 && <div className="w-2 sm:w-4 flex items-center justify-center text-border font-black text-lg sm:text-2xl shrink-0">-</div>}
                        <div className="flex-1 max-w-[2.5rem] sm:max-w-[3rem] aspect-[4/5] bg-surface border border-border/50 rounded-lg sm:rounded-xl flex items-center justify-center text-lg sm:text-2xl md:text-3xl font-black text-accent shadow-inner">
                            {char}
                        </div>
                    </React.Fragment>
                ))}
            </div>
            <div className="flex flex-col sm:flex-row justify-center gap-3 mt-6">
                <button
                    onClick={handleCopyCode}
                    className="inline-flex items-center justify-center gap-2 px-6 py-3.5 sm:py-2.5 rounded-xl sm:rounded-full bg-surface border border-border hover:bg-surface-hover text-sm font-bold text-text-main transition-colors w-full sm:w-auto active:scale-95"
                >
                    {codeCopied ? <Check size={16} className="text-success" /> : <Copy size={16} />}
                    {codeCopied ? 'Copied!' : 'Copy Code'}
                </button>
                <button
                    onClick={onRegenerate}
                    className="inline-flex items-center justify-center gap-2 px-6 py-3.5 sm:py-2.5 rounded-xl sm:rounded-full bg-surface-muted hover:bg-border text-text-muted hover:text-text-main transition-colors w-full sm:w-auto text-sm font-bold active:scale-95"
                >
                    <RefreshCw size={16} />
                    Regenerate
                </button>
            </div>

            <div className="p-4 rounded-xl bg-accent/5 border border-accent/10 flex items-center justify-center gap-3">
                {pairStatus === 'connected' ? (
                    <>
                        <div className="w-8 h-8 rounded-full bg-success/20 flex items-center justify-center text-success">
                            <Check size={16} />
                        </div>
                        <p className="text-sm font-black text-success uppercase tracking-widest">Connected!</p>
                    </>
                ) : (
                    <div className="flex flex-col items-center gap-2">
                        <div className="flex items-center gap-3">
                            <Loader2 size={16} className="animate-spin text-accent" />
                            <p className="text-sm font-bold text-text-muted">Waiting for connection...</p>
                        </div>
                        <button 
                            onClick={() => {
                                const supabase = createClient();
                                supabase.from('companies').select('auth_status').eq('id', companyId).single().then(({ data }) => {
                                    if (data?.auth_status === 'active') handleConnected();
                                });
                            }}
                            className="text-[10px] font-black uppercase tracking-widest text-accent hover:underline flex items-center gap-1"
                        >
                            <RefreshCw size={10} /> Check Status Manually
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
