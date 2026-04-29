import React from 'react';
import { AlertCircle, CheckCircle2 } from 'lucide-react';

interface AuthBannerProps {
    error?: string | null;
    successMessage?: string | null;
}

export function AuthBanner({ error, successMessage }: AuthBannerProps) {
    if (error) {
        return (
            <div className="mb-6 p-4 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-3 text-error text-sm animate-fade-in w-full">
                <AlertCircle size={18} className="shrink-0" />
                <p className="font-bold">{error}</p>
            </div>
        );
    }
    
    if (successMessage) {
        return (
            <div className="mb-6 p-4 bg-success/10 border border-success/20 rounded-2xl flex items-center gap-3 text-success text-sm animate-fade-in w-full">
                <CheckCircle2 size={18} className="shrink-0" />
                <p className="font-bold">{successMessage}</p>
            </div>
        );
    }

    return null;
}
