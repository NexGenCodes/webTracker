import React, { memo } from 'react';
import Link from 'next/link';
import { Package } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useCompanySettings } from '@/hooks/useCompanySettings';

import { useI18n } from '@/components/providers/I18nContext';

interface LogoProps {
    className?: string;
    iconClassName?: string;
    href?: string;
}

const LogoComponent: React.FC<LogoProps> = memo(({ className, iconClassName, href = "/" }) => {
    const { dict } = useI18n();
    const { settings } = useCompanySettings();

    return (
        <Link href={href} className={cn("flex items-center gap-4 group", className)}>
            <div className="relative">
                <div className="absolute inset-0 bg-accent blur-2xl opacity-0 group-hover:opacity-30 transition-opacity rounded-full scale-150" />
                <div className="relative bg-primary/10 p-3 rounded-2xl shadow-xl shadow-primary/10 group-hover:scale-105 group-hover:bg-accent  transition-all duration-500">
                    {settings.logoUrl ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={settings.logoUrl} alt={settings.companyName || "Company"} className={cn("w-[22px] h-[22px] object-contain", iconClassName)} />
                    ) : (
                        <Package className={cn("text-accent transition-transform duration-500 group-hover:text-white", iconClassName)} size={22} strokeWidth={2.5} />
                    )}
                </div>
            </div>
            <div className="hidden min-[470px]:flex flex-col">
                <span className="text-gradient uppercase font-black text-lg sm:text-2xl leading-none tracking-tighter">
                    {settings.companyName || "CargoHive"}
                </span>
                <span className="text-[9px] font-black uppercase tracking-[0.4em] text-text-muted opacity-60 ml-0.5 mt-1">
                    {dict.common.tagline}
                </span>
            </div>
        </Link>
    );
});

LogoComponent.displayName = 'Logo';
export const Logo = LogoComponent;
