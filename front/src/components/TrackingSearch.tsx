"use client";

import React, { useState, useCallback, memo } from 'react';
import { Search, Loader2 } from 'lucide-react';
import { useI18n } from './I18nContext';
import { isValidTrackingNumber } from '@/lib/constants';
import { toast } from 'react-hot-toast';

interface TrackingSearchProps {
    onSearch: (trackingNumber: string) => Promise<void>;
    isLoading: boolean;
}

export const TrackingSearch: React.FC<TrackingSearchProps> = memo(({ onSearch, isLoading }) => {
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

    const handleSubmit = useCallback(async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = input.trim().toUpperCase();
        if (!trimmed) return;

        if (!isValidTrackingNumber(trimmed)) {
            toast.error("Invalid tracking format. Please check your ID.", {
                icon: '🎫',
            });
            return;
        }

        await onSearch(trimmed);
    }, [input, onSearch]);

    const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        setInput(e.target.value);
    }, []);

    return (
        <div className="w-full max-w-2xl mx-auto p-1 bg-linear-to-br from-border/50 via-accent/5 to-border/50 rounded-[2.5rem] shadow-3xl group/container">
            <div className="glass-panel p-5 md:p-14 relative overflow-hidden rounded-[2.2rem]">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent/5 rounded-full blur-3xl -mr-16 -mt-16 pointer-events-none" />

                <h1 className="text-3xl md:text-6xl font-black mb-6 text-center text-gradient leading-[0.9] tracking-tighter uppercase">
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
                    <p className="text-center text-text-muted mb-8 md:mb-12 text-base md:text-xl font-bold max-w-md mx-auto leading-relaxed border-l-2 border-accent/20 pl-6 h-[92px] flex items-center">
                        {dict.hero.subtitle}
                    </p>
                )}

                <form onSubmit={handleSubmit} className="relative group/form mt-4">
                    <div className="absolute -inset-1 bg-linear-to-r from-accent to-accent-deep rounded-3xl blur opacity-0 group-focus-within/form:opacity-20 transition-opacity duration-500" />
                    <input
                        type="text"
                        value={input}
                        onChange={handleInputChange}
                        placeholder={dict.hero.placeholder}
                        className="relative z-10 w-full bg-surface-muted text-text-main py-4 md:py-6 pl-6 md:pl-10 pr-16 sm:pr-44 rounded-3xl border-2 border-transparent outline-none focus:border-accent/30 focus:bg-surface focus:ring-8 focus:ring-accent/5 transition-all placeholder:text-text-muted/40 font-black tracking-widest text-lg md:text-xl shadow-inner uppercase"
                        disabled={isLoading}
                    />
                    <button
                        type="submit"
                        disabled={isLoading || !input.trim()}
                        className="absolute right-2 md:right-3 top-2 md:top-3 bottom-2 md:bottom-3 z-20 btn-primary py-0! px-4! sm:px-10! flex items-center justify-center gap-3 transition-all hover:scale-[1.03] active:scale-95 shadow-2xl shadow-accent/30 disabled:grayscale disabled:opacity-50"
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
});
TrackingSearch.displayName = 'TrackingSearch';
