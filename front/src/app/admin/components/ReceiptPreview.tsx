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
            {/* Background Texture Overlay (Noise) */}
            <div 
                className="absolute inset-0 pointer-events-none z-0" 
                style={{ 
                    opacity: noiseOpacity,
                    backgroundColor: '#000',
                    mixBlendMode: 'overlay',
                }} 
            />

            {/* Guilloche Patterns (Subtle Waves) */}
            <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-[0.08]" xmlns="http://www.w3.org/2000/svg">
                <defs>
                    <pattern id="guilloche" x="0" y="0" width="1800" height="90" patternUnits="userSpaceOnUse">
                        <path 
                            d="M 40 45 Q 70 33, 100 45 T 160 45 T 220 45 T 280 45 T 340 45 T 400 45 T 460 45 T 520 45 T 580 45 T 640 45 T 700 45 T 760 45 T 820 45 T 880 45 T 940 45 T 1000 45 T 1060 45 T 1120 45 T 1180 45 T 1240 45 T 1300 45 T 1360 45 T 1420 45 T 1480 45 T 1540 45 T 1600 45 T 1660 45 T 1720 45 T 1760 45" 
                            fill="none" 
                            stroke="#8b0000" 
                            strokeWidth="0.8"
                        />
                    </pattern>
                </defs>
                <rect x="0" y="450" width="1800" height="600" fill="url(#guilloche)" />
            </svg>

            {/* Fold Lines */}
            <div className="absolute top-[466px] left-0 w-full h-[1px] bg-black/5 z-0" />
            <div className="absolute top-[933px] left-0 w-full h-[1px] bg-black/5 z-0" />

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
                    <div className="w-[800px] h-[45px] bg-[#8b0000] flex items-center justify-center mt-6 mb-4 shadow-sm">
                        <span className="text-white text-[20px] font-black tracking-[0.4em] uppercase">
                            INTERNATIONAL SPECIAL DELIVERY SERVICE
                        </span>
                    </div>
                    <div className="flex justify-between w-full px-20">
                         <span className="text-[26px] font-black tracking-[0.3em] font-mono opacity-80">
                            {data.trackingId}
                        </span>
                        <span className="text-[42px] font-black text-[#cc0000] font-mono">
                            № 00{data.trackingId}
                        </span>
                    </div>
                </div>

                {/* Main Information Grid (gX=20, gY=420, gW=Width-40, gH=624) */}
                <div className="flex flex-wrap border-[3px] border-black w-full bg-white/40 backdrop-blur-[1px]">
                    {/* Destination/Origin Row */}
                    <div className="flex w-full border-b-[3px] border-black h-[104px]">
                        <div className="w-[30%] p-6 border-r-[3px] border-black flex flex-col justify-center">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60">DESTINATION</span>
                            <span className="text-[34px] text-[#2d2d2d] font-black uppercase leading-none mt-1">{(data.receiverCountry || 'TBD')}</span>
                        </div>
                        <div className="w-[22%] border-r-[3px] border-black flex flex-col">
                            <div className="h-[40px] bg-black text-white flex items-center justify-center font-black text-[14px] tracking-widest">SERVICE</div>
                            <div className="flex-1 flex items-center gap-4 px-6">
                                <span className="text-[16px] font-black opacity-40">DIPLOMATIC</span>
                                <div className="w-10 h-10 border-2 border-black flex items-center justify-center font-black text-2xl">X</div>
                            </div>
                        </div>
                        <div className="w-[22%] border-r-[3px] border-black flex flex-col">
                            <div className="h-[40px] bg-black text-white flex items-center justify-center font-black text-[14px] tracking-widest">PAYMENT</div>
                            <div className="flex-1 flex items-center gap-4 px-6">
                                <span className="text-[16px] font-black opacity-40">ACCOUNT</span>
                                <div className="w-10 h-10 border-2 border-black flex items-center justify-center font-black text-2xl">X</div>
                            </div>
                        </div>
                        <div className="w-[26%] p-6 flex flex-col justify-center">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60">ORIGIN</span>
                            <span className="text-[34px] text-[#2d2d2d] font-black uppercase leading-none mt-1">{(data.senderCountry || 'TBD')}</span>
                        </div>
                    </div>
                    
                    {/* Participants Row */}
                    <div className="flex w-full border-b-[3px] border-black h-[208px]">
                        <div className="w-[52%] p-8 border-r-[3px] border-black flex flex-col">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60 mb-2">CONSIGNEE (RECEIVER DETAILS)</span>
                            <span className="text-[44px] text-[#2d2d2d] font-black uppercase mb-1 drop-shadow-sm">{(data.receiverName || '---')}</span>
                            <span className="text-[22px] font-bold text-black/70 italic leading-tight">{(data.receiverAddress || 'ADDR: N/A')}</span>
                        </div>
                        <div className="w-[48%] p-8 flex flex-col">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60 mb-2">CONSIGNOR (SENDER)</span>
                            <span className="text-[44px] text-[#2d2d2d] font-black uppercase drop-shadow-sm">{(data.senderName || '---')}</span>
                            <span className="text-[22px] font-bold text-black/50 mt-3">SENDER ID: {data.trackingId.substring(0,8)}</span>
                        </div>
                    </div>

                    {/* Cargo Specifications Row */}
                    <div className="flex w-full border-b-[3px] border-black h-[104px]">
                        <div className="w-[30%] p-6 border-r-[3px] border-black flex flex-col justify-center">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60">DESCRIPTION</span>
                            <span className="text-[30px] text-[#2d2d2d] font-black uppercase mt-1">{(data.cargoType || 'CONSIGNMENT')}</span>
                        </div>
                        <div className="w-[22%] p-6 border-r-[3px] border-black flex flex-col justify-center bg-accent/5">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60">GROSS WT</span>
                            <span className="text-[48px] text-[#8b0000] font-black leading-none mt-1">{(data.weight || '0.00')} <span className="text-xl">KGS</span></span>
                        </div>
                        <div className="w-[22%] p-6 border-r-[3px] border-black flex flex-col justify-center">
                            <span className="text-[14px] text-[#464646] font-black tracking-widest uppercase opacity-60">DEP DATE</span>
                            <span className="text-[30px] text-[#2d2d2d] font-black uppercase mt-1">{today}</span>
                        </div>
                        <div className="w-[26%] flex flex-col items-center justify-center p-4">
                             <div className="bg-[#cc0000] text-white p-2 w-full text-center text-[11px] font-black tracking-[0.2em] mb-2 shadow-sm">! CONFIDENTIAL !</div>
                             <span className="text-[11px] font-black leading-tight opacity-50 italic text-center uppercase tracking-tighter">
                                Unauthorized opening is a federal offense.<br/>
                                Anti-tamper seal protected.
                             </span>
                        </div>
                    </div>
                </div>

                {/* Footer Section */}
                <div className="flex mt-auto justify-between items-end w-full pb-10">
                    <div className="flex flex-col gap-6 ml-10">
                         <div className="flex gap-[6px] items-end">
                            {[15, 8, 22, 12, 18, 10, 28, 6, 15, 8, 12, 22, 24, 6, 15, 8, 18, 6].map((h, i) => (
                                <div key={i} className="w-[7px] bg-[#2d2d2d]" style={{ height: `${h * 4.5}px` }} />
                            ))}
                        </div>
                        <div className="border-t-[3px] border-black w-[450px] pt-4 relative">
                             {/* Signature Path Simulation (Quadratic-like) */}
                             <svg className="absolute -top-12 left-10 w-[300px] h-[80px] pointer-events-none opacity-80" viewBox="0 0 300 80">
                                <path 
                                    d="M 10 60 Q 50 20, 100 50 T 200 40 T 280 60" 
                                    fill="none" 
                                    stroke="#00008b" 
                                    strokeWidth="4"
                                    strokeLinecap="round"
                                />
                             </svg>
                             <span className="text-[38px] font-serif italic text-[#00008b] font-black leading-none ml-4">
                                {data.senderName} 
                             </span>
                             <p className="text-[14px] uppercase font-black opacity-30 mt-2 tracking-widest">Authorized Dispatcher Header</p>
                        </div>
                    </div>

                    {/* Authorized Stamp (MATCHING GO FOIL EXACTLY) */}
                    <div className="relative mr-40 mb-4 scale-110">
                        {/* Foil regular polygon (12 sides) */}
                        <div 
                            className="w-[240px] h-[240px] flex items-center justify-center relative"
                            style={{ transform: `rotate(${rotateStamp}deg)` }}
                        >
                            {/* SVG for 12-sided polygon Foil */}
                            <svg className="absolute inset-0 w-full h-full drop-shadow-xl opacity-[0.9]" viewBox="0 0 100 100">
                                <path 
                                    d="M 50 0 L 75 6.7 L 93.3 25 L 100 50 L 93.3 75 L 75 93.3 L 50 100 L 25 93.3 L 6.7 75 L 0 50 L 6.7 25 L 25 6.7 Z" 
                                    fill="#d4af37" 
                                    fillOpacity="0.25"
                                    stroke="#82641e"
                                    strokeWidth="0.5"
                                />
                                {/* Starburst lines inside */}
                                <g transform="translate(50,50)" stroke="#ffffff" strokeWidth="0.3" strokeOpacity="0.5">
                                    {[...Array(24)].map((_, i) => (
                                        <line key={i} x1="0" y1="0" x2="45" y2="0" transform={`rotate(${i * 15})`} />
                                    ))}
                                </g>
                            </svg>

                            <div className="border-[6px] border-[#8b0000]/30 rounded-full w-[200px] h-[200px] flex items-center justify-center font-black text-[#8b0000]/50 text-[20px] text-center leading-tight uppercase p-6 relative">
                                <div className="border-2 border-dashed border-[#8b0000]/20 rounded-full w-full h-full flex items-center justify-center">
                                    VERIFIED<br/>& SECURED<br/>DIPLOMATIC
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    {/* QR and Security */}
                    <div className="flex flex-col items-end gap-3 pr-10 pb-4">
                         <div className="w-28 h-28 border-2 border-black p-1 bg-white flex flex-wrap gap-[1px] shadow-sm">
                                {[...Array(100)].map((_, i) => (
                                    <div key={i} className={`w-[9px] h-[9px] ${Math.random() > 0.4 ? 'bg-black' : 'bg-transparent'}`} />
                                ))}
                         </div>
                        <span className="text-[12px] text-black/40 font-sans uppercase font-black tracking-[0.2em] bg-black/5 px-2 py-1 rounded">
                            WTB-V1.2-[{data.trackingId.substring(0,8).toUpperCase()}]
                        </span>
                    </div>
                </div>
            </div>
            
            {/* Corner Foil Sticker (NEW) */}
            <div className="absolute top-[-40px] right-[-40px] w-64 h-64 border-[12px] border-[#d4af37]/10 rounded-full rotate-45 flex items-center justify-center bg-[#d4af37]/5">
                 <div className="text-[#82641e]/30 font-black text-center text-xl uppercase tracking-widest -rotate-45 mt-10 mr-10">
                    AIR FREIGHT<br/>MANIFEST
                 </div>
            </div>
        </div>
    );
});

ReceiptPreview.displayName = 'ReceiptPreview';
