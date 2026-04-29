"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { getApiUrl } from '@/lib/utils';
import { useRouter, usePathname } from 'next/navigation';

type AppUser = {
    email: string;
    company_name: string;
    plan_type: string;
};

type MultiTenantContextType = {
    user: AppUser | null;
    companyId: string | null;
    loading: boolean;
    refreshAuth: () => Promise<void>;
};

const MultiTenantContext = createContext<MultiTenantContextType>({
    user: null,
    companyId: null,
    loading: true,
    refreshAuth: async () => {},
});

export const useMultiTenant = () => useContext(MultiTenantContext);

export default function MultiTenantProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<AppUser | null>(null);
    const [companyId, setCompanyId] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();
    const pathname = usePathname();

    const initializeAuth = async () => {
        try {
            const res = await fetch(`${getApiUrl()}/api/auth/me`, {
                credentials: 'include'
            });
            
            if (res.ok) {
                const data = await res.json();
                setUser({
                    email: data.email,
                    company_name: data.company_name,
                    plan_type: data.plan_type
                });
                setCompanyId(data.company_id);
            } else {
                setUser(null);
                setCompanyId(null);
                // Redirect if unauthenticated on protected routes
                if (res.status === 401 && (pathname.startsWith('/dashboard') || pathname.startsWith('/track') || pathname.startsWith('/admin'))) {
                    router.push('/auth');
                }
            }
        } catch (err) {
            console.error("Auth check failed:", err);
            setUser(null);
            setCompanyId(null);
        } finally {
            setLoading(false);
        }
    };

    const refreshAuth = async () => {
        setLoading(true);
        await initializeAuth();
    };

    useEffect(() => {
        initializeAuth();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Re-validate auth when user returns to the tab (prevents stale JWT sessions)
    useEffect(() => {
        const handleVisibility = () => {
            if (document.visibilityState === 'visible') {
                initializeAuth();
            }
        };
        document.addEventListener('visibilitychange', handleVisibility);
        return () => document.removeEventListener('visibilitychange', handleVisibility);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
        <MultiTenantContext.Provider value={{ user, companyId, loading, refreshAuth }}>
            {children}
        </MultiTenantContext.Provider>
    );
}
