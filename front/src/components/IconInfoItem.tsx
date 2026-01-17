"use client";

import { cn } from '@/lib/utils';
import { LucideIcon } from 'lucide-react';

interface IconInfoItemProps {
    icon: LucideIcon;
    label: string;
    value: string;
    className?: string;
    iconClassName?: string;
}

export const IconInfoItem: React.FC<IconInfoItemProps> = ({
    icon: Icon,
    label,
    value,
    className,
    iconClassName
}) => {
    return (
        <div className={cn("flex items-start gap-5 group", className)}>
            <div className={cn(
                "p-3 bg-accent/10 rounded-xl text-accent group-hover:scale-110 transition-transform shadow-inner",
                iconClassName
            )}>
                <Icon size={20} strokeWidth={2.5} />
            </div>
            <div className="text-sm">
                <p className="text-text-main font-bold mb-0.5 tracking-tight">{label}</p>
                <p className="text-text-muted font-medium italic">{value}</p>
            </div>
        </div>
    );
};
