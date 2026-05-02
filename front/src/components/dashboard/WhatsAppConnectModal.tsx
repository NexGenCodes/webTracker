import React, { useState, useEffect, useTransition, useCallback } from 'react';
import { Smartphone, Copy, Check, Loader2, XCircle, ChevronRight, AlertTriangle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { pairWhatsApp, getWhatsAppQR } from '@/app/actions/setup';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import { QRCodeSVG } from 'qrcode.react';
import { createClient } from '@/lib/supabase/client';

const formatLocalPhone = (countryCode: string, phone: string) => {
    let localPhone = phone.replace(/[\s\-()]/g, '');
    if (localPhone.startsWith('0')) {
        localPhone = localPhone.substring(1);
    }
    return localPhone.startsWith('+') ? localPhone : `${countryCode}${localPhone}`;
};

// --- CONFIGURATION ---
const COUNTRY_CODES = [
    { value: '+234', label: '🇳🇬 +234' },
    { value: '+27', label: '🇿🇦 +27' },
    { value: '+254', label: '🇰🇪 +254' },
    { value: '+233', label: '🇬🇭 +233' },
    { value: '+1', label: '🇺🇸 +1' },
    { value: '+44', label: '🇬🇧 +44' },
];

// --- ZOD SCHEMA (Defined Outside Component) ---
const phoneSchema = z.object({
    countryCode: z.string(),
    phone: z.string().min(1, "Phone number is required")
}).superRefine((data, ctx) => {
    const fullNumber = formatLocalPhone(data.countryCode, data.phone);
    const phoneNumber = parsePhoneNumberFromString(fullNumber);

    if (!phoneNumber || !phoneNumber.isValid()) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Please enter a valid phone number",
            path: ["phone"]
        });
    }
});

type PhoneFormValues = z.infer<typeof phoneSchema>;

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
    const [pairingCode, setPairingCode] = useState('');
    const [qrCodeData, setQrCodeData] = useState('');
    const [pairError, setPairError] = useState('');
    const [codeCopied, setCodeCopied] = useState(false);
    const [pairStatus, setPairStatus] = useState<'idle' | 'waiting' | 'connected'>('idle');
    const [connectMode, setConnectMode] = useState<'qr' | 'phone'>('qr');
    
    // React 19 / Next.js 15: useTransition for Server Action
    const [isPending, startTransition] = useTransition();

    // React Hook Form
    const { register, handleSubmit, reset, watch, formState: { errors } } = useForm<PhoneFormValues>({
        resolver: zodResolver(phoneSchema),
        defaultValues: {
            countryCode: '+234',
            phone: '',
        }
    });

    const watchPhone = watch('phone');

    // Handle connected logic
    const handleConnected = useCallback(() => {
        setPairStatus(prev => {
            if (prev === 'connected') return prev;
            // Execute side-effect
            setTimeout(() => {
                onSuccess();
                onClose();
            }, 2500);
            return 'connected';
        });
        setPairError('');
    }, [onSuccess, onClose]);

    const handleFetchQR = useCallback(() => {
        if (!companyId) return;
        startTransition(async () => {
            try {
                const response = await getWhatsAppQR(companyId);
                if (response.success && response.pairingCode) {
                    setQrCodeData(response.pairingCode);
                    setPairStatus(prev => prev === 'idle' ? 'waiting' : prev);
                } else {
                    setPairError(response.error || 'Could not fetch QR code.');
                }
            } catch (err: unknown) {
                const msg = err instanceof Error ? err.message : 'An error occurred.';
                if (msg.toLowerCase().includes('already connected')) {
                    handleConnected();
                } else {
                    setPairError(msg);
                }
            }
        });
    }, [companyId, handleConnected]);

    // Reset state when opened
    useEffect(() => {
        if (isOpen) {
            setPairingCode('');
            setQrCodeData('');
            setPairError('');
            setPairStatus('idle');
            setConnectMode('qr');
            reset();
            handleFetchQR();
        }
    }, [isOpen, reset, handleFetchQR]);

    // Auto-refresh QR code every 20 seconds while in QR mode and waiting
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (isOpen && connectMode === 'qr' && pairStatus === 'waiting') {
            interval = setInterval(() => {
                handleFetchQR();
            }, 20000); // 20 seconds
        }
        return () => {
            if (interval) clearInterval(interval);
        };
    }, [isOpen, connectMode, pairStatus, handleFetchQR]);

    // Handle pairing modal success state when auth_status changes to active (via Realtime)
    useEffect(() => {
        if (!isOpen || !companyId) return;

        // Fallback: check if already active via props (e.g. from React Query)
        if (companyData?.auth_status === 'active' && pairStatus === 'waiting') {
            handleConnected();
        }

        // Initialize Supabase Realtime
        const supabase = createClient();
        const channel = supabase
            .channel(`company_status_${companyId}`)
            .on(
                'postgres_changes',
                {
                    event: 'UPDATE',
                    schema: 'public',
                    table: 'companies',
                    filter: `id=eq.${companyId}`,
                },
                (payload) => {
                    console.log('[Realtime] Received company update:', payload.new);
                    if (payload.new.auth_status === 'active' && pairStatus === 'waiting') {
                        handleConnected();
                    }
                }
            )
            .subscribe((status) => {
                if (status === 'SUBSCRIBED') {
                    console.log('[Realtime] Subscribed to company status changes successfully.');
                } else if (status === 'CHANNEL_ERROR') {
                    console.error('[Realtime] Subscription failed. Check RLS or JWT config.');
                }
            });

        return () => {
            supabase.removeChannel(channel);
        };
    }, [companyId, companyData?.auth_status, isOpen, pairStatus, handleConnected]);

    useEffect(() => {
        let copyTimer: NodeJS.Timeout;
        if (codeCopied) {
            copyTimer = setTimeout(() => setCodeCopied(false), 2000);
        }
        return () => { if (copyTimer) clearTimeout(copyTimer); };
    }, [codeCopied]);

    const handleCopyCode = async () => {
        if (!pairingCode) return;
        const formatted = pairingCode.length === 8 
            ? `${pairingCode.slice(0, 4)}-${pairingCode.slice(4)}`
            : pairingCode;
        await navigator.clipboard.writeText(formatted);
        setCodeCopied(true);
    };

    const onSubmit = (data: PhoneFormValues) => {
        setPairError('');

        // Guard: companyId must be present
        if (!companyId) {
            console.error('[WhatsApp Pair] companyId is missing — session may have expired.');
            setPairError('Session expired. Please refresh the page and try again.');
            return;
        }

        // Subscription validation check
        const subStatus = companyData?.subscription_status ?? 'active';
        if (subStatus !== 'active' && subStatus !== 'trialing') {
            setPairError('Subscription is inactive. Please renew to use the tracking bot.');
            return;
        }

        // Process Phone Number securely using libphonenumber-js
        const fullNumber = formatLocalPhone(data.countryCode, data.phone);
        const phoneNumber = parsePhoneNumberFromString(fullNumber);

        if (!phoneNumber) {
            setPairError('Invalid phone number structure.');
            return;
        }

        // Extract strictly digits for the backend (E.164 without the '+')
        const formattedPhone = phoneNumber.number.replace('+', '');

        // Execute Server Action within transition
        startTransition(async () => {
            try {
                const response = await pairWhatsApp(companyId, formattedPhone);

                if (response.success && response.pairingCode) {
                    setPairingCode(response.pairingCode);
                    setPairStatus('waiting');
                } else {
                    setPairError(response.error || 'Code not received. Check phone format.');
                }
            } catch (err: unknown) {
                const msg = err instanceof Error ? err.message : 'An error occurred during pairing.';
                if (msg.toLowerCase().includes('already connected')) {
                    handleConnected();
                } else {
                    setPairError(msg);
                }
            }
        });
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
                        className="w-full max-w-lg border border-border/50 shadow-2xl overflow-hidden rounded-3xl bg-surface"
                        style={{ boxShadow: 'var(--glass-shadow)' }}
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

                            {/* Mode Toggle */}
                            {!pairingCode && pairStatus !== 'connected' && (
                                <div className="flex bg-surface-muted p-1 rounded-xl mb-6">
                                    <button
                                        type="button"
                                        onClick={() => { setConnectMode('qr'); if (!qrCodeData) handleFetchQR(); }}
                                        className={`flex-1 py-2 text-sm font-bold rounded-lg transition-all ${connectMode === 'qr' ? 'bg-surface shadow-sm text-text-main' : 'text-text-muted hover:text-text-main'}`}
                                    >
                                        QR Code
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setConnectMode('phone')}
                                        className={`flex-1 py-2 text-sm font-bold rounded-lg transition-all ${connectMode === 'phone' ? 'bg-surface shadow-sm text-text-main' : 'text-text-muted hover:text-text-main'}`}
                                    >
                                        Phone Number
                                    </button>
                                </div>
                            )}

                            {connectMode === 'qr' && !pairingCode ? (
                                <div className="text-center space-y-6">
                                    <div className="space-y-2">
                                        <h4 className="text-lg font-black text-text-main">Scan QR Code</h4>
                                        <p className="text-sm font-medium text-text-muted max-w-sm mx-auto">
                                            Open WhatsApp {'>'} Linked Devices {'>'} Link a device
                                        </p>
                                    </div>

                                    <div className="flex justify-center p-4 bg-white rounded-2xl mx-auto w-fit shadow-sm border border-border/50">
                                        {qrCodeData ? (
                                            <QRCodeSVG value={qrCodeData} size={200} />
                                        ) : (
                                            <div className="w-[200px] h-[200px] flex items-center justify-center bg-surface-muted rounded-xl">
                                                <Loader2 className="animate-spin text-accent" size={32} />
                                            </div>
                                        )}
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
                                            <>
                                                <Loader2 size={16} className="animate-spin text-accent" />
                                                <p className="text-sm font-bold text-text-muted">Waiting for scan...</p>
                                            </>
                                        )}
                                    </div>
                                </div>
                            ) : !pairingCode ? (
                                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                                    <div className="space-y-2">
                                        <label className="text-xs font-black uppercase tracking-widest text-text-muted">Business Number</label>
                                        <div className="flex items-stretch gap-0 rounded-2xl border-2 border-transparent focus-within:border-accent/20 focus-within:ring-4 focus-within:ring-accent-soft overflow-hidden transition-all bg-surface-muted">
                                            <select
                                                {...register('countryCode')}
                                                className="shrink-0 appearance-none border-r border-border px-4 py-4 text-sm font-black cursor-pointer focus:outline-none bg-surface text-text-main w-[110px]"
                                            >
                                                {COUNTRY_CODES.map(c => (
                                                    <option key={c.value} value={c.value}>{c.label}</option>
                                                ))}
                                            </select>
                                            <input
                                                type="tel"
                                                {...register('phone')}
                                                placeholder="803 000 0000"
                                                className="flex-1 min-w-0 px-4 py-4 text-base font-medium outline-none bg-transparent text-text-main caret-text-main"
                                            />
                                        </div>
                                        {errors.phone && (
                                            <p className="text-xs font-bold text-error mt-1">{errors.phone.message}</p>
                                        )}
                                        <p className="text-[10px] font-medium text-text-muted pl-1">Exclude the leading zero if applicable (e.g. 803 instead of 0803)</p>
                                    </div>

                                    <button
                                        type="submit"
                                        disabled={isPending || !watchPhone}
                                        className="btn-primary w-full py-4 text-sm flex items-center justify-center gap-2 group disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        {isPending ? (
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
                                </form>
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
