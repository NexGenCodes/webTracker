"use client";

import { AlertCircle, RefreshCcw } from 'lucide-react';
import { useEffect } from 'react';

export default function BillingError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Billing error:', error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] p-8 text-center space-y-6">
      <div className="w-16 h-16 rounded-full bg-red-500/10 flex items-center justify-center">
        <AlertCircle className="w-8 h-8 text-red-500" />
      </div>
      <div className="space-y-2 max-w-md">
        <h2 className="text-2xl font-bold tracking-tight">Something went wrong</h2>
        <p className="text-text-muted">
          We encountered an error while loading your billing information. Please try again.
        </p>
      </div>
      <button
        onClick={() => reset()}
        className="flex items-center gap-2 px-6 py-3 bg-card hover:bg-card-hover border border-border rounded-lg transition-colors"
      >
        <RefreshCcw className="w-4 h-4" />
        <span>Try Again</span>
      </button>
    </div>
  );
}
