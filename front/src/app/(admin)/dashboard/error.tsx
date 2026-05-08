"use client";

import { useEffect } from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";
import * as Sentry from "@sentry/nextjs";

export default function DashboardError({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    useEffect(() => {
        Sentry.captureException(error);
    }, [error]);

    return (
        <div className="min-h-[60vh] flex items-center justify-center px-4">
            <div className="glass-panel p-8 md:p-12 max-w-md w-full text-center space-y-6">
                <div className="w-16 h-16 rounded-2xl bg-error/10 border border-error/20 flex items-center justify-center mx-auto">
                    <AlertTriangle className="text-error" size={28} />
                </div>

                <div className="space-y-2">
                    <h2 className="text-xl font-black uppercase tracking-tight text-text-main">
                        Something Went Wrong
                    </h2>
                    <p className="text-sm font-medium text-text-muted leading-relaxed">
                        The dashboard encountered an unexpected error. Your data is safe — try refreshing.
                    </p>
                </div>

                <button
                    onClick={reset}
                    className="btn-primary inline-flex items-center gap-2 mx-auto"
                >
                    <RefreshCw size={16} />
                    Try Again
                </button>

                {error.digest && (
                    <p className="text-[10px] font-bold uppercase tracking-widest text-text-muted opacity-40">
                        Error ID: {error.digest}
                    </p>
                )}
            </div>
        </div>
    );
}
