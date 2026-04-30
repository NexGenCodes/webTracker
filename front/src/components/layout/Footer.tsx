"use client";

import Link from 'next/link';
import { Logo } from './Logo';
import { useI18n } from '@/components/providers/I18nContext';
import { cn } from '@/lib/utils';
import { useCompanySettings } from '@/hooks/useCompanySettings';

interface FooterProps {
    minimal?: boolean;
    className?: string;
}

export const Footer: React.FC<FooterProps> = ({ minimal = false, className }) => {
    const { dict } = useI18n();
    const { settings } = useCompanySettings();
    const currentYear = new Date().getFullYear();

    if (minimal) {
        return (
            <footer className={cn("mt-auto py-20 text-center", className)}>
                <div className="h-px w-10 bg-accent/20 mx-auto mb-10" />
                <p className="text-text-muted text-[10px] font-black uppercase tracking-[0.5em] opacity-40">
                    &copy; {currentYear} {settings.companyName || "CargoHive"} &bull; {dict.common.globalSystems}
                </p>
            </footer>
        );
    }

    return (
        <footer className={cn("mt-auto py-24 border-t border-border", className)}>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-10 md:gap-16 mb-20">
                <div className="col-span-1 md:col-span-1">
                    <Logo className="mb-8" iconClassName="group-hover:rotate-12" />
                    <p className="text-text-muted text-sm leading-relaxed mb-8 font-medium">
                        {dict.common.footerDesc}
                    </p>
                </div>

                <div>
                    <h4 className="font-black uppercase text-xs tracking-[0.2em] mb-8 text-text-main opacity-60">{dict.common.product || "Product"}</h4>
                    <ul className="space-y-5 text-sm text-text-muted font-bold">
                        <li><Link href="/track" className="hover:text-accent transition-colors">{dict.common.trackShipment || "Track Shipment"}</Link></li>
                        <li><Link href="/pricing" className="hover:text-accent transition-colors">{dict.common.pricing || "Pricing"}</Link></li>
                    </ul>
                </div>

                <div>
                    <h4 className="font-black uppercase text-xs tracking-[0.2em] mb-8 text-text-main opacity-60">{dict.common.company}</h4>
                    <ul className="space-y-5 text-sm text-text-muted font-bold">
                        <li><Link href="/about" className="hover:text-accent transition-colors">{dict.common.about}</Link></li>
                        <li><a href="mailto:support@cargohive.com" className="hover:text-accent transition-colors">{dict.common.contact}</a></li>
                    </ul>
                </div>

                <div>
                    <h4 className="font-black uppercase text-xs tracking-[0.2em] mb-8 text-text-main opacity-60">{dict.common.legal}</h4>
                    <ul className="space-y-5 text-sm text-text-muted font-bold">
                        <li><Link href="/privacy" className="hover:text-accent transition-colors">{dict.common.privacy}</Link></li>
                        <li><Link href="/terms" className="hover:text-accent transition-colors">{dict.common.terms}</Link></li>
                    </ul>
                </div>
            </div>

            <div className="pt-10 border-t border-border flex flex-col md:flex-row justify-between items-center gap-8">
                <p className="text-text-muted text-[10px] font-black uppercase tracking-[0.3em]">
                    &copy; {currentYear} {settings.companyName || "CargoHive"} {dict.common.corp}
                </p>
                <div className="flex gap-6">
                    {[
                        { s: 'FB', h: 'https://facebook.com' },
                        { s: 'TW', h: 'https://x.com' },
                        { s: 'LI', h: 'https://linkedin.com' }
                    ].map(link => (
                        <a
                            key={link.s}
                            href={link.h}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="w-10 h-10 rounded-xl border border-border flex items-center justify-center text-xs font-black text-text-muted hover:border-accent hover:text-accent hover:bg-accent/5 hover:-translate-y-1 transition-all shadow-sm"
                        >
                            {link.s}
                        </a>
                    ))}
                </div>
            </div>
        </footer>
    );
};
