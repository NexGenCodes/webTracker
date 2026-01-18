"use client";

import React, { memo, useState, useRef, useEffect } from "react";
import { useI18n } from "./I18nContext";
import { Locale } from "../lib/dictionaries";
import { Globe, ChevronDown, Check } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";

const languages: { code: Locale; label: string }[] = [
    { code: 'en', label: 'English' },
    { code: 'fr', label: 'Français' },
    { code: 'es', label: 'Español' },
    { code: 'de', label: 'Deutsch' },
    { code: 'zh', label: '中文' },
    { code: 'pt', label: 'Português' },
    { code: 'it', label: 'Italiano' },
    { code: 'ar', label: 'العربية' },
];

export const LanguageToggle = memo(() => {
    const { locale, setLocale } = useI18n();
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    const currentLanguage = languages.find(l => l.code === locale) || languages[0];

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    return (
        <div className="relative" ref={containerRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 px-3 py-2 rounded-xl hover:bg-surface-muted transition-all duration-300 group"
            >
                <div className="bg-accent/10 p-1.5 rounded-lg group-hover:bg-accent group-hover:text-white transition-colors duration-300">
                    <Globe className="w-4 h-4 text-accent group-hover:text-white" />
                </div>
                <div className="flex flex-col items-start leading-none gap-1">
                    <span className="text-[10px] font-black uppercase tracking-widest text-text-muted opacity-60">Language</span>
                    <div className="flex items-center gap-1">
                        <span className="text-sm font-black uppercase tracking-tight">{currentLanguage.label}</span>
                        <ChevronDown className={cn("w-3 h-3 transition-transform duration-300", isOpen && "rotate-180")} />
                    </div>
                </div>
            </button>

            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                        animate={{ opacity: 1, y: 0, scale: 1 }}
                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                        transition={{ duration: 0.2, ease: "easeOut" }}
                        className="absolute right-0 mt-2 w-48 bg-surface/90 border border-border rounded-2xl shadow-2xl z-100 overflow-hidden backdrop-blur-xl"
                    >
                        <div className="py-2 grid grid-cols-1 gap-1 px-2">
                            {languages.map((lang) => (
                                <button
                                    key={lang.code}
                                    onClick={() => {
                                        setLocale(lang.code);
                                        setIsOpen(false);
                                    }}
                                    className={cn(
                                        "flex items-center justify-between px-3 py-2 rounded-xl text-sm font-bold transition-all duration-200 group/item",
                                        locale === lang.code
                                            ? "bg-accent text-white"
                                            : "text-text-main hover:bg-surface-muted"
                                    )}
                                >
                                    <span>{lang.label}</span>
                                    {locale === lang.code && <Check className="w-4 h-4" />}
                                </button>
                            ))}
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
});
