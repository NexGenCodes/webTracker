"use client";

import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { LogOut, LayoutDashboard, CreditCard, Menu, X, Settings, ArrowLeft, Shield } from 'lucide-react';
import { Header } from './Header';
import { Logo } from './Logo';
import { LanguageToggle } from '@/components/shared/LanguageToggle';
import { ThemeToggle } from '@/components/shared/ThemeToggle';
import { cn } from '@/lib/utils';
import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { MobileNavOverlay } from './MobileNavOverlay';

import { logoutAction, checkAuthAction } from '@/app/actions/auth';

const dashboardNavLinks = [
    { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/dashboard/billing', label: 'Billing', icon: CreditCard },
    { href: '/dashboard/settings', label: 'Settings', icon: Settings },
];

export function LayoutHeader() {
    const pathname = usePathname();
    const isDashboard = pathname?.startsWith('/dashboard') || pathname?.startsWith('/super-admin');
    const isAuth = pathname?.startsWith('/auth');
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [isSuperAdmin, setIsSuperAdmin] = useState(false);

    // Check role on mount via server action (reads HttpOnly JWT cookie server-side)
    useEffect(() => {
        checkAuthAction().then(({ user }) => {
            if (user?.role === 'super_admin') setIsSuperAdmin(true);
        });
    }, []);

    // Close mobile menu on route change
    useEffect(() => {
        setMobileMenuOpen(false);
    }, [pathname]);

    const handleSignOut = async () => {
        await logoutAction();
        // Force a full browser reload to the home page to clear all router caches
        window.location.href = '/';
    };

    if (isAuth) {
        return null;
    }

    if (isDashboard) {
        return (
            <>
                <header className="sticky top-0 z-[1000] bg-surface/95 backdrop-blur-md border-b border-border shadow-sm">
                    <div className="container-wide flex flex-row justify-between items-center gap-4 py-4">
                        <Logo href="/dashboard" className="cursor-pointer" />

                        {/* Desktop Navigation */}
                        <nav className="hidden md:flex items-center gap-1">
                            {isSuperAdmin && (
                                <Link
                                    href="/super-admin"
                                    className={cn(
                                        "flex items-center justify-center gap-2 p-2.5 lg:px-4 lg:py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all duration-200",
                                        pathname?.startsWith('/super-admin')
                                            ? "bg-accent/10 text-accent"
                                            : "text-text-muted hover:text-text-main hover:bg-surface-muted"
                                    )}
                                    title="Super Admin"
                                >
                                    <Shield className="w-5 h-5 lg:w-4 lg:h-4" />
                                    <span className="hidden lg:inline">Admin</span>
                                </Link>
                            )}
                            {dashboardNavLinks.map((link) => {
                                const isActive = pathname === link.href || (link.href !== '/dashboard' && pathname?.startsWith(link.href));
                                const isExactDashboard = link.href === '/dashboard' && pathname === '/dashboard';
                                const active = isExactDashboard || (link.href !== '/dashboard' && isActive);
                                return (
                                    <Link
                                        key={link.href}
                                        href={link.href}
                                        className={cn(
                                            "flex items-center justify-center gap-2 p-2.5 lg:px-4 lg:py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all duration-200",
                                            active
                                                ? "bg-accent/10 text-accent"
                                                : "text-text-muted hover:text-text-main hover:bg-surface-muted"
                                        )}
                                        title={link.label}
                                    >
                                        <link.icon className="w-5 h-5 lg:w-4 lg:h-4" />
                                        <span className="hidden lg:inline">{link.label}</span>
                                    </Link>
                                );
                            })}
                        </nav>

                        <div className="flex items-center gap-2 sm:gap-4">
                            <Link href="/" className="hidden md:flex items-center gap-2 p-2 lg:px-3 lg:py-2 text-xs font-black uppercase tracking-widest text-text-muted hover:text-text-main hover:bg-surface-muted rounded-xl transition-colors" title="Back to Website">
                                <ArrowLeft size={20} className="xl:hidden" />
                                <ArrowLeft size={14} className="hidden xl:block" />
                                <span className="hidden xl:inline">Website</span>
                            </Link>
                            <LanguageToggle />
                            <ThemeToggle />
                            <div className="hidden md:block h-6 w-px bg-border mx-1" />
                            <button
                                onClick={handleSignOut}
                                className="hidden md:flex bg-error/10 hover:bg-error text-error hover:text-white rounded-xl transition-all active:scale-95 items-center justify-center p-2.5 lg:px-4 lg:py-2.5"
                                aria-label="Logout"
                            >
                                <LogOut size={20} className="xl:hidden" />
                                <span className="hidden xl:inline text-xs font-black uppercase tracking-widest">Logout</span>
                            </button>

                            {/* Mobile Menu Toggle */}
                            <button
                                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                                className="md:hidden p-2.5 text-text-main hover:bg-surface-muted rounded-xl transition-colors"
                                aria-label="Toggle menu"
                            >
                                {mobileMenuOpen ? <X size={22} /> : <Menu size={22} />}
                            </button>
                        </div>
                    </div>
                </header>

                {/* Mobile Navigation Overlay */}
                <MobileNavOverlay isOpen={mobileMenuOpen} onClose={() => setMobileMenuOpen(false)} zIndex="z-[9999]">
                    <nav className="flex flex-col gap-2">
                        {isSuperAdmin && (
                            <motion.div
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ delay: 0.02, ease: "easeOut" }}
                            >
                                <Link
                                    href="/super-admin"
                                    onClick={() => setMobileMenuOpen(false)}
                                    className={cn(
                                        "flex items-center gap-4 px-6 py-5 rounded-2xl text-lg font-black uppercase tracking-tight transition-all active:scale-[0.98]",
                                        pathname?.startsWith('/super-admin')
                                            ? "bg-accent/10 text-accent border border-accent/20"
                                            : "text-text-main hover:bg-surface-muted border border-transparent"
                                    )}
                                >
                                    <Shield size={22} />
                                    Super Admin
                                </Link>
                            </motion.div>
                        )}
                        {dashboardNavLinks.map((link, i) => {
                            const isActive = pathname === link.href || (link.href !== '/dashboard' && pathname?.startsWith(link.href));
                            const isExactDashboard = link.href === '/dashboard' && pathname === '/dashboard';
                            const active = isExactDashboard || (link.href !== '/dashboard' && isActive);
                            return (
                                <motion.div
                                    key={link.href}
                                    initial={{ opacity: 0, x: -20 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    transition={{ delay: 0.05 + i * 0.05, ease: "easeOut" }}
                                >
                                    <Link
                                        href={link.href}
                                        onClick={() => setMobileMenuOpen(false)}
                                        className={cn(
                                            "flex items-center gap-4 px-6 py-5 rounded-2xl text-lg font-black uppercase tracking-tight transition-all active:scale-[0.98]",
                                            active
                                                ? "bg-accent/10 text-accent border border-accent/20"
                                                : "text-text-main hover:bg-surface-muted border border-transparent"
                                        )}
                                    >
                                        <link.icon size={22} />
                                        {link.label}
                                    </Link>
                                </motion.div>
                            );
                        })}
                    </nav>

                    {/* Mobile Go Home / Sign Out */}
                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.25 }}
                        className="flex flex-col gap-3 pt-6 pb-10"
                    >
                        <Link
                            href="/"
                            onClick={() => setMobileMenuOpen(false)}
                            className="flex items-center justify-center gap-3 w-full py-4 bg-surface hover:bg-surface-muted text-text-main rounded-2xl font-black uppercase tracking-widest text-sm transition-all active:scale-95 border border-border"
                        >
                            <ArrowLeft size={18} />
                            Back to Website
                        </Link>
                        <button
                            onClick={() => {
                                setMobileMenuOpen(false);
                                handleSignOut();
                            }}
                            className="flex items-center justify-center gap-3 w-full py-4 bg-error/10 hover:bg-error text-error hover:text-white rounded-2xl font-black uppercase tracking-widest text-sm transition-all active:scale-95 border border-error/20"
                        >
                            <LogOut size={18} />
                            Sign Out
                        </button>
                    </motion.div>
                </MobileNavOverlay>

            </>
        );
    }

    return <Header />;
}
