"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { dictionaries, Locale, Dictionary } from '../lib/dictionaries';

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

    useEffect(() => {
        const saved = localStorage.getItem('locale') as Locale;
        if (saved && dictionaries[saved]) {
            setLocaleState(saved);
        }
    }, []);

    // Apply RTL direction to document
    useEffect(() => {
        const isRTL = RTL_LOCALES.includes(locale);
        document.documentElement.dir = isRTL ? 'rtl' : 'ltr';
        document.documentElement.lang = locale;
    }, [locale]);

    const setLocale = React.useCallback((newLocale: Locale) => {
        setLocaleState(newLocale);
        localStorage.setItem('locale', newLocale);
    }, []);

    const isRTL = RTL_LOCALES.includes(locale);

    const contextValue = React.useMemo(() => ({
        locale,
        setLocale,
        dict: dictionaries[locale],
        isRTL
    }), [locale, setLocale, isRTL]);

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
