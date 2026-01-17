import React from 'react';
import { Zap, Cpu, FileText, XCircle, PlusCircle, ChevronLeft } from 'lucide-react';

interface ManifestEditorProps {
    emailText: string;
    setEmailText: (text: string) => void;
    onGenerate: () => void;
    loading: boolean;
    error: string | null;
    dict: any;
    onAbort: () => void;
}

export const ManifestEditor: React.FC<ManifestEditorProps> = ({
    emailText,
    setEmailText,
    onGenerate,
    loading,
    error,
    dict,
    onAbort
}) => {
    return (
        <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-8 animate-fade-in pb-20">
            <div className="lg:col-span-2 space-y-6">
                <div className="flex items-center gap-3 mb-2">
                    <div className="bg-accent/10 p-2 rounded-xl">
                        <Zap className="text-accent" size={20} />
                    </div>
                    <div>
                        <h2 className="text-xl font-black text-text-main uppercase tracking-tight">Intelligence Intake</h2>
                        <p className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] opacity-60">Telemetry Data Uplink Port</p>
                    </div>
                </div>

                <div className="relative group">
                    <div className="absolute -inset-1 bg-linear-to-r from-accent/20 to-primary/20 rounded-3xl blur opacity-25 group-focus-within:opacity-100 transition duration-1000 group-focus-within:duration-200" />
                    <div className="relative">
                        <div className="absolute top-4 left-4 flex gap-1.5 pointer-events-none z-10">
                            <div className="w-2 h-2 rounded-full bg-error/40" />
                            <div className="w-2 h-2 rounded-full bg-warning/40" />
                            <div className="w-2 h-2 rounded-full bg-success/40" />
                        </div>
                        <textarea
                            className="w-full h-[400px] bg-bg/60 backdrop-blur-md text-text-main p-12 pt-16 rounded-3xl border border-border group-focus-within:border-accent/50 outline-none resize-none font-mono text-sm transition-all shadow-2xl tracking-tight leading-relaxed"
                            placeholder={dict.admin.placeholder}
                            value={emailText}
                            onChange={(e) => setEmailText(e.target.value)}
                        />
                        <div className="absolute bottom-6 right-8 flex items-center gap-4 text-[10px] font-black text-text-muted uppercase tracking-widest pointer-events-none opacity-40">
                            <span>Parser v2.4.0</span>
                            <div className="w-1 h-1 bg-border rounded-full" />
                            <span>{emailText.length} Chars</span>
                        </div>
                    </div>
                </div>

                {error && (
                    <div className="p-4 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-3 text-error text-sm animate-fade-in font-bold">
                        <XCircle size={18} />
                        {error}
                    </div>
                )}

                <button
                    disabled={loading || !emailText.trim()}
                    onClick={onGenerate}
                    className="btn-primary w-full py-5 text-lg flex items-center justify-center gap-3 group relative overflow-hidden active:scale-[0.98] disabled:opacity-50 disabled:active:scale-100"
                >
                    <div className="absolute inset-0 bg-linear-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-1000" />
                    {loading ? (
                        <>
                            <div className="w-5 h-5 border-3 border-white/30 border-t-white rounded-full animate-spin" />
                            <span>Initializing Stream...</span>
                        </>
                    ) : (
                        <>
                            <Zap size={22} className="fill-white" />
                            <span>Initiate Manifest Deployment</span>
                        </>
                    )}
                </button>
            </div>

            <div className="space-y-6">
                <div className="glass-panel p-6 border-accent/20">
                    <div className="flex items-center gap-3 mb-6">
                        <div className="bg-accent/10 p-2 rounded-lg">
                            <Cpu className="text-accent" size={18} />
                        </div>
                        <h3 className="font-black text-xs uppercase tracking-widest text-text-main">Neural Templates</h3>
                    </div>

                    <div className="space-y-3">
                        {[
                            { name: 'Trans-Atlantic', country: 'USA ➔ UK', content: `From: NY Logistics\nSender Country: USA\nTo: John Watson\nAddress: 221B Baker St, London\nCountry: UK\nContact: +44 20 7946 0000` },
                            { name: 'Euro Link', country: 'Germany ➔ France', content: `From: Berlin Express\nSender Country: Germany\nTo: Jean-Luc\nAddress: 42 Rue de Rivoli, Paris\nCountry: France\nContact: +33 1 42 77 00 00` },
                            { name: 'Pacific Rim', country: 'Japan ➔ Australia', content: `From: Tokyo Port\nSender Country: Japan\nTo: Kenji Sato\nAddress: 123 George St, Sydney\nCountry: Australia\nContact: +61 2 9255 1777` }
                        ].map((template) => (
                            <button
                                key={template.name}
                                onClick={() => setEmailText(template.content)}
                                className="w-full p-4 bg-surface-muted/50 hover:bg-accent/10 border border-border hover:border-accent/40 rounded-2xl text-left transition-all group"
                            >
                                <div className="flex justify-between items-start mb-1">
                                    <span className="text-[10px] font-black text-accent uppercase tracking-widest">{template.name}</span>
                                    <PlusCircle size={14} className="text-text-muted opacity-0 group-hover:opacity-100 transition-opacity" />
                                </div>
                                <p className="text-xs font-bold text-text-main opacity-80">{template.country}</p>
                            </button>
                        ))}
                    </div>
                </div>

                <div className="glass-panel p-6 bg-accent/5">
                    <div className="flex items-center gap-3 mb-6">
                        <div className="bg-accent/10 p-2 rounded-lg">
                            <FileText className="text-accent" size={18} />
                        </div>
                        <h3 className="font-black text-xs uppercase tracking-widest text-text-main">Syntax Guide</h3>
                    </div>

                    <div className="space-y-4">
                        <div className="p-3 bg-bg/50 rounded-xl border border-border/50">
                            <code className="text-[10px] font-mono text-accent">Sender Country: [Name]</code>
                            <p className="text-[9px] text-text-muted mt-1 uppercase font-bold tracking-tighter opacity-70">Defines the origin node</p>
                        </div>
                        <div className="p-3 bg-bg/50 rounded-xl border border-border/50">
                            <code className="text-[10px] font-mono text-accent">Address: [Full Line]</code>
                            <p className="text-[9px] text-text-muted mt-1 uppercase font-bold tracking-tighter opacity-70">Geographic target coordinates</p>
                        </div>
                        <p className="text-[10px] text-text-muted leading-relaxed font-medium pt-2">
                            Our engine automatically sanitizes and validates manifest data through the Global Tracking Network.
                        </p>
                    </div>
                </div>

                <button
                    onClick={onAbort}
                    className="w-full flex items-center justify-center gap-2 text-text-muted hover:text-accent py-2 transition-colors text-xs font-black uppercase tracking-widest"
                >
                    <ChevronLeft size={16} />
                    Abort and View Manifests
                </button>
            </div>
        </div>
    );
};
