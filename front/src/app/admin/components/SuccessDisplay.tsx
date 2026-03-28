import React, { useState, useRef } from 'react';
import { Check, Copy, ChevronLeft, Download, RefreshCw, Eye, Image as ImageIcon, Edit3 } from 'lucide-react';
import { toast } from 'react-hot-toast';
import html2canvas from 'html2canvas';

import { Dictionary, CreateShipmentDto } from '@/types/shipment';
import { ReceiptPreview, ReceiptData } from './ReceiptPreview';

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
    const [imageError, setImageError] = useState(false);
    const [viewMode, setViewMode] = useState<'comparison' | 'frontend' | 'legacy'>('comparison');
    const [isExporting, setIsExporting] = useState(false);
    const receiptRef = useRef<HTMLDivElement>(null);

    // Local editable state for the receipt Hub
    const [editableData, setEditableData] = useState<ReceiptData>({
        trackingId: trackingId,
        senderName: shipmentData.senderName,
        senderCountry: shipmentData.senderCountry || 'TBD',
        receiverName: shipmentData.receiverName,
        receiverCountry: shipmentData.receiverCountry || 'TBD',
        receiverPhone: shipmentData.receiverPhone || 'N/A',
        receiverAddress: shipmentData.receiverAddress || 'N/A',
        weight: shipmentData.weight ?? 0,
        cargoType: shipmentData.cargoType || 'CONSIGNMENT BOX',
        status: 'PENDING',
        noiseSeed: Math.floor(Math.random() * 1000)
    });

    const receiptUrl = `/api/receipt/${trackingId}?status=${editableData.status}&origin=${editableData.senderCountry}&dest=${editableData.receiverCountry}&sender=${editableData.senderName}&receiver=${editableData.receiverName}&weight=${editableData.weight}%20KGS&content=${editableData.cargoType}`;

    const handleDownloadFrontend = async () => {
        if (!receiptRef.current) return;
        setIsExporting(true);
        try {
            const canvas = await html2canvas(receiptRef.current, {
                scale: 1, // Already 1800x1400
                useCORS: true,
                backgroundColor: '#f4f2eb',
            });
            const link = document.createElement('a');
            link.download = `Receipt-${trackingId}.png`;
            link.href = canvas.toDataURL('image/png');
            link.click();
            toast.success("Receipt downloaded!");
        } catch {
            toast.error("Export failed");
        } finally {
            setIsExporting(false);
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center animate-fade-in space-y-6 sm:space-y-8">
            <div className="space-y-2">
                <h2 className="text-2xl sm:text-3xl font-black text-success tracking-tight uppercase">{dict.admin.success}</h2>
                <p className="text-text-muted font-medium text-sm sm:text-base">{dict.admin.successDesc}</p>
            </div>

            {/* Hub Controls */}
            <div className="flex flex-wrap items-center justify-center gap-2 mb-4">
                <button 
                    onClick={() => setViewMode('comparison')}
                    className={`px-4 py-2 rounded-lg text-xs font-black uppercase tracking-widest flex items-center gap-2 transition-all ${viewMode === 'comparison' ? 'bg-primary text-white scale-105 shadow-lg' : 'bg-surface-muted text-text-muted hover:bg-surface'}`}
                >
                    <Eye size={14} /> Side-by-Side
                </button>
                <button 
                    onClick={() => setViewMode('frontend')}
                    className={`px-4 py-2 rounded-lg text-xs font-black uppercase tracking-widest flex items-center gap-2 transition-all ${viewMode === 'frontend' ? 'bg-accent text-white scale-105 shadow-lg' : 'bg-surface-muted text-text-muted hover:bg-surface'}`}
                >
                    <Edit3 size={14} /> Admin Hub (React)
                </button>
                <button 
                    onClick={() => setViewMode('legacy')}
                    className={`px-4 py-2 rounded-lg text-xs font-black uppercase tracking-widest flex items-center gap-2 transition-all ${viewMode === 'legacy' ? 'bg-black text-white scale-105 shadow-lg' : 'bg-surface-muted text-text-muted hover:bg-surface'}`}
                >
                    <ImageIcon size={14} /> Legacy Image (Go)
                </button>
            </div>

            <div className={`w-full ${viewMode === 'comparison' ? 'max-w-7xl' : 'max-w-4xl'} bg-surface border border-border rounded-2xl p-4 sm:p-6 shadow-2xl transition-all duration-500`}>
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                    
                    {/* HUB: Data Tweak Panel */}
                    {(viewMode === 'frontend' || viewMode === 'comparison') && (
                        <div className="lg:col-span-3 space-y-4 text-left p-4 bg-bg rounded-xl border border-border">
                            <h3 className="text-xs font-black uppercase tracking-[0.2em] text-primary mb-4">Receipt Controls</h3>
                            
                            <div className="space-y-2">
                                <label className="text-[10px] uppercase font-black opacity-50">Receiver Name</label>
                                <input 
                                    className="w-full bg-surface border border-border rounded p-2 text-sm font-bold"
                                    value={editableData.receiverName}
                                    onChange={(e) => setEditableData({...editableData, receiverName: e.target.value})}
                                />
                            </div>
                            
                            <div className="space-y-2">
                                <label className="text-[10px] uppercase font-black opacity-50">Gross Weight (KG)</label>
                                <input 
                                    type="number"
                                    className="w-full bg-surface border border-border rounded p-2 text-sm font-bold"
                                    value={editableData.weight}
                                    onChange={(e) => setEditableData({...editableData, weight: parseFloat(e.target.value)})}
                                />
                            </div>

                            <button 
                                onClick={() => setEditableData({...editableData, noiseSeed: Math.floor(Math.random() * 1000)})}
                                className="w-full py-2 bg-surface hover:bg-surface-muted border border-border rounded flex items-center justify-center gap-2 text-[10px] font-black uppercase transition-colors"
                            >
                                <RefreshCw size={12} /> Shift Noise/Stamp
                            </button>

                            <button 
                                onClick={handleDownloadFrontend}
                                disabled={isExporting}
                                className="w-full py-3 bg-accent text-white rounded-lg flex items-center justify-center gap-2 text-xs font-black uppercase tracking-widest hover:bg-black transition-all shadow-lg active:scale-95 disabled:opacity-50"
                            >
                                {isExporting ? <RefreshCw className="animate-spin" /> : <Download size={14} />}
                                Download Hub PNG
                            </button>
                        </div>
                    )}

                    {/* RECEIPT VIEWPORT */}
                    <div className={`${(viewMode === 'frontend' || viewMode === 'comparison') ? 'lg:col-span-9' : 'lg:col-span-12'} space-y-6`}>
                        <div className={`grid grid-cols-1 ${viewMode === 'comparison' ? 'md:grid-cols-2' : ''} gap-6`}>
                            
                            {/* Option 1: Frontend Perfect Clone */}
                            {(viewMode === 'frontend' || viewMode === 'comparison') && (
                                <div className="space-y-2">
                                    <span className="text-[10px] font-black uppercase text-accent bg-accent/10 px-2 py-1 rounded">Option 1: React Clone (0% VPS Load)</span>
                                    <div className="border border-border rounded-xl overflow-hidden bg-bg shadow-lg transform origin-top-left" style={{ scale: viewMode === 'comparison' ? '0.4' : '0.6' }}>
                                        <ReceiptPreview ref={receiptRef} data={editableData} />
                                    </div>
                                </div>
                            )}

                            {/* Option 2: Legacy/Optimized Go Interface */}
                            {(viewMode === 'legacy' || viewMode === 'comparison') && (
                                <div className="space-y-2">
                                    <span className="text-[10px] font-black uppercase text-primary bg-primary/10 px-2 py-1 rounded">Option 2: Optimized Go (Template Rend)</span>
                                    <div className="border border-border rounded-xl overflow-hidden bg-black flex flex-col min-h-[400px]">
                                        {imageError ? (
                                            <div className="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-4">
                                                <p className="text-error text-xs font-bold uppercase tracking-widest">Backend Generation Failed</p>
                                                <button onClick={() => setImageError(false)} className="btn-primary py-2 px-4 text-xs"><RefreshCw size={14} /> Retry</button>
                                            </div>
                                        ) : (
                                            <img 
                                                src={receiptUrl} 
                                                alt="Go Backend Receipt" 
                                                className="w-full object-contain"
                                                onError={() => setImageError(true)}
                                            />
                                        )}
                                        <div className="p-3 bg-surface-muted border-t border-border flex gap-2 justify-end mt-auto">
                                            <a 
                                                href={receiptUrl}
                                                download={`Legacy-Receipt-${trackingId}.png`}
                                                className="text-[10px] px-3 py-2 bg-primary text-white rounded hover:bg-black transition-colors font-black uppercase"
                                            >
                                                Download Go PNG
                                            </a>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex flex-col w-full max-w-sm gap-4">
                <button
                    onClick={onCopy}
                    className="bg-surface-muted hover:bg-surface border border-border text-text-main flex items-center justify-center gap-2 py-3 sm:py-4 text-base sm:text-lg w-full rounded-xl transition-all font-black uppercase tracking-widest"
                >
                    {copied ? <Check /> : <Copy />}
                    {copied ? dict.admin.copied : dict.admin.copy + " ID"}
                </button>
                <button
                    onClick={onBack}
                    className="flex items-center justify-center gap-2 text-gray-400 hover:text-white py-2 transition-colors text-sm sm:text-base font-bold uppercase tracking-widest"
                >
                    <ChevronLeft size={20} />
                    {dict.admin.createAnother}
                </button>
            </div>
        </div>
    );
};

