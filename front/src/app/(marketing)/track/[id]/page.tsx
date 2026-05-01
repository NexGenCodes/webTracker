import { Metadata } from 'next';
import { getTracking } from '@/app/actions/shipment';
import { Suspense } from 'react';
import { TrackContent } from '@/components/tracking/TrackContent';

interface Props {
  params: Promise<{ id: string }>;
}

export async function generateMetadata(
  { params }: Props
): Promise<Metadata> {
  const { id: rawId } = await params;
  const id = rawId.toUpperCase();
  const shipment = await getTracking(id);

  if (!shipment) {
    return {
      title: 'Shipment Not Found | WebTracker',
    };
  }

  const status = shipment.status.replace(/_/g, ' ');
  return {
    title: `Track ${id} - ${status} | WebTracker`,
    description: `Track your shipment ${id}. Current status: ${status}. Origin: ${shipment.senderCountry}, Destination: ${shipment.receiverCountry}.`,
    openGraph: {
      title: `Shipment ${id}: ${status}`,
      description: `Real-time tracking for manifest ${id}. Currently ${status.toLowerCase()}.`,
      type: 'website',
    },
  };
}

export default async function TrackPage({ params }: Props) {
  const { id } = await params;

  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
      </div>
    }>
      <TrackContent initialId={id} />
    </Suspense>
  );
}
