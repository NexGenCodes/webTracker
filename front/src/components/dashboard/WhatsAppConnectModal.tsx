import React, { useState, useEffect, useTransition, useCallback } from 'react';
import { Smartphone, XCircle, AlertTriangle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { pairWhatsApp, getWhatsAppQR } from '@/app/actions/setup';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import { createClient } from '@/lib/supabase/client';

import { QRCodeView } from './whatsapp/QRCodeView';
import { PhoneConnectForm, type PhoneFormValues } from './whatsapp/PhoneConnectForm';
import { PhoneCodeView } from './whatsapp/PhoneCodeView';

const formatLocalPhone = (countryCode: string, phone: string) => {
    let localPhone = phone.replace(/[\s\-()]/g, '');
    if (localPhone.startsWith('0')) {
        localPhone = localPhone.substring(1);
    }
    return localPhone.startsWith('+') ? localPhone : `${countryCode}${localPhone}`;
};

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

    const [isPending, startTransition] = useTransition();

    const handleConnected = useCallback(() => {
        setPairStatus(prev => {
            if (prev === 'connected') return prev;
            setTimeout(() => {
                onSuccess();
                onClose();
            }, 1000);
            return 'connected';
        });
        setPairError('');
    }, [onSuccess, onClose]);

    const handleFetchQR = useCallback(() => {
        if (!companyId) return;
        startTransition(async () => {
            const response = await getWhatsAppQR(companyId);
            if (response.success && response.data?.pairingCode) {
                setQrCodeData(response.data.pairingCode);
                setPairStatus(prev => prev === 'idle' ? 'waiting' : prev);
            } else {
                const msg = response.error || 'Could not fetch QR code.';
                if (msg.toLowerCase().includes('already connected')) {
                    handleConnected();
                } else {
                    setPairError(msg);
                }
            }
        });
    }, [companyId, handleConnected]);

    useEffect(() => {
        if (isOpen) {
            setPairingCode('');
            setQrCodeData('');
            setPairError('');
            setPairStatus('idle');
            setConnectMode('qr');
            handleFetchQR();
        }
    }, [isOpen, handleFetchQR]);

    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (isOpen && connectMode === 'qr' && pairStatus === 'waiting') {
            interval = setInterval(() => {
                handleFetchQR();
            }, 20000);
        }
        return () => {
            if (interval) clearInterval(interval);
        };
    }, [isOpen, connectMode, pairStatus, handleFetchQR]);

    useEffect(() => {
        if (!isOpen || !companyId) return;
        if (companyData?.auth_status === 'active') {
            handleConnected();
        }
    }, [isOpen, companyId, companyData?.auth_status, handleConnected]);

    useEffect(() => {
        if (!isOpen || !companyId) return;

        const supabase = createClient();
        const channelName = `company_status_modal_${companyId}`;
        const channel = supabase
            .channel(channelName)
            .on(
                'postgres_changes',
                {
                    event: 'UPDATE',
                    schema: 'public',
                    table: 'companies',
                    filter: `id=eq.${companyId}`,
                },
                (payload) => {
                    if (payload.new.auth_status === 'active') {
                        handleConnected();
                    }
                }
            )
            .subscribe();

        return () => {
            supabase.removeChannel(channel);
        };
    }, [isOpen, companyId, handleConnected]);

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

    const onSubmitPhoneForm = (data: PhoneFormValues) => {
        setPairError('');

        if (!companyId) {
            setPairError('Session expired. Please refresh the page and try again.');
            return;
        }

        const subStatus = companyData?.subscription_status ?? 'active';
        if (subStatus !== 'active' && subStatus !== 'trialing') {
            setPairError('Subscription is inactive. Please renew to use the tracking bot.');
            return;
        }

        const fullNumber = formatLocalPhone(data.countryCode, data.phone);
        const phoneNumber = parsePhoneNumberFromString(fullNumber);

        if (!phoneNumber) {
            setPairError('Invalid phone number structure.');
            return;
        }

        const formattedPhone = phoneNumber.number.replace('+', '');

        startTransition(async () => {
            const response = await pairWhatsApp(companyId, formattedPhone);

            if (response.success && response.data?.code) {
                setPairingCode(response.data.code);
                setPairStatus('waiting');
            } else {
                const msg = response.error || 'Code not received. Check phone format.';
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
                    className="fixed inset-0 z-[9999] flex items-center justify-center p-4 backdrop-blur-sm bg-background/80"
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
                                <QRCodeView qrCodeData={qrCodeData} pairStatus={pairStatus} />
                            ) : !pairingCode ? (
                                <PhoneConnectForm isPending={isPending} onSubmit={onSubmitPhoneForm} />
                            ) : (
                                <PhoneCodeView
                                    pairingCode={pairingCode}
                                    pairStatus={pairStatus}
                                    handleCopyCode={handleCopyCode}
                                    codeCopied={codeCopied}
                                    onRegenerate={() => setPairStatus('idle')}
                                    companyId={companyId}
                                    handleConnected={handleConnected}
                                />
                            )}
                        </div>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}
