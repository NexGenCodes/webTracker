import { Suspense } from 'react';
import { TrackContent } from '@/components/tracking/TrackContent';

export default function Track() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
      </div>
    }>
      <TrackContent />
    </Suspense>
  );
}
