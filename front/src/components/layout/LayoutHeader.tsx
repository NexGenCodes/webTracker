"use client";

import { usePathname } from 'next/navigation';
import { Header } from './Header';
import { Logo } from './Logo';
import { LanguageToggle } from '@/components/shared/LanguageToggle';
import { ThemeToggle } from '@/components/shared/ThemeToggle';
import { getApiUrl } from '@/lib/utils';

import { useRouter } from 'next/navigation';

export function LayoutHeader() {
    const pathname = usePathname();
    const isDashboard = pathname?.startsWith('/dashboard') || pathname?.startsWith('/super-admin');
    const isAuth = pathname?.startsWith('/auth');
    const router = useRouter();

    const handleSignOut = async () => {
        await fetch(`${getApiUrl()}/api/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        router.push('/auth');
    };

    if (isAuth) {
        return null;
    }

    if (isDashboard) {
        return (
            <header className="sticky top-0 z-[1000] bg-bg/80 backdrop-blur-xl border-b border-border py-4">
                <div className="container-wide flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <Logo href="/dashboard" />
                    <div className="flex items-center gap-4">
                        <LanguageToggle />
                        <ThemeToggle />
                        <div className="h-6 w-px bg-border mx-2" />
                        <button
                            onClick={handleSignOut}
                            className="bg-error/10 hover:bg-error text-error hover:text-white px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all"
                        >
                            Logout
                        </button>
                    </div>
                </div>
            </header>
        );
    }

    return <Header />;
}
