import React from 'react';
import { Check, Copy, ChevronLeft, Share2 } from 'lucide-react';
import { toast } from 'react-hot-toast';

import { Dictionary, CreateShipmentDto } from '@/types/shipment';

interface SuccessDisplayProps {
    trackingId: string;
    shipmentData: CreateShipmentDto;
    copied: boolean;
    onCopy: () => void;
    onBack: () => void;
    dict: Dictionary;
}

export const SuccessDisplay: React.FC<SuccessDisplayProps> = ({
    trackingId,
    shipmentData,
    copied,
    onCopy,
    onBack,
    dict
}) => {
    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center animate-fade-in space-y-6 sm:space-y-8">
            <div className="space-y-2">
                <h2 className="text-2xl sm:text-3xl font-black text-success tracking-tight uppercase">{dict.admin.success}</h2>
                <p className="text-text-muted font-medium text-sm sm:text-base">{dict.admin.successDesc}</p>
            </div>

            <div className="w-full max-w-2xl bg-surface border border-border rounded-2xl p-6 sm:p-8 shadow-xl flex flex-col gap-6 relative overflow-hidden text-left">
                <div className="absolute top-0 right-0 w-64 h-64 bg-success/5 rounded-full -mr-32 -mt-32 blur-3xl pointer-events-none" />

                <div className="flex flex-col items-center gap-2 border-b border-border pb-6 text-center">
                    <span className="text-xs text-text-muted uppercase tracking-widest font-black">{dict.shipment.trackingId}</span>
                    <span className="text-3xl sm:text-5xl font-mono text-text-main font-black tracking-widest break-all">
                        {trackingId}
                    </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 relative z-10">
                    <div className="space-y-4">
                        <h4 className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Sender Details</h4>
                        <div className="space-y-1">
                            <p className="font-bold text-text-main">{shipmentData.senderName}</p>
                            <p className="text-sm text-text-muted">{shipmentData.senderCountry}</p>
                        </div>
                    </div>

                    <div className="space-y-4">
                        <h4 className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Receiver Details</h4>
                        <div className="space-y-1">
                            <p className="font-bold text-text-main">{shipmentData.receiverName}</p>
                            <p className="text-sm text-text-muted">{shipmentData.receiverPhone}</p>
                            <p className="text-sm text-text-muted">{shipmentData.receiverAddress}</p>
                            <p className="text-sm text-text-muted">{shipmentData.receiverCountry}</p>
                        </div>
                    </div>

                    <div className="sm:col-span-2 grid grid-cols-2 gap-4 pt-4 border-t border-border">
                        <div>
                            <span className="block text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-1">Weight</span>
                            <span className="font-bold text-text-main">{shipmentData.weight} KG</span>
                        </div>
                        <div>
                            <span className="block text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-1">Cargo Type</span>
                            <span className="font-bold text-text-main">{shipmentData.cargoType || 'Standard'}</span>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex flex-col w-full max-w-sm gap-4">
                <button
                    onClick={() => {
                        const shareUrl = `${window.location.origin}/api/receipt/${trackingId}?status=PENDING&origin=${shipmentData.senderCountry}&dest=${shipmentData.receiverCountry}&sender=${shipmentData.senderName}&receiver=${shipmentData.receiverName}&weight=${shipmentData.weight}%20KGS&content=${shipmentData.cargoType}`;
                        if (navigator.share) {
                            navigator.share({
                                title: `Shipment ${trackingId}`,
                                text: `New shipment created: ${trackingId}`,
                                url: shareUrl
                            });
                        } else {
                            navigator.clipboard.writeText(shareUrl);
                            toast.success("Receipt link copied!");
                        }
                    }}
                    className="btn-primary bg-accent hover:bg-accent/80 text-white flex items-center justify-center gap-2 py-3 sm:py-4 text-base sm:text-lg w-full"
                >
                    <Share2 size={20} />
                    Share Receipt
                </button>
                <button
                    onClick={onCopy}
                    className="bg-surface-muted hover:bg-surface border border-border text-text-main flex items-center justify-center gap-2 py-3 sm:py-4 text-base sm:text-lg w-full rounded-xl transition-all font-black uppercase tracking-widest"
                >
                    {copied ? <Check /> : <Copy />}
                    {copied ? dict.admin.copied : dict.admin.copy + " ID"}
                </button>
                <button
                    onClick={onBack}
                    className="flex items-center justify-center gap-2 text-gray-400 hover:text-white py-2 transition-colors text-sm sm:text-base"
                >
                    <ChevronLeft size={20} />
                    {dict.admin.createAnother}
                </button>
            </div>
        </div>
    );
};
