import { Metadata } from 'next';
import { getTracking } from '@/app/actions/shipment';
import Home from '@/app/page';

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
  // We reuse the Home component logic which already supports ?id= query param
  // but we can also just render a specialized view if needed.
  // We pass the ID to Home to ensure it loads directly.

  return <Home initialId={id} />;
}
