"use client";

import React, { useState } from 'react';
import { Search, Loader2 } from 'lucide-react';
import { useI18n } from './I18nContext';

interface TrackingSearchProps {
    onSearch: (trackingNumber: string) => Promise<void>;
    isLoading: boolean;
}

export const TrackingSearch: React.FC<TrackingSearchProps> = ({ onSearch, isLoading }) => {
    const { dict } = useI18n();
    const [input, setInput] = useState('');
    const [loadingMsgIndex, setLoadingMsgIndex] = useState(0);

    React.useEffect(() => {
        let interval: NodeJS.Timeout;
        if (isLoading) {
            interval = setInterval(() => {
                setLoadingMsgIndex(prev => (prev + 1) % (dict.hero.loadingMessages?.length || 1));
            }, 1200);
        } else {
            setLoadingMsgIndex(0);
        }
        return () => clearInterval(interval);
    }, [isLoading, dict.hero.loadingMessages]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            await onSearch(input.trim());
        }
    };

    return (
        <div className="w-full max-w-2xl mx-auto p-1 bg-linear-to-br from-border/50 via-accent/5 to-border/50 rounded-[2.5rem] shadow-3xl group/container">
            <div className="glass-panel p-10 md:p-14 relative overflow-hidden rounded-[2.2rem]">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent/5 rounded-full blur-3xl -mr-16 -mt-16 pointer-events-none" />

                <h1 className="text-4xl md:text-6xl font-black mb-6 text-center text-gradient leading-[0.9] tracking-tighter uppercase">
                    {dict.hero.title}
                </h1>

                {isLoading ? (
                    <div className="h-[92px] flex flex-col items-center justify-center animate-pulse">
                        <div className="flex items-center gap-3 text-accent mb-2">
                            <Loader2 className="animate-spin" size={20} />
                            <span className="text-[10px] font-black uppercase tracking-[0.4em]">{dict.common.loading}</span>
                        </div>
                        <p className="text-text-muted text-sm font-bold tracking-tight uppercase opacity-60">
                            {dict.hero.loadingMessages?.[loadingMsgIndex]}
                        </p>
                    </div>
                ) : (
                    <p className="text-center text-text-muted mb-12 text-lg md:text-xl font-bold max-w-md mx-auto leading-relaxed border-l-2 border-accent/20 pl-6 h-[92px] flex items-center">
                        {dict.hero.subtitle}
                    </p>
                )}

                <form onSubmit={handleSubmit} className="relative group/form mt-4">
                    <div className="absolute -inset-1 bg-linear-to-r from-accent to-accent-deep rounded-3xl blur opacity-0 group-focus-within/form:opacity-20 transition-opacity duration-500" />
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder={dict.hero.placeholder}
                        className="relative z-10 w-full bg-surface-muted text-text-main py-6 pl-10 pr-44 rounded-3xl border-2 border-transparent outline-none focus:border-accent/30 focus:bg-surface focus:ring-8 focus:ring-accent/5 transition-all placeholder:text-text-muted/40 font-black tracking-widest text-xl shadow-inner uppercase"
                        disabled={isLoading}
                    />
                    <button
                        type="submit"
                        disabled={isLoading || !input.trim()}
                        className="absolute right-3 top-3 bottom-3 z-20 btn-primary py-0! px-10! flex items-center justify-center gap-3 transition-all hover:scale-[1.03] active:scale-95 shadow-2xl shadow-accent/30 disabled:grayscale disabled:opacity-50"
                    >
                        {isLoading ? <Loader2 className="animate-spin" size={20} /> : (
                            <>
                                <Search size={20} strokeWidth={3} />
                                <span className="hidden sm:inline font-black uppercase tracking-[0.2em] text-xs">{dict.hero.track}</span>
                            </>
                        )}
                    </button>
                </form>
            </div>
        </div>
    );
};
