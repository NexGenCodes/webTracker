import React from 'react';
import { Loader2, Check } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';

interface QRCodeViewProps {
    qrCodeData: string;
    pairStatus: 'idle' | 'waiting' | 'connected';
}

export function QRCodeView({ qrCodeData, pairStatus }: QRCodeViewProps) {
    return (
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
    );
}
