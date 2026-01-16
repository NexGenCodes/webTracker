"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { dictionaries, Locale, Dictionary } from '../lib/dictionaries';

interface I18nContextProps {
    locale: Locale;
    setLocale: (locale: Locale) => void;
    dict: Dictionary;
}

const I18nContext = createContext<I18nContextProps | undefined>(undefined);

export const I18nProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [locale, setLocaleState] = useState<Locale>('en');

    useEffect(() => {
        const saved = localStorage.getItem('locale') as Locale;
        if (saved && dictionaries[saved]) {
            setLocaleState(saved);
        }
    }, []);

    const setLocale = React.useCallback((newLocale: Locale) => {
        setLocaleState(newLocale);
        localStorage.setItem('locale', newLocale);
    }, []);

    const contextValue = React.useMemo(() => ({
        locale,
        setLocale,
        dict: dictionaries[locale]
    }), [locale, setLocale]);

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
