"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Logo } from './Logo';
import { LanguageToggle } from '@/components/shared/LanguageToggle';
import { ThemeToggle } from '@/components/shared/ThemeToggle';
import { useI18n } from '@/components/providers/I18nContext';
import { cn } from '@/lib/utils';
import { Menu, X, ArrowRight } from 'lucide-react';
import { motion, useScroll, useSpring } from 'framer-motion';
import { useState, useEffect } from 'react';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { MobileNavOverlay } from './MobileNavOverlay';

interface HeaderProps {
    showNav?: boolean;
    className?: string;
}

export const Header: React.FC<HeaderProps> = ({ showNav = true, className }) => {
    const { dict } = useI18n();
    const pathname = usePathname();
    const { user } = useMultiTenant();
    const [scrolled, setScrolled] = useState(false);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    useEffect(() => {
        const handleScroll = () => setScrolled(window.scrollY > 0);
        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    // Close mobile menu when route changes
    useEffect(() => {
        setMobileMenuOpen(false);
    }, [pathname]);

    const { scrollYProgress } = useScroll();
    const scaleX = useSpring(scrollYProgress, {
        stiffness: 100,
        damping: 30,
        restDelta: 0.001
    });

    const navLinks = [
        { href: '/', label: dict.common.home || 'Home' },
        { href: '/track', label: dict.common.track || 'Track' },
        { href: '/pricing', label: dict.common.pricing || 'Pricing' },
        { href: '/about', label: dict.common.about || 'About' },
    ];

    return (
        <>
            <header className={cn(
                "sticky top-0 z-[1000] transition-all duration-200 ease-in-out",
                scrolled || mobileMenuOpen ? "py-4 bg-surface/95 backdrop-blur-md border-b border-border shadow-md" : "py-6 md:py-8 bg-transparent",
                className
            )}>
                {/* Scroll Progress Bar */}
                <motion.div
                    className="absolute bottom-0 left-0 right-0 h-[2px] bg-accent origin-left"
                    style={{ scaleX }}
                />
                <div className="container-wide flex items-center justify-between relative z-100">
                    <Logo href={user ? "/dashboard" : "/"} />

                    <div className="flex items-center gap-4 sm:gap-6">
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

                        <div className="flex items-center gap-4">
                            <LanguageToggle />
                            <ThemeToggle />

                            {user ? (
                                <Link
                                    href="/dashboard"
                                    className="hidden lg:flex items-center gap-2 px-5 py-2.5 bg-accent text-white rounded-xl text-xs font-black uppercase tracking-widest transition-all hover:bg-accent/90 active:scale-95 shadow-lg shadow-accent/20"
                                >
                                    {dict.common?.dashboard || 'Dashboard'}
                                    <ArrowRight size={14} />
                                </Link>
                            ) : (
                                <Link
                                    href="/auth"
                                    className="hidden lg:flex items-center gap-2 px-5 py-2.5 bg-accent text-white rounded-xl text-xs font-black uppercase tracking-widest transition-all hover:bg-accent/90 active:scale-95 shadow-lg shadow-accent/20"
                                >
                                    {dict.auth?.getStarted || 'Get Started'}
                                    <ArrowRight size={14} />
                                </Link>
                            )}
                        </div>

                        {/* Mobile Menu Toggle */}
                        {showNav && (
                            <button
                                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                                className="lg:hidden p-2 text-text-main hover:bg-surface-muted rounded-xl transition-colors relative z-110"
                                aria-label="Toggle menu"
                            >
                                {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
                            </button>
                        )}
                    </div>
                </div>
            </header>

            {/* Mobile Navigation Overlay */}
            <MobileNavOverlay isOpen={mobileMenuOpen} onClose={() => setMobileMenuOpen(false)} zIndex="z-[9999]">
                <nav className="flex flex-col gap-4">
                    {navLinks.map((link, i) => {
                        const isActive = pathname === link.href;
                        return (
                            <motion.div
                                key={link.href}
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ delay: 0.1 + i * 0.05, ease: "easeOut" }}
                            >
                                <Link
                                    href={link.href}
                                    onClick={() => setMobileMenuOpen(false)}
                                    className={cn(
                                        "text-4xl font-black uppercase tracking-tighter py-5 border-b border-border/40 block transition-all active:scale-[0.98]",
                                        isActive ? "text-accent" : "text-text-main hover:text-accent"
                                    )}
                                >
                                    {link.label}
                                </Link>
                            </motion.div>
                        );
                    })}
                </nav>

                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.3 }}
                    className="mt-8"
                >
                    {user ? (
                        <Link
                            href="/dashboard"
                            onClick={() => setMobileMenuOpen(false)}
                            className="flex items-center justify-center gap-3 w-full py-5 bg-accent text-white rounded-2xl font-black uppercase tracking-widest text-base transition-all active:scale-95 shadow-xl shadow-accent/20"
                        >
                            {dict.common?.dashboard || 'Dashboard'}
                            <ArrowRight size={18} />
                        </Link>
                    ) : (
                        <Link
                            href="/auth"
                            onClick={() => setMobileMenuOpen(false)}
                            className="flex items-center justify-center gap-3 w-full py-5 bg-accent text-white rounded-2xl font-black uppercase tracking-widest text-base transition-all active:scale-95 shadow-xl shadow-accent/20"
                        >
                            {dict.auth?.getStarted || 'Get Started'}
                            <ArrowRight size={18} />
                        </Link>
                    )}
                </motion.div>

                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.4 }}
                    className="mt-auto pb-10"
                >
                    <div className="pt-10 border-t border-border/40 text-center">
                        <p className="text-[10px] font-black uppercase tracking-[0.4em] text-accent opacity-60 mb-2">
                            {dict.common.safeLogistics}
                        </p>
                        <p className="text-[10px] font-bold uppercase tracking-widest text-text-muted opacity-30">
                            {dict.common.tagline || ""}
                        </p>
                    </div>
                </motion.div>
            </MobileNavOverlay>
        </>
    );
};
