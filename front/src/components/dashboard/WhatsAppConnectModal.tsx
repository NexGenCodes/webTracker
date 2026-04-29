import React, { useState, useEffect } from 'react';
import { Smartphone, Copy, Check, Loader2, XCircle, ChevronRight, AlertTriangle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { z } from 'zod';
import { getApiUrl } from '@/lib/utils';

interface WhatsAppConnectModalProps {
    isOpen: boolean;
    onClose: () => void;
    companyId: string;
    companyData: {
        auth_status?: string | null;
        subscription_status?: string | null;
    } | null;
    onSuccess: () => void;
}

export default function WhatsAppConnectModal({ isOpen, onClose, companyId, companyData, onSuccess }: WhatsAppConnectModalProps) {
    const [countryCode, setCountryCode] = useState('+234');
    const [connectPhone, setConnectPhone] = useState('');
    const [pairingCode, setPairingCode] = useState('');
    const [isPairing, setIsPairing] = useState(false);
    const [pairError, setPairError] = useState('');
    const [codeCopied, setCodeCopied] = useState(false);
    const [pairStatus, setPairStatus] = useState<'idle' | 'waiting' | 'connected'>('idle');

    // Reset state when opened
    useEffect(() => {
        if (isOpen) {
            setPairingCode('');
            setPairError('');
            setPairStatus('idle');
            setConnectPhone('');
        }
    }, [isOpen]);

    // Handle pairing modal success state when auth_status changes to active
    useEffect(() => {
        if (isOpen && companyData?.auth_status === 'active' && pairStatus === 'waiting') {
            setPairStatus('connected');
            setTimeout(() => {
                onSuccess();
                onClose();
            }, 2500);
        }
    }, [companyData?.auth_status, isOpen, pairStatus, onClose, onSuccess]);

    const handleCopyCode = async () => {
        if (!pairingCode) return;
        await navigator.clipboard.writeText(pairingCode);
        setCodeCopied(true);
        setTimeout(() => setCodeCopied(false), 2000);
    };

    const handlePair = async () => {
        setIsPairing(true);
        setPairError('');

        // Subscription validation check (Client side pre-check)
        if (companyData?.subscription_status !== 'active' && companyData?.subscription_status !== 'trialing') {
            setPairError('Subscription is inactive. Please renew to use the tracking bot.');
            setIsPairing(false);
            return;
        }

        const codeDigits = countryCode.replace('+', '');
        let processedPhone = connectPhone.replace(/[\s\-()]/g, ''); // strip formatting chars

        if (processedPhone.startsWith('+')) {
            processedPhone = processedPhone.substring(1);
        }

        if (processedPhone.startsWith(codeDigits)) {
            processedPhone = processedPhone.substring(codeDigits.length);
        }

        processedPhone = processedPhone.replace(/\D/g, '');

        if (processedPhone.startsWith('0') && processedPhone.length >= 10) {
            processedPhone = processedPhone.substring(1);
        }

        const phoneSchema = z.string()
            .min(7, "Number is too short")
            .max(12, "Number is too long")
            .regex(/^\d+$/, "Please enter only digits");

        try {
            phoneSchema.parse(processedPhone);
        } catch (e: unknown) {
            if (e instanceof z.ZodError) {
                setPairError(e.issues?.[0]?.message || "Invalid phone number format");
            } else {
                setPairError("Invalid phone number format");
            }
            setIsPairing(false);
            return;
        }

        const fullPhone = `${codeDigits}${processedPhone}`;

        try {
            const res = await fetch('/api/setup/pair', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ company_id: companyId, phone: fullPhone })
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Failed to generate code');

            // Safely extract pairing code from varied backend responses
            const code = data.pairing_code || data.code || data.data?.pairing_code || data.data?.code || data.pairingCode || data.data?.pairingCode;
            
            if (code) {
                setPairingCode(code);
                setPairStatus('waiting');

                // Save phone to company setup on backend (non-blocking, but handled)
                fetch(`${getApiUrl()}/api/auth/setup`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    credentials: 'include',
                    body: JSON.stringify({ whatsapp_phone: fullPhone })
                }).catch(err => console.error('Failed to save phone number:', err));
            } else {
                setPairError('Code not received. Check phone format.');
            }
        } catch (err: unknown) {
            setPairError(err instanceof Error ? err.message : 'An error occurred');
        } finally {
            setIsPairing(false);
        }
    };

    return (
        <AnimatePresence>
            {isOpen && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="fixed inset-0 z-50 flex items-center justify-center p-4 backdrop-blur-sm bg-background/80"
                >
                <motion.div
                    initial={{ scale: 0.95, opacity: 0, y: 20 }}
                    animate={{ scale: 1, opacity: 1, y: 0 }}
                    exit={{ scale: 0.95, opacity: 0, y: 20 }}
                    className="glass-panel w-full max-w-lg border border-border/50 shadow-2xl overflow-hidden"
                >
                    <div className="p-6 md:p-8 relative">
                        <button
                            onClick={onClose}
                            className="absolute top-6 right-6 text-text-muted hover:text-text-main transition-colors"
                        >
                            <XCircle size={24} />
                        </button>

                        <div className="flex items-center gap-4 mb-8">
                            <div className="w-12 h-12 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent">
                                <Smartphone size={24} />
                            </div>
                            <div>
                                <h3 className="text-xl font-black uppercase tracking-tight text-text-main">Link Device</h3>
                                <p className="text-xs font-bold uppercase tracking-widest text-text-muted mt-1">WhatsApp Business</p>
                            </div>
                        </div>

                        {pairError && (
                            <div className="mb-6 p-4 rounded-xl bg-error/10 border border-error/20 flex items-start gap-3">
                                <AlertTriangle className="text-error shrink-0 mt-0.5" size={16} />
                                <p className="text-sm font-medium text-error leading-relaxed">{pairError}</p>
                            </div>
                        )}

                        {!pairingCode ? (
                            <div className="space-y-6">
                                <div className="space-y-2">
                                    <label className="text-xs font-black uppercase tracking-widest text-text-muted">Business Number</label>
                                    <div className="flex relative items-center">
                                        <select
                                            value={countryCode}
                                            onChange={(e) => setCountryCode(e.target.value)}
                                            className="absolute left-0 top-0 bottom-0 z-10 appearance-none bg-surface border border-border rounded-l-xl px-4 py-3 text-sm font-black text-text-main focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all cursor-pointer"
                                            style={{ width: '100px' }}
                                        >
                                            <option value="+234">NG (+234)</option>
                                            <option value="+27">SA (+27)</option>
                                            <option value="+254">KE (+254)</option>
                                            <option value="+233">GH (+233)</option>
                                            <option value="+1">US (+1)</option>
                                            <option value="+44">UK (+44)</option>
                                        </select>
                                        <input
                                            type="tel"
                                            value={connectPhone}
                                            onChange={(e) => setConnectPhone(e.target.value)}
                                            placeholder="803 000 0000"
                                            className="input-premium pl-[110px]"
                                            onKeyDown={(e) => { if (e.key === 'Enter') handlePair(); }}
                                        />
                                    </div>
                                    <p className="text-[10px] font-medium text-text-muted pl-1">Exclude the leading zero if applicable (e.g. 803 instead of 0803)</p>
                                </div>

                                <button
                                    onClick={handlePair}
                                    disabled={isPairing || !connectPhone}
                                    className="btn-primary w-full py-4 text-sm flex items-center justify-center gap-2 group"
                                >
                                    {isPairing ? (
                                        <>
                                            <Loader2 size={18} className="animate-spin" />
                                            <span>Processing...</span>
                                        </>
                                    ) : (
                                        <>
                                            <span>Generate Code</span>
                                            <ChevronRight size={18} className="group-hover:translate-x-1 transition-transform" />
                                        </>
                                    )}
                                </button>
                            </div>
                        ) : (
                            <div className="text-center space-y-8">
                                <div className="space-y-2">
                                    <h4 className="text-lg font-black text-text-main">Enter Code on WhatsApp</h4>
                                    <p className="text-sm font-medium text-text-muted max-w-sm mx-auto">
                                        Open WhatsApp {'>'} Linked Devices {'>'} Link with phone number instead
                                    </p>
                                </div>

                                <div className="flex justify-center gap-3">
                                    {pairingCode.split('').map((char, i) => (
                                        <React.Fragment key={i}>
                                            {i === 4 && <div className="w-4 flex items-center justify-center text-border font-black text-2xl">-</div>}
                                            <div className="w-10 h-14 md:w-12 md:h-16 bg-surface border border-border/50 rounded-xl flex items-center justify-center text-2xl md:text-3xl font-black text-accent shadow-inner">
                                                {char}
                                            </div>
                                        </React.Fragment>
                                    ))}
                                </div>
                                <button
                                    onClick={handleCopyCode}
                                    className="inline-flex items-center gap-2 px-6 py-2.5 rounded-full bg-surface border border-border hover:bg-surface-hover text-sm font-bold text-text-main transition-colors mx-auto"
                                >
                                    {codeCopied ? <Check size={16} className="text-success" /> : <Copy size={16} />}
                                    {codeCopied ? 'Copied!' : 'Copy Code'}
                                </button>

                                <div className="p-4 rounded-xl bg-accent/5 border border-accent/10 flex items-center justify-center gap-3">
                                    {pairStatus === 'connected' ? (
                                        <>
                                            <div className="w-8 h-8 rounded-full bg-success/20 flex items-center justify-center text-success">
                                                <Check size={16} />
                                            </div>
                                            <p className="text-sm font-black text-success uppercase tracking-widest">Connected!</p>
                                        </>
                                    ) : (
                                        <>
                                            <Loader2 size={16} className="animate-spin text-accent" />
                                            <p className="text-sm font-bold text-text-muted">Waiting for connection...</p>
                                        </>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}
