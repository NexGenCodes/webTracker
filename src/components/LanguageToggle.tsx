"use client";

import React, { memo } from "react";
import { useI18n } from "./I18nContext";
import { Locale } from "../lib/dictionaries";
import { Globe } from "lucide-react";

const languages: { code: Locale; label: string }[] = [
    { code: 'en', label: 'English' },
    { code: 'fr', label: 'Français' },
    { code: 'es', label: 'Español' },
    { code: 'de', label: 'Deutsch' },
    { code: 'zh', label: '中文' },
];

export const LanguageToggle = memo(() => {
    const { locale, setLocale } = useI18n();

    return (
        <div className="flex items-center gap-2">
            <Globe className="w-4 h-4 text-gray-500" />
            <select
                value={locale}
                onChange={(e) => setLocale(e.target.value as Locale)}
                className="bg-transparent text-sm font-bold text-text-main focus:outline-none cursor-pointer hover:text-accent transition-colors"
            >
                {languages.map((lang) => (
                    <option key={lang.code} value={lang.code} className="bg-surface text-text-main">
                        {lang.label}
                    </option>
                ))}
            </select>
        </div>
    );
});
