'use client';

import React, { useState } from 'react';
import { Sparkles, Loader2, Clipboard, CheckCircle2, Package, MapPin, AlertCircle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { parseManifestAction } from '@/app/actions/shipment';
import { createShipmentAction } from '@/app/actions/shipment';
import toast from 'react-hot-toast';

interface AIParserTabProps {
    companyId: string;
}

interface ParsedData {
    receiverName?: string;
    receiverPhone?: string;
    receiverAddress?: string;
    receiverCountry?: string;
    receiverEmail?: string;
    receiverID?: string;
    senderName?: string;
    senderCountry?: string;
    cargoType?: string;
    weight?: number;
}

export function AIParserTab({ companyId }: AIParserTabProps) {
    const [text, setText] = useState('');
    const [isParsing, setIsParsing] = useState(false);
    const [parsedData, setParsedData] = useState<ParsedData | null>(null);
    const [isCreating, setIsCreating] = useState(false);

    const handleParse = async () => {
        if (!text.trim()) {
            toast.error('Please paste manifest text first');
            return;
        }

        setIsParsing(true);
        try {
            const result = await parseManifestAction(text);
            if (result.success && 'data' in result && result.data) {
                setParsedData(result.data as ParsedData);
                toast.success('Manifest parsed successfully!');
            } else {
                toast.error(result.error || 'Failed to parse manifest');
            }
        } catch {
            toast.error('An unexpected error occurred');
        } finally {
            setIsParsing(false);
        }
    };

    const handleCreate = async () => {
        if (!parsedData) return;

        setIsCreating(true);
        try {
            // Map parsed data to shipment form fields
            const shipmentData = {
                recipient_name: parsedData.receiverName || '',
                recipient_phone: parsedData.receiverPhone || '',
                recipient_address: parsedData.receiverAddress || '',
                destination: parsedData.receiverCountry || '',
                sender_name: parsedData.senderName || 'AI Manifest',
                origin: parsedData.senderCountry || 'N/A',
                weight: Number(parsedData.weight) || 0.1,
                cargo_type: parsedData.cargoType || 'General Cargo',
            };

            const result = await createShipmentAction(companyId, shipmentData);
            if (result.success) {
                toast.success('Shipment created successfully!');
                setParsedData(null);
                setText('');
            } else {
                toast.error(result.error || 'Failed to create shipment');
            }
        } catch {
            toast.error('Failed to create shipment');
        } finally {
            setIsCreating(false);
        }
    };

    return (
        <div className="max-w-4xl mx-auto space-y-8 pb-20">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-black text-text-main uppercase tracking-tighter flex items-center gap-3">
                        <Sparkles className="text-accent" /> AI Manifest Parser
                    </h2>
                    <p className="text-sm font-bold text-text-muted uppercase tracking-widest mt-1">
                        Paste manifest text to extract shipment details instantly
                    </p>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Input Section */}
                <div className="space-y-4">
                    <div className="glass-panel p-6 border-border/50 relative overflow-hidden group">
                        <div className="absolute inset-0 bg-accent/5 opacity-0 group-focus-within:opacity-100 transition-opacity" />
                        <label className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-3 block ml-1">
                            Manifest Content
                        </label>
                        <textarea
                            value={text}
                            onChange={(e) => setText(e.target.value)}
                            placeholder="Paste your manifest, invoice text, or shipping list here..."
                            className="w-full h-[300px] bg-transparent border-none outline-none resize-none text-sm font-medium text-text-main custom-scrollbar placeholder:text-text-muted/40"
                        />
                        <div className="mt-4 flex items-center justify-between border-t border-border/20 pt-4">
                            <button
                                onClick={() => setText('')}
                                className="text-[10px] font-black uppercase tracking-widest text-text-muted hover:text-error transition-colors"
                            >
                                Clear Text
                            </button>
                            <button
                                onClick={handleParse}
                                disabled={isParsing || !text.trim()}
                                className="btn-primary py-3 px-6 text-xs flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {isParsing ? (
                                    <><Loader2 className="animate-spin" size={14} /> Parsing...</>
                                ) : (
                                    <><Sparkles size={14} /> Extract Data</>
                                )}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Results Section */}
                <div className="space-y-4">
                    <AnimatePresence mode="wait">
                        {!parsedData ? (
                            <motion.div
                                initial={{ opacity: 0, scale: 0.95 }}
                                animate={{ opacity: 1, scale: 1 }}
                                exit={{ opacity: 0, scale: 0.95 }}
                                className="h-full min-h-[400px] glass-panel border-dashed border-border/50 flex flex-col items-center justify-center p-12 text-center"
                            >
                                <div className="w-16 h-16 rounded-3xl bg-surface border border-border flex items-center justify-center text-text-muted mb-6 shadow-inner">
                                    <Clipboard size={24} />
                                </div>
                                <h3 className="text-sm font-black text-text-main uppercase tracking-widest">No Data Extracted</h3>
                                <p className="text-xs font-bold text-text-muted mt-2 max-w-[200px]">
                                    Paste text and click extract to see the magic happen.
                                </p>
                            </motion.div>
                        ) : (
                            <motion.div
                                initial={{ opacity: 0, x: 20 }}
                                animate={{ opacity: 1, x: 0 }}
                                exit={{ opacity: 0, x: -20 }}
                                className="glass-panel p-8 border-accent/20 bg-accent/5 relative overflow-hidden flex flex-col h-full"
                            >
                                <div className="absolute top-0 right-0 p-4">
                                    <div className="px-3 py-1 rounded-full bg-success/10 border border-success/20 text-[10px] font-black text-success uppercase tracking-widest flex items-center gap-1.5">
                                        <CheckCircle2 size={10} /> Data Ready
                                    </div>
                                </div>

                                <div className="flex-1 space-y-6">
                                    <div>
                                        <h3 className="text-xs font-black text-text-main uppercase tracking-widest mb-4 flex items-center gap-2">
                                            <Package size={14} className="text-accent" /> Extracted Fields
                                        </h3>

                                        <div className="space-y-4">
                                            <div className="grid grid-cols-2 gap-4">
                                                <div className="p-4 bg-surface rounded-2xl border border-border/50">
                                                    <p className="text-[9px] font-black text-text-muted uppercase tracking-widest mb-1">Receiver</p>
                                                    <p className="text-xs font-black text-text-main truncate">{parsedData.receiverName || 'Not Found'}</p>
                                                </div>
                                                <div className="p-4 bg-surface rounded-2xl border border-border/50">
                                                    <p className="text-[9px] font-black text-text-muted uppercase tracking-widest mb-1">Phone</p>
                                                    <p className="text-xs font-black text-text-main truncate">{parsedData.receiverPhone || 'Not Found'}</p>
                                                </div>
                                            </div>

                                            <div className="p-4 bg-surface rounded-2xl border border-border/50">
                                                <div className="flex items-center gap-2 mb-1">
                                                    <MapPin size={10} className="text-accent" />
                                                    <p className="text-[9px] font-black text-text-muted uppercase tracking-widest">Destination</p>
                                                </div>
                                                <p className="text-xs font-black text-text-main">{parsedData.receiverCountry || 'Not Found'}</p>
                                            </div>

                                            <div className="grid grid-cols-2 gap-4">
                                                <div className="p-4 bg-surface rounded-2xl border border-border/50">
                                                    <p className="text-[9px] font-black text-text-muted uppercase tracking-widest mb-1">Weight</p>
                                                    <p className="text-xs font-black text-text-main">{parsedData.weight ? `${parsedData.weight}kg` : 'N/A'}</p>
                                                </div>
                                                <div className="p-4 bg-surface rounded-2xl border border-border/50">
                                                    <p className="text-[9px] font-black text-text-muted uppercase tracking-widest mb-1">Origin</p>
                                                    <p className="text-xs font-black text-text-main truncate">{parsedData.senderCountry || 'N/A'}</p>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="p-4 rounded-2xl bg-warning/5 border border-warning/20 flex gap-3">
                                        <AlertCircle size={16} className="text-warning shrink-0" />
                                        <p className="text-[10px] font-bold text-text-muted leading-tight">
                                            Please verify the extracted data before creating the shipment. AI extraction may vary based on text quality.
                                        </p>
                                    </div>
                                </div>

                                <div className="mt-8 pt-6 border-t border-border/20 flex gap-3">
                                    <button
                                        onClick={() => setParsedData(null)}
                                        className="flex-1 px-6 py-4 bg-surface hover:bg-surface-muted text-text-muted hover:text-text-main rounded-2xl border border-border text-[10px] font-black uppercase tracking-widest transition-all"
                                    >
                                        Discard
                                    </button>
                                    <button
                                        onClick={handleCreate}
                                        disabled={isCreating}
                                        className="flex-[2] btn-primary py-4 text-[10px] flex items-center justify-center gap-2"
                                    >
                                        {isCreating ? (
                                            <><Loader2 className="animate-spin" size={14} /> Creating...</>
                                        ) : (
                                            <><CheckCircle2 size={14} /> Confirm & Create</>
                                        )}
                                    </button>
                                </div>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
            </div>
        </div>
    );
}
