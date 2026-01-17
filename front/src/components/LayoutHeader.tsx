"use client";

import { usePathname } from 'next/navigation';
import { Header } from './Header';
import { Logo } from './Logo';
import { LanguageToggle } from './LanguageToggle';
import { ThemeToggle } from './ThemeToggle';
import { signOut } from 'next-auth/react';

export function LayoutHeader() {
    const pathname = usePathname();
    const isAdmin = pathname?.startsWith('/admin');

    if (isAdmin) {
        return (
            <header className="fixed top-0 left-0 right-0 z-50 bg-bg/80 backdrop-blur-xl border-b border-border py-4">
                <div className="container-wide flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <Logo />
                    <div className="flex items-center gap-4">
                        <LanguageToggle />
                        <ThemeToggle />
                        <div className="h-6 w-px bg-border mx-2" />
                        <button
                            onClick={() => signOut()}
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
