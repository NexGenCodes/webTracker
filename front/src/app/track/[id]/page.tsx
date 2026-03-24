import { Metadata, ResolvingMetadata } from 'next';
import { getTracking } from '@/app/actions/shipment';
import Home from '@/app/page';

interface Props {
  params: { id: string };
}

export async function generateMetadata(
  { params }: Props,
  parent: ResolvingMetadata
): Promise<Metadata> {
  const id = params.id.toUpperCase();
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

export default function TrackPage({ params }: Props) {
  // We reuse the Home component logic which already supports ?id= query param
  // but we can also just render a specialized view if needed.
  // For now, redirecting to home with the ID is easiest for code reuse,
  // but the generateMetadata here will still work for the link preview.
  
  // Actually, let's just render the Home component and pass the ID via a redirected prop or similar,
  // but Home uses useSearchParams.
  // A cleaner way is to refactor Home to accept an initialId prop.
  
  return <Home />;
}
