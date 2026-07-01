import { Loader2 } from 'lucide-react';

export default function BillingLoading() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
      <Loader2 className="w-8 h-8 text-accent animate-spin" />
      <p className="text-text-muted font-medium animate-pulse">Loading billing details...</p>
    </div>
  );
}
