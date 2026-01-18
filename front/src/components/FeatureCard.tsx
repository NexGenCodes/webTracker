"use client";

import React, { memo } from 'react';
import { cn } from '@/lib/utils';
import { ReactNode } from 'react';

interface FeatureCardProps {
    icon: ReactNode;
    title: string;
    description: string;
    className?: string;
}

export const FeatureCard: React.FC<FeatureCardProps> = memo(({
    icon,
    title,
    description,
    className
}) => {
    return (
        <div className={cn(
            "glass-panel p-8 md:p-12 hover:scale-[1.02] hover:-translate-y-2 transition-all duration-500 cursor-default shadow-xl border-border/50 hover:border-accent/30 group",
            className
        )}>
            <div className="w-16 h-16 bg-accent/5 rounded-2xl flex items-center justify-center text-accent mb-10 shadow-inner group-hover:bg-accent group-hover:text-white transition-all duration-500 ring-1 ring-accent/10">
                <div className="scale-125 transition-transform duration-500">{icon}</div>
            </div>
            <h3 className="text-2xl font-black mb-4 text-text-main tracking-tight uppercase leading-none">{title}</h3>
            <p className="text-text-muted text-base leading-relaxed font-semibold opacity-80 group-hover:opacity-100 transition-opacity">{description}</p>
        </div>
    );
});
