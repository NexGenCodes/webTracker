"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { checkAuthAction } from '@/app/actions/auth';
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
    loading: false,
    refreshAuth: async () => {},
});

export const useMultiTenant = () => useContext(MultiTenantContext);

export default function MultiTenantProvider({ 
    children, 
    initialUser = null, 
    initialCompanyId = null 
}: { 
    children: React.ReactNode;
    initialUser?: AppUser | null;
    initialCompanyId?: string | null;
}) {
    const [user, setUser] = useState<AppUser | null>(initialUser);
    const [companyId, setCompanyId] = useState<string | null>(initialCompanyId);
    const [loading, setLoading] = useState(false);
    const router = useRouter();
    const pathname = usePathname();

    const initializeAuth = async () => {
        try {
            const { user } = await checkAuthAction();
            
            if (user) {
                setUser({
                    email: user.email,
                    company_name: user.company_name,
                    plan_type: user.plan_type
                });
                setCompanyId(user.company_id || null);
            } else {
                setUser(null);
                setCompanyId(null);
                if (pathname.startsWith('/dashboard') || pathname.startsWith('/track') || pathname.startsWith('/admin')) {
                    router.push('/auth');
                }
            }
        } catch (err) {
            console.error("Auth check failed:", err);
            setUser(null);
            setCompanyId(null);
            if (pathname.startsWith('/dashboard') || pathname.startsWith('/track') || pathname.startsWith('/admin')) {
                router.push('/auth');
            }
        } finally {
            setLoading(false);
        }
    };

    const refreshAuth = async () => {
        setLoading(true);
        await initializeAuth();
    };

    // We no longer fetch on mount because the Server Component pre-fills the data.
    // We only listen for visibility changes to catch if the user logged out in another tab.
    useEffect(() => {
        const handleVisibility = () => {
            if (document.visibilityState === 'visible') {
                checkAuthAction().then(({ user }) => {
                    if (!user && (pathname.startsWith('/dashboard') || pathname.startsWith('/track') || pathname.startsWith('/admin'))) {
                        router.push('/auth');
                    }
                }).catch(() => {});
            }
        };
        document.addEventListener('visibilitychange', handleVisibility);
        return () => document.removeEventListener('visibilitychange', handleVisibility);
    }, [pathname, router]);

    return (
        <MultiTenantContext.Provider value={{ user, companyId, loading, refreshAuth }}>
            {children}
        </MultiTenantContext.Provider>
    );
}
