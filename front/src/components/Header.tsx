"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Logo } from './Logo';
import { LanguageToggle } from './LanguageToggle';
import { ThemeToggle } from './ThemeToggle';
import { useI18n } from './I18nContext';
import { cn } from '@/lib/utils';
import { Menu, X } from 'lucide-react';
import { motion, AnimatePresence, useScroll, useSpring } from 'framer-motion';
import { useState, useEffect } from 'react';

interface HeaderProps {
    showNav?: boolean;
    className?: string;
}

export const Header: React.FC<HeaderProps> = ({ showNav = true, className }) => {
    const { dict } = useI18n();
    const pathname = usePathname();
    const [scrolled, setScrolled] = useState(false);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    useEffect(() => {
        const handleScroll = () => setScrolled(window.scrollY > 20);
        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    // Close mobile menu when route changes
    useEffect(() => {
        setMobileMenuOpen(false);
    }, [pathname]);

    // Lock body scroll when mobile menu is open
    useEffect(() => {
        if (mobileMenuOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = '';
        }
        return () => {
            document.body.style.overflow = '';
        };
    }, [mobileMenuOpen]);

    const { scrollYProgress } = useScroll();
    const scaleX = useSpring(scrollYProgress, {
        stiffness: 100,
        damping: 30,
        restDelta: 0.001
    });

    const navLinks = [
        { href: '/', label: dict.common.home },
        { href: '/about', label: dict.common.about },
        { href: '/contact', label: dict.common.contact },
    ];

    return (
        <header className={cn(
            "fixed top-0 left-0 right-0 z-60 transition-all duration-500",
            scrolled || mobileMenuOpen ? "py-4 bg-surface/80 backdrop-blur-xl border-b border-border shadow-md" : "py-6 md:py-8 bg-transparent",
            className
        )}>
            {/* Scroll Progress Bar */}
            <motion.div
                className="absolute bottom-0 left-0 right-0 h-[2px] bg-accent origin-left"
                style={{ scaleX }}
            />
            <div className="container-wide flex items-center justify-between relative z-100">
                <Logo />

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

            {/* Mobile Navigation Overlay */}
            <AnimatePresence mode="wait">
                {mobileMenuOpen && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        transition={{ duration: 0.2 }}
                        className="fixed inset-0 z-120 bg-surface flex flex-col p-8 pt-32 h-screen w-screen overflow-y-auto"
                    >
                        {/* Internal Close Button (Failsafe) */}
                        <button
                            onClick={() => setMobileMenuOpen(false)}
                            className="absolute top-6 right-6 p-4 text-text-muted hover:text-accent transition-colors bg-surface-muted rounded-2xl"
                            aria-label="Close menu"
                        >
                            <X size={32} strokeWidth={3} />
                        </button>

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
                    </motion.div>
                )}
            </AnimatePresence>
        </header >
    );
};
