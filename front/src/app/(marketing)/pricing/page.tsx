import { PricingSection } from '@/components/landing/PricingSection';
import { Metadata } from 'next';

export const metadata: Metadata = {
    title: 'Pricing - WebTracker',
    description: 'Simple, transparent pricing for logistics companies.',
};

export default function PricingPage() {
    return (
        <main className="min-h-screen pt-24 pb-16">
            <PricingSection />
        </main>
    );
}
