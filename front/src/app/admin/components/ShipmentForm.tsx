import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Package, CheckCircle, XCircle, Loader2, StickyNote, Play, AlertCircle, UploadCloud } from 'lucide-react';
import { parseShipmentAI } from '@/app/actions/ai';
import { CreateShipmentDto } from '@/types/shipment';
import { Dictionary } from '@/lib/dictionaries';
import { shipmentSchema, ShipmentFormData } from '@/lib/schemas/shipment';

interface ShipmentFormProps {
    onSubmit: (data: CreateShipmentDto) => Promise<void>;
    loading: boolean;
    error: string | null;
}

export const ShipmentForm: React.FC<ShipmentFormProps> = ({ onSubmit, loading, error }) => {
    const [aiText, setAiText] = useState('');
    const [isParsing, setIsParsing] = useState(false);
    const [parseError, setParseError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const [showManualForm, setShowManualForm] = useState(false);

    const {
        register,
        handleSubmit,
        setValue,
        reset,
        formState: { errors, isValid }
    } = useForm<ShipmentFormData>({
        resolver: zodResolver(shipmentSchema),
        mode: 'onChange',
        defaultValues: {
            weight: 15,
        }
    });

    const handleAIParsing = async () => {
        if (!aiText || aiText.trim().length < 5) {
            setParseError('Please enter shipment details.');
            return;
        }

        setIsParsing(true);
        setParseError(null);
        setSuccessMessage(null);

        try {
            const result = await parseShipmentAI(aiText);

            if (!result.success || !result.data) {
                setParseError(result.error || 'Failed to extract data. Please try providing more details.');
                return;
            }

            // Populate form with AI results
            Object.entries(result.data).forEach(([key, value]) => {
                if (value) {
                    setValue(key as keyof ShipmentFormData, value as ShipmentFormData[keyof ShipmentFormData]);
                }
            });
            
            setShowManualForm(true);
            setParseError(null);
        } catch (err: unknown) {
            setParseError('An unexpected error occurred during processing.');
            console.error(err);
        } finally {
            setIsParsing(false);
        }
    };

    const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setParseError(null);
        try {
            if (file.type === 'application/pdf' || file.name.endsWith('.pdf')) {
                // Dynamically import pdfjs-dist to prevent SSR issues and keep main bundle small
                const pdfjsLib = await import('pdfjs-dist');
                pdfjsLib.GlobalWorkerOptions.workerSrc = `//cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjsLib.version}/pdf.worker.min.js`;

                const arrayBuffer = await file.arrayBuffer();
                const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
                let text = '';
                // Limit to first 3 pages to save memory and parsing time
                for (let i = 1; i <= Math.min(pdf.numPages, 3); i++) {
                    const page = await pdf.getPage(i);
                    const content = await page.getTextContent();
                    text += content.items.map((item: any) => item.str).join(' ') + '\n';
                }
                setAiText(prev => prev + (prev ? '\n\n' : '') + text);
            } else {
                // For TXT and CSV
                const text = await file.text();
                setAiText(prev => prev + (prev ? '\n\n' : '') + text);
            }
        } catch (err) {
            console.error('File parsing error:', err);
            setParseError('Failed to read file. Ensure it is a valid text, CSV, or PDF.');
        } finally {
            e.target.value = ''; // Reset input
        }
    };

    const onFormSubmit = async (data: ShipmentFormData) => {
        try {
            await onSubmit(data as unknown as CreateShipmentDto);
            setAiText('');
            reset();
            setShowManualForm(false);
            setSuccessMessage(`Manifest created successfully for ${data.receiverName}`);
        } catch (err) {
            console.error(err);
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
                {!showManualForm ? (
                    <>
                        <h3 className="font-black text-xs uppercase tracking-widest text-text-main mb-3 flex items-center gap-2">
                            <StickyNote size={14} className="text-accent" /> Input Stream
                        </h3>
                        <p className="text-[10px] text-text-muted mb-4 font-medium">
                            Paste raw shipment details below. The system will automatically parse and validate.
                        </p>

                        <div className="relative group mb-4">
                            <div className="absolute -inset-0.5 bg-linear-to-r from-accent/20 to-primary/20 rounded-xl blur opacity-20 group-focus-within:opacity-100 transition duration-500" />
                            <textarea
                                value={aiText}
                                onChange={(e) => setAiText(e.target.value)}
                                placeholder="e.g. Package from Amazon US to John Doe at 123 Main St, London, UK. Phone: +44 7700 900000"
                                className="relative w-full h-40 bg-bg/50 backdrop-blur-sm p-4 rounded-xl border border-border/50 text-sm focus:border-accent/50 outline-none resize-none transition-all font-mono leading-relaxed"
                            />
                        </div>

                        <div className="flex items-center gap-2 mb-4">
                            <label className="cursor-pointer group flex items-center justify-center gap-2 py-2 px-4 rounded-xl border border-dashed border-accent/40 hover:border-accent hover:bg-accent/5 transition-all text-xs font-bold text-text-muted hover:text-accent w-full text-center">
                                <UploadCloud size={16} />
                                <span>Upload .TXT, .CSV, or .PDF (Client-Side Parsed)</span>
                                <input 
                                    type="file" 
                                    accept=".txt,.csv,.pdf,application/pdf,text/plain,text/csv" 
                                    className="hidden" 
                                    onChange={handleFileUpload} 
                                />
                            </label>
                        </div>

                        {parseError && (
                            <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-xl flex items-start gap-3 text-error text-xs font-bold animate-fade-in">
                                <XCircle size={16} className="mt-0.5 shrink-0" />
                                <span>{parseError}</span>
                            </div>
                        )}

                        <button
                            onClick={handleAIParsing}
                            disabled={isParsing || aiText.length < 5}
                            className="btn-primary w-full py-4 text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-3 disabled:opacity-50 transition-all active:scale-[0.98]"
                        >
                            {isParsing ? (
                                <>
                                    <Loader2 size={18} className="animate-spin" />
                                    <span>Parsing Stream...</span>
                                </>
                            ) : (
                                <>
                                    <Play size={18} className="fill-white" />
                                    <span>Process Input</span>
                                </>
                            )}
                        </button>
                    </>
                ) : (
                    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4 animate-in fade-in slide-in-from-bottom-2">
                        <div className="flex items-center justify-between mb-2">
                            <h3 className="font-black text-xs uppercase tracking-widest text-text-main flex items-center gap-2">
                                <AlertCircle size={14} className="text-accent" /> Review & Correct
                            </h3>
                            <button 
                                type="button"
                                onClick={() => setShowManualForm(false)}
                                className="text-[10px] underline uppercase tracking-widest text-text-muted hover:text-accent"
                            >
                                Back to Stream
                            </button>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Sender Name</label>
                                <input {...register("senderName")} className="shipment-input" />
                                {errors.senderName && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.senderName.message}</p>}
                            </div>
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Origin Country</label>
                                <input {...register("senderCountry")} className="shipment-input" />
                                {errors.senderCountry && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.senderCountry.message}</p>}
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Receiver Name</label>
                                <input {...register("receiverName")} className="shipment-input" />
                                {errors.receiverName && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.receiverName.message}</p>}
                            </div>
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Receiver Phone</label>
                                <input {...register("receiverPhone")} className="shipment-input" />
                                {errors.receiverPhone && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.receiverPhone.message}</p>}
                            </div>
                        </div>

                        <div className="space-y-1">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Receiver Email</label>
                            <input {...register("receiverEmail")} className="shipment-input" />
                            {errors.receiverEmail && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.receiverEmail.message}</p>}
                        </div>

                        <div className="space-y-1">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Delivery Address</label>
                            <textarea {...register("receiverAddress")} className="shipment-input h-20 resize-none" />
                            {errors.receiverAddress && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.receiverAddress.message}</p>}
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Destination Country</label>
                                <input {...register("receiverCountry")} className="shipment-input" />
                                {errors.receiverCountry && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.receiverCountry.message}</p>}
                            </div>
                            <div className="space-y-1">
                                <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Weight (KG)</label>
                                <input type="number" step="0.1" {...register("weight", { valueAsNumber: true })} className="shipment-input" />
                                {errors.weight && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.weight.message}</p>}
                            </div>
                        </div>

                        <div className="space-y-1">
                            <label className="text-[10px] font-bold uppercase tracking-widest text-text-muted ml-2">Cargo Type</label>
                            <input {...register("cargoType")} placeholder="e.g. Documents, Electronics" className="shipment-input" />
                            {errors.cargoType && <p className="text-error text-[9px] font-bold uppercase ml-2">{errors.cargoType.message}</p>}
                        </div>

                        <button
                            type="submit"
                            disabled={loading || !isValid}
                            className="btn-primary w-full py-4 text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-3 disabled:opacity-50 transition-all active:scale-[0.98]"
                        >
                            {loading ? (
                                <>
                                    <Loader2 size={18} className="animate-spin" />
                                    <span>Syncing with Carrier...</span>
                                </>
                            ) : (
                                <>
                                    <CheckCircle size={18} />
                                    <span>Confirm & Create Manifest</span>
                                </>
                            )}
                        </button>
                    </form>
                )}

                {error && (
                    <div className="mt-4 p-4 bg-error/10 border border-error/20 rounded-xl flex items-start gap-3 text-error text-xs font-bold animate-fade-in">
                        <XCircle size={16} className="mt-0.5 shrink-0" />
                        <span>System Error: {error}</span>
                    </div>
                )}

                {successMessage && (
                    <div className="mt-4 p-4 bg-success/10 border border-success/20 rounded-xl flex items-start gap-3 text-success text-xs font-bold animate-fade-in">
                        <CheckCircle size={16} className="mt-0.5 shrink-0" />
                        <span>{successMessage}</span>
                    </div>
                )}
            </div>

            <style jsx>{`
                .shipment-input {
                    width: 100%;
                    background: var(--glass-bg);
                    backdrop-filter: blur(10px);
                    padding: 0.75rem 1rem;
                    border-radius: 0.75rem;
                    border: 1px solid var(--color-border);
                    color: var(--color-text-main);
                    font-size: 0.875rem;
                    transition: all 0.2s;
                    outline: none;
                }
                .shipment-input:focus {
                    border-color: var(--color-accent);
                    box-shadow: 0 0 0 2px var(--color-accent-transparent);
                }
            `}</style>
        </div>
    );
};
