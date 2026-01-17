"use client";

import { cn } from '@/lib/utils';
import { InputHTMLAttributes, forwardRef } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    containerClassName?: string;
}

export const PremiumInput = forwardRef<HTMLInputElement, InputProps>(
    ({ label, className, containerClassName, ...props }, ref) => {
        return (
            <div className={cn("space-y-3", containerClassName)}>
                {label && (
                    <label className="text-xs font-black uppercase tracking-[0.2em] text-text-muted ml-1 opacity-60">
                        {label}
                    </label>
                )}
                <input
                    ref={ref}
                    className={cn(
                        "w-full bg-surface-muted text-text-main py-4 px-6 rounded-2xl border-2 border-transparent outline-none focus:border-accent/30 focus:bg-surface focus:ring-4 focus:ring-accent/10 transition-all font-medium shadow-inner",
                        className
                    )}
                    {...props}
                />
            </div>
        );
    }
);

PremiumInput.displayName = 'PremiumInput';

interface TextareaProps extends InputHTMLAttributes<HTMLTextAreaElement> {
    label?: string;
    containerClassName?: string;
    rows?: number;
}

export const PremiumTextarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
    ({ label, className, containerClassName, rows = 4, ...props }, ref) => {
        return (
            <div className={cn("space-y-3", containerClassName)}>
                {label && (
                    <label className="text-xs font-black uppercase tracking-[0.2em] text-text-muted ml-1 opacity-60">
                        {label}
                    </label>
                )}
                <textarea
                    ref={ref}
                    rows={rows}
                    className={cn(
                        "w-full bg-surface-muted text-text-main py-4 px-6 rounded-2xl border-2 border-transparent outline-none focus:border-accent/30 focus:bg-surface focus:ring-4 focus:ring-accent/10 transition-all font-medium shadow-inner resize-none",
                        className
                    )}
                    {...props}
                />
            </div>
        );
    }
);

PremiumTextarea.displayName = 'PremiumTextarea';
