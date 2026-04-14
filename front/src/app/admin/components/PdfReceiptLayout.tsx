import React from 'react';
import { APP_NAME } from '@/lib/constants';
import { CreateShipmentDto } from '@/types/shipment';

interface PdfReceiptLayoutProps {
    trackingId: string;
    shipmentData: CreateShipmentDto;
}

export const PdfReceiptLayout = React.forwardRef<HTMLDivElement, PdfReceiptLayoutProps>(({ trackingId, shipmentData }, ref) => {
    // Current date format
    const today = new Date().toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });

    return (
        <div 
            ref={ref}
            className="bg-[#f4f2eb] p-10 font-serif relative"
            style={{ 
                width: '1200px', 
                height: '630px', 
                color: '#2d2d2d',
                backgroundImage: 'radial-gradient(#000 1px, transparent 1px)',
                backgroundSize: '10px 10px',
                // Keep the noise very subtle
            }}
        >
            {/* Background Noise Overlay */}
            <div className="absolute inset-0 bg-white/95 pointer-events-none" />

            {/* ORIGINAL Watermark */}
            <div 
                className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 -rotate-[30deg] font-black pointer-events-none z-0"
                style={{ color: 'rgba(139, 0, 0, 0.04)', fontSize: '200px' }}
            >
                ORIGINAL
            </div>

            <div className="relative z-10 h-full flex flex-col">
                {/* Header */}
                <div className="flex flex-col items-center w-full mb-6">
                    <h1 className="text-[50px] m-0 text-[#8b0000] uppercase font-black tracking-tight leading-none">
                        {APP_NAME}
                    </h1>
                    <div className="w-[60%] h-[24px] bg-[#8b0000] flex items-center justify-center mt-2 mb-2">
                        <span className="text-white text-[11px] font-extrabold tracking-[0.2em] uppercase">
                            INTERNATIONAL SPECIAL DELIVERY SERVICE
                        </span>
                    </div>
                    <span className="text-[16px] font-bold mt-2 tracking-widest">{trackingId}</span>
                </div>

                {/* Info Grid */}
                <div className="flex flex-wrap border border-black w-full bg-white/40">
                    {/* Row 1 */}
                    <div className="flex w-full border-b border-black">
                        <div className="w-1/2 p-4 border-r border-black flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">DESTINATION</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.receiverCountry || 'TBD')}</span>
                        </div>
                        <div className="w-1/2 p-4 flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">ORIGIN</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.senderCountry || 'TBD')}</span>
                        </div>
                    </div>
                    
                    {/* Row 2 */}
                    <div className="flex w-full border-b border-black">
                        <div className="w-1/2 p-4 border-r border-black flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">RECEIVER</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.receiverName || '---')}</span>
                        </div>
                        <div className="w-1/2 p-4 flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">SENDER</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.senderName || '---')}</span>
                        </div>
                    </div>

                    {/* Row 3 */}
                    <div className="flex w-full border-b border-black">
                        <div className="w-1/2 p-4 border-r border-black flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">CONTENT / TYPE</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.cargoType || 'CONSIGNMENT')}</span>
                        </div>
                        <div className="w-1/2 p-4 flex flex-col">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">WEIGHT</span>
                            <span className="text-[18px] text-[#2d2d2d] font-black uppercase mt-1">{(shipmentData.weight || '0.00')} KGS</span>
                        </div>
                    </div>
                    
                    {/* Row 4 (Status) */}
                    <div className="flex w-full">
                        <div className="w-full p-4 flex flex-col items-center bg-[rgba(139,0,0,0.03)] selection:bg-transparent">
                            <span className="text-[10px] text-[#464646] font-extrabold tracking-widest">CURRENT TRACKING STATUS</span>
                            <span className="text-[22px] text-[#8b0000] font-black uppercase mt-1">PENDING</span>
                        </div>
                    </div>
                </div>

                {/* Footer Section */}
                <div className="flex mt-auto justify-between items-end w-full">
                    {/* Barcode Simulation */}
                    <div className="flex gap-[2px]">
                        {[10, 4, 15, 6, 12, 8, 20, 4, 10, 6, 8, 12, 18, 4, 10, 6, 12, 4].map((h, i) => (
                            <div key={i} className="w-[3px] bg-[#2d2d2d]" style={{ height: `${h * 3}px` }} />
                        ))}
                    </div>

                    {/* Approved Stamp */}
                    <div className="border-[3px] border-[rgba(139,0,0,0.4)] rounded-full w-[110px] h-[110px] flex items-center justify-center -rotate-[15deg] text-[rgba(139,0,0,0.4)] font-black text-[14px] text-center leading-tight">
                        SECURED<br/>DIPLOMATIC
                    </div>
                    
                    <div className="flex flex-col items-end">
                        <span className="text-[8px] text-black/40 mb-1 font-sans">OFFICIAL FREIGHT MANIFEST v1.1 • {today}</span>
                        <div className="w-[150px] h-[1px] bg-black" />
                        <span className="text-[12px] font-serif italic text-[#00008b] mt-1">{(shipmentData.senderCountry?.split(',')[0] || 'TBD')} Transit Auth</span>
                    </div>
                </div>
            </div>
        </div>
    );
});

PdfReceiptLayout.displayName = 'PdfReceiptLayout';
