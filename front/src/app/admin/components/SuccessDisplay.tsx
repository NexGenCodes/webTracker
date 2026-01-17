import React from 'react';
import { Check, Copy, ChevronLeft } from 'lucide-react';

interface SuccessDisplayProps {
    trackingId: string;
    copied: boolean;
    onCopy: () => void;
    onBack: () => void;
    dict: any;
}

export const SuccessDisplay: React.FC<SuccessDisplayProps> = ({
    trackingId,
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

            <div className="w-full max-w-md bg-surface border border-border rounded-2xl p-6 sm:p-8 shadow-xl flex flex-col items-center gap-4 relative overflow-hidden">
                <div className="absolute top-0 right-0 w-32 h-32 bg-success/5 rounded-full -mr-16 -mt-16 blur-2xl" />
                <span className="text-xs text-text-muted uppercase tracking-widest font-black">{dict.shipment.trackingId}</span>
                <span className="text-2xl sm:text-4xl font-mono text-text-main font-black tracking-widest break-all">
                    {trackingId}
                </span>
            </div>

            <div className="flex flex-col w-full max-w-sm gap-4">
                <button
                    onClick={onCopy}
                    className="btn-primary flex items-center justify-center gap-2 py-3 sm:py-4 text-base sm:text-lg w-full"
                >
                    {copied ? <Check /> : <Copy />}
                    {copied ? dict.admin.copied : dict.admin.copy}
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
