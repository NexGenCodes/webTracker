
import React, { useState } from 'react';
import { Package, CheckCircle, XCircle, Loader2, StickyNote, Play } from 'lucide-react';
import { parseShipmentAI } from '@/app/actions/ai';
import { CreateShipmentDto } from '@/types/shipment';
import { Dictionary } from '@/lib/dictionaries';

interface ShipmentFormProps {
    onSubmit: (data: CreateShipmentDto) => void;
    loading: boolean;
    error: string | null;
    marketingDict: Dictionary;
}

export const ShipmentForm: React.FC<ShipmentFormProps> = ({ onSubmit, loading, error, marketingDict }) => {
    const [aiText, setAiText] = useState('');
    const [isParsing, setIsParsing] = useState(false);
    const [parseError, setParseError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const handleProcess = async () => {
        if (!aiText || aiText.trim().length < 5) {
            setParseError('Please enter shipment details.');
            return;
        }

        setIsParsing(true);
        setParseError(null);
        setSuccessMessage(null);

        try {
            // Step 1: Parse with AI
            const result = await parseShipmentAI(aiText);

            if (!result.success || !result.data) {
                setParseError(result.error || 'Failed to extract data. Please try providing more details.');
                return;
            }

            // Step 2: Validate Data Completeness
            const missingFields = [];
            if (!result.data.senderName) missingFields.push('Sender Name');
            if (!result.data.senderCountry) missingFields.push('Sender Country');
            if (!result.data.receiverName) missingFields.push('Receiver Name');
            if (!result.data.receiverAddress) missingFields.push('Receiver Address');
            if (!result.data.receiverCountry) missingFields.push('Receiver Country');

            if (missingFields.length > 0) {
                setParseError(`Missing information: ${missingFields.join(', ')}. Please update the text and try again.`);
                return;
            }

            // Step 3: Auto-Submit if everything is valid
            // We pass the parsed data to the parent component which handles the DB creation
            onSubmit(result.data);
            setAiText(''); // Clear input on success
            setSuccessMessage(`Manifest processed successfully for ${result.data.receiverName}`);

        } catch (err: any) {
            setParseError('An unexpected error occurred during processing.');
            console.error(err);
        } finally {
            setIsParsing(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto animate-fade-in pb-20">
            <div className="flex items-center gap-3 mb-6">
                <div className="bg-accent/10 p-3 rounded-2xl">
                    <Package className="text-accent" size={24} />
                </div>
                <div>
                    <h2 className="text-xl font-black text-text-main uppercase tracking-tight">Rapid Manifest</h2>
                    <p className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] opacity-60">AI-Powered Entry</p>
                </div>
            </div>

            <div className="glass-panel p-6 border-accent/20 animate-fade-in relative overflow-hidden">
                <h3 className="font-black text-xs uppercase tracking-widest text-text-main mb-3 flex items-center gap-2">
                    <StickyNote size={14} className="text-accent" /> Input Stream
                </h3>
                <p className="text-[10px] text-text-muted mb-4 font-medium">
                    Paste raw shipment details below. The system will automatically parse, validate, and create the manifest.
                </p>

                <div className="relative group mb-6">
                    <div className="absolute -inset-0.5 bg-linear-to-r from-accent/20 to-primary/20 rounded-xl blur opacity-20 group-focus-within:opacity-100 transition duration-500" />
                    <textarea
                        value={aiText}
                        onChange={(e) => setAiText(e.target.value)}
                        placeholder="e.g. Package from Amazon US to John Doe at 123 Main St, London, UK. Phone: +44 7700 900000"
                        className="relative w-full h-40 bg-bg/50 backdrop-blur-sm p-4 rounded-xl border border-border/50 text-sm focus:border-accent/50 outline-none resize-none transition-all font-mono leading-relaxed"
                    />
                </div>

                {parseError && (
                    <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-xl flex items-start gap-3 text-error text-xs font-bold animate-fade-in">
                        <XCircle size={16} className="mt-0.5 shrink-0" />
                        <span>{parseError}</span>
                    </div>
                )}

                {error && (
                    <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-xl flex items-start gap-3 text-error text-xs font-bold animate-fade-in">
                        <XCircle size={16} className="mt-0.5 shrink-0" />
                        <span>System Error: {error}</span>
                    </div>
                )}

                {successMessage && (
                    <div className="mb-6 p-4 bg-success/10 border border-success/20 rounded-xl flex items-start gap-3 text-success text-xs font-bold animate-fade-in">
                        <CheckCircle size={16} className="mt-0.5 shrink-0" />
                        <span>{successMessage}</span>
                    </div>
                )}

                <button
                    onClick={handleProcess}
                    disabled={isParsing || loading || aiText.length < 5}
                    className="btn-primary w-full py-4 text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-3 disabled:opacity-50 disabled:grayscale transition-all active:scale-[0.98]"
                >
                    {isParsing || loading ? (
                        <>
                            <Loader2 size={18} className="animate-spin" />
                            <span>Processing Stream...</span>
                        </>
                    ) : (
                        <>
                            <Play size={18} className="fill-white" />
                            <span>Process & Create Manifest</span>
                        </>
                    )}
                </button>
            </div>
        </div>
    );
};
