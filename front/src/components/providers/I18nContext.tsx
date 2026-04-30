"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { loadDictionary, isValidLocale, type Locale, type Dictionary } from '@/lib/dictionaries';

// Import English JSON synchronously as the initial/fallback dictionary
// so the first render is never blank.
import enDict from '@/lib/dictionaries/en.json';

interface I18nContextProps {
    locale: Locale;
    setLocale: (locale: Locale) => void;
    dict: Dictionary;
    isRTL: boolean;
}

const I18nContext = createContext<I18nContextProps | undefined>(undefined);

// RTL languages
const RTL_LOCALES: Locale[] = ['ar'];

export const I18nProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [locale, setLocaleState] = useState<Locale>('en');
    const [dict, setDict] = useState<Dictionary>(enDict as Dictionary);

    // Load dictionary when locale changes
    useEffect(() => {
        let cancelled = false;

        loadDictionary(locale).then((loaded) => {
            if (!cancelled) setDict(loaded);
        });

        return () => { cancelled = true; };
    }, [locale]);

    useEffect(() => {
        const saved = localStorage.getItem('locale');
        if (saved && isValidLocale(saved)) {
            setLocaleState(saved);
        } else if (typeof navigator !== 'undefined') {
            const browserLocale = navigator.language.split('-')[0];
            if (isValidLocale(browserLocale)) {
                setLocaleState(browserLocale);
            }
        }

        // Sync language across tabs/windows
        const handleStorage = (e: StorageEvent) => {
            if (e.key === 'locale' && e.newValue && isValidLocale(e.newValue)) {
                setLocaleState(e.newValue);
            }
        };
        window.addEventListener('storage', handleStorage);
        return () => window.removeEventListener('storage', handleStorage);
    }, []);

    // Apply RTL direction to document
    useEffect(() => {
        const isRTL = RTL_LOCALES.includes(locale);
        document.documentElement.dir = isRTL ? 'rtl' : 'ltr';
        document.documentElement.lang = locale;
    }, [locale]);

    const setLocale = useCallback((newLocale: Locale) => {
        setLocaleState(newLocale);
        localStorage.setItem('locale', newLocale);
    }, []);

    const isRTL = RTL_LOCALES.includes(locale);

    const contextValue = React.useMemo(() => ({
        locale,
        setLocale,
        dict,
        isRTL
    }), [locale, setLocale, dict, isRTL]);

    return (
        <I18nContext.Provider value={contextValue}>
            {children}
        </I18nContext.Provider>
    );
};

export const useI18n = () => {
    const context = useContext(I18nContext);
    if (!context) throw new Error('useI18n must be used within I18nProvider');
    return context;
};
