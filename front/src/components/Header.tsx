"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Logo } from './Logo';
import { LanguageToggle } from './LanguageToggle';
import { ThemeToggle } from './ThemeToggle';
import { useI18n } from './I18nContext';
import { cn } from '@/lib/utils';
import { useEffect, useState } from 'react';

interface HeaderProps {
    showNav?: boolean;
    className?: string;
}

export const Header: React.FC<HeaderProps> = ({ showNav = true, className }) => {
    const { dict } = useI18n();
    const pathname = usePathname();
    const [scrolled, setScrolled] = useState(false);

    useEffect(() => {
        const handleScroll = () => setScrolled(window.scrollY > 20);
        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    const navLinks = [
        { href: '/', label: dict.common.home },
        { href: '/about', label: dict.common.about },
        { href: '/contact', label: dict.common.contact },
    ];

    return (
        <header className={cn(
            "fixed top-0 left-0 right-0 z-50 transition-all duration-500",
            scrolled ? "py-4 bg-bg/80 backdrop-blur-xl border-b border-border shadow-sm" : "py-8 bg-transparent",
            className
        )}>
            <div className="container-wide flex justify-between items-center">
                <Logo />

                <div className="flex items-center gap-10">
                    {showNav && (
                        <nav className="hidden lg:flex items-center gap-12 text-xs font-black uppercase tracking-[0.2em]">
                            {navLinks.map((link) => {
                                const isActive = pathname === link.href;
                                return (
                                    <Link
                                        key={link.href}
                                        href={link.href}
                                        className={cn(
                                            "transition-all duration-300 relative group",
                                            isActive
                                                ? "text-accent"
                                                : "text-text-muted hover:text-accent"
                                        )}
                                    >
                                        {link.label}
                                        <span className={cn(
                                            "absolute -bottom-2 left-0 h-0.5 bg-accent transition-all duration-300",
                                            isActive ? "w-full" : "w-0 group-hover:w-full"
                                        )} />
                                    </Link>
                                );
                            })}
                        </nav>
                    )}

                    <div className="flex items-center gap-6">
                        <div className="h-6 w-px bg-border hidden lg:block" />
                        <div className="flex items-center gap-3">
                            <LanguageToggle />
                            <ThemeToggle />
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
};
