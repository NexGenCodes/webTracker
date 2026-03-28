import React from 'react';

export interface ReceiptData {
    trackingId: string;
    senderName: string;
    senderCountry: string;
    receiverName: string;
    receiverCountry: string;
    receiverPhone: string;
    receiverAddress: string;
    weight: number;
    cargoType: string;
    status: string;
    noiseSeed?: number;
}

interface ReceiptPreviewProps {
    data: ReceiptData;
}

export const ReceiptPreview = React.forwardRef<HTMLDivElement, ReceiptPreviewProps>(({ data }, ref) => {
    const today = new Date().toLocaleDateString('en-US', {
        year: 'numeric', month: 'short', day: 'numeric'
    });

    const noiseOpacity = 0.02 + ((data.noiseSeed || 0) % 3) / 100;
    const rotateStamp = -20 + ((data.noiseSeed || 0) % 10); // Matches V11Stamps

    return (
        <div 
            ref={ref}
            className="bg-[#f4f2eb] p-12 font-serif relative overflow-hidden select-none"
            style={{ 
                width: '1800px', 
                height: '1400px', 
                color: '#2d2d2d',
                backgroundImage: 'radial-gradient(#000 1px, transparent 1px)',
                backgroundSize: '15px 15px',
            }}
        >
            {/* Background Texture Overlay */}
            <div 
                className="absolute inset-0 pointer-events-none z-0" 
                style={{ 
                    opacity: noiseOpacity,
                    backgroundColor: '#000',
                    mixBlendMode: 'overlay',
                }} 
            />

            {/* ORIGINAL Watermark (Go matches -30deg) */}
            <div 
                className="absolute top-1/2 left-1/2 -rotate-[30deg] font-black pointer-events-none z-0"
                style={{ 
                    color: 'rgba(139, 0, 0, 0.05)', 
                    fontSize: '280px', 
                    letterSpacing: '0.05em',
                    transform: 'translate(-50%, -50%)'
                }}
            >
                ORIGINAL
            </div>

            <div className="relative z-10 h-full flex flex-col">
                {/* Header Section (yH = 160.0 in Go) */}
                <div className="flex flex-col items-center w-full mb-10 pt-10">
                    <h1 className="text-[90px] m-0 text-[#8b0000] uppercase font-black tracking-tight leading-none italic">
                        Airway Bill
                    </h1>
                    <div className="w-[800px] h-[45px] bg-[#8b0000] flex items-center justify-center mt-6 mb-4">
                        <span className="text-white text-[20px] font-black tracking-[0.4em] uppercase">
                            INTERNATIONAL SPECIAL DELIVERY SERVICE
                        </span>
                    </div>
                    <div className="flex justify-between w-full px-20">
                         <span className="text-[26px] font-black tracking-[0.3em] font-mono">
                            {data.trackingId}
                        </span>
                        <span className="text-[40px] font-black text-[#cc0000] font-mono">
                            № 00{data.trackingId}
                        </span>
                    </div>
                </div>

                {/* Main Information Grid (gX=20, gY=420, gW=Width-40, gH=624) */}
                <div className="flex flex-wrap border-[3px] border-black w-full bg-white/40">
                    {/* Destination/Origin Row */}
                    <div className="flex w-full border-b-[3px] border-black">
                        <div className="w-[30%] p-8 border-r-[3px] border-black flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">DESTINATION</span>
                            <span className="text-[32px] text-[#2d2d2d] font-black uppercase leading-tight">{(data.receiverCountry || 'TBD')}</span>
                        </div>
                        <div className="w-[44%] border-r-[3px] border-black grid grid-cols-2">
                             <div className="p-4 border-b border-black text-center bg-black text-white flex items-center justify-center font-bold">SERVICE</div>
                             <div className="p-4 border-b border-black text-center bg-black text-white flex items-center justify-center font-bold">PAYMENT</div>
                             <div className="p-6 flex flex-col items-center border-r border-black relative">
                                 <span className="text-sm font-black opacity-50 mb-2">DIPLOMATIC</span>
                                 <div className="w-16 h-16 border-4 border-black flex items-center justify-center font-black text-4xl">X</div>
                             </div>
                             <div className="p-6 flex flex-col items-center relative">
                                 <span className="text-sm font-black opacity-50 mb-2">ACCOUNT</span>
                                 <div className="w-16 h-16 border-4 border-black flex items-center justify-center font-black text-4xl">X</div>
                             </div>
                        </div>
                        <div className="w-[26%] p-8 flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">ORIGIN</span>
                            <span className="text-[32px] text-[#2d2d2d] font-black uppercase leading-tight">{(data.senderCountry || 'TBD')}</span>
                        </div>
                    </div>
                    
                    {/* Participants Row */}
                    <div className="flex w-full border-b-[3px] border-black">
                        <div className="w-[52%] p-8 border-r-[3px] border-black flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">CONSIGNEE (RECEIVER DETAILS)</span>
                            <span className="text-[36px] text-[#2d2d2d] font-black uppercase mb-1">{(data.receiverName || '---')}</span>
                            <span className="text-[20px] font-bold text-black/70 italic">{(data.receiverAddress || 'ADDR: N/A')}</span>
                        </div>
                        <div className="w-[48%] p-8 flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">CONSIGNOR (SENDER)</span>
                            <span className="text-[36px] text-[#2d2d2d] font-black uppercase">{(data.senderName || '---')}</span>
                            <span className="text-[20px] font-bold text-black/50 mt-2">SENDER ID: {data.trackingId.substring(0,8)}</span>
                        </div>
                    </div>

                    {/* Cargo Specifications Row */}
                    <div className="flex w-full border-b-[3px] border-black">
                        <div className="w-[30%] p-8 border-r-[3px] border-black flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">DESCRIPTION</span>
                            <span className="text-[28px] text-[#2d2d2d] font-black uppercase">{(data.cargoType || 'CONSIGNMENT')}</span>
                        </div>
                        <div className="w-[22%] p-8 border-r-[3px] border-black flex flex-col bg-accent/5">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">GROSS WT</span>
                            <span className="text-[42px] text-[#8b0000] font-black leading-none">{(data.weight || '0.00')} <span className="text-xl">KGS</span></span>
                        </div>
                        <div className="w-[22%] p-8 border-r-[3px] border-black flex flex-col">
                            <span className="text-[18px] text-[#464646] font-black tracking-widest uppercase mb-2">DEP DATE</span>
                            <span className="text-[28px] text-[#2d2d2d] font-black uppercase">{today}</span>
                        </div>
                        <div className="w-[26%] p-8 flex flex-col items-center justify-center">
                             <div className="bg-[#cc0000] text-white p-2 w-full text-center text-sm font-black tracking-widest mb-2">CONFIDENTIAL</div>
                             <span className="text-[12px] font-bold leading-tight opacity-70 italic text-center uppercase">Unauthorized opening is a federal offense.<br/>Diplomatic Secure Transit Protocol Restricted.</span>
                        </div>
                    </div>
                </div>

                {/* Footer Section */}
                <div className="flex mt-auto justify-between items-end w-full pb-10">
                    <div className="flex flex-col gap-6 ml-10">
                         <div className="flex gap-[6px]">
                            {[15, 8, 22, 6, 18, 10, 28, 6, 15, 8, 12, 18, 24, 6, 15, 8, 18, 6].map((h, i) => (
                                <div key={i} className="w-[6px] bg-[#2d2d2d]" style={{ height: `${h * 4}px` }} />
                            ))}
                        </div>
                        <div className="border-t-2 border-black w-[400px] pt-4">
                             <span className="text-3xl font-serif italic text-[#00008b] font-black leading-none">
                                {data.senderName} 
                             </span>
                             <p className="text-sm uppercase font-black opacity-40 mt-1">Authorized Dispatcher Signature</p>
                        </div>
                    </div>

                    {/* Authorized Stamp */}
                    <div className="relative mr-40">
                        <div 
                            className="border-[6px] border-[#8b0000]/40 rounded-full w-[240px] h-[240px] flex items-center justify-center font-black text-[#8b0000]/40 text-[22px] text-center leading-tight uppercase p-6"
                            style={{ transform: `rotate(${rotateStamp}deg)` }}
                        >
                            <div className="border-2 border-dashed border-[#8b0000]/30 rounded-full w-full h-full flex items-center justify-center">
                                VERIFIED<br/>& SECURED<br/>DIPLOMATIC
                            </div>
                        </div>
                    </div>
                    
                    {/* QR and Security */}
                    <div className="flex flex-col items-end gap-3 pr-10 pb-4">
                         <div className="w-28 h-28 border-2 border-black p-1 bg-white flex flex-wrap gap-[2px]">
                                {[...Array(36)].map((_, i) => (
                                    <div key={i} className={`w-[14px] h-[14px] ${Math.random() > 0.5 ? 'bg-black' : 'bg-transparent'}`} />
                                ))}
                         </div>
                        <span className="text-[12px] text-black/50 font-sans uppercase font-black tracking-widest">
                            WTB-V1.2-SECURE-{data.trackingId.substring(0,10)}
                        </span>
                    </div>
                </div>
            </div>
            
            {/* Corner Foil */}
            <div className="absolute top-10 right-10 w-48 h-48 border-8 border-[#d4af37]/20 rounded-full -rotate-12 flex items-center justify-center text-[#d4af37]/40 font-black text-center text-base p-4">
                 DIPLOMATIC<br/>CARGO<br/>TRANSIT
            </div>
        </div>
    );
});

ReceiptPreview.displayName = 'ReceiptPreview';
