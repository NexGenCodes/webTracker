"use client";

import { useState, useCallback, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import { toast } from 'react-hot-toast';
import { TrackingSearch } from '@/components/tracking/TrackingSearch';
import { getTracking } from '@/app/actions/shipment';
import { createClient } from '@/lib/supabase/client';
import { AlertCircle, ShieldCheck, Package } from 'lucide-react';
import { ShipmentData } from '@/types/shipment';
import { useI18n } from '@/components/providers/I18nContext';
import { Footer } from '@/components/layout/Footer';
import { useShipmentProgress } from '@/hooks/useShipmentProgress';

import { ShipmentStatusHeader } from './ShipmentStatusHeader';
import { ShipmentMapBar } from './ShipmentMapBar';
import { ShipmentDetails } from './ShipmentDetails';
import { ShipmentTimeline } from './ShipmentTimeline';
import { ShipmentTerminalState } from './ShipmentTerminalStates';

export interface TrackProps {
  initialId?: string | null;
}

export function TrackContent({ initialId: propId }: TrackProps) {
  const { dict } = useI18n();
  const searchParams = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [shippingData, setShippingData] = useState<ShipmentData | null>(null);
  const liveProgress = useShipmentProgress(shippingData);
  
  const [originCoords, setOriginCoords] = useState<[number, number]>([51.5074, -0.1278]);
  const [destCoords, setDestCoords] = useState<[number, number]>([40.7128, -74.0060]);
  
  useEffect(() => {
    if (!shippingData) return;

    let isMounted = true;

    const fetchCoords = async (country: string): Promise<[number, number] | null> => {
      try {
        const res = await fetch(`https://restcountries.com/v3.1/name/${encodeURIComponent(country)}?fullText=true`);
        if (!res.ok) {
           const resFallback = await fetch(`https://restcountries.com/v3.1/name/${encodeURIComponent(country)}`);
           if (!resFallback.ok) return null;
           const data = await resFallback.json();
           return data[0]?.latlng as [number, number];
        }
        const data = await res.json();
        return data[0]?.latlng as [number, number];
      } catch {
        return null;
      }
    };

    if (shippingData.originCoords) {
      setOriginCoords(shippingData.originCoords);
    } else if (shippingData.senderCountry && shippingData.senderCountry !== 'N/A') {
      fetchCoords(shippingData.senderCountry).then(coords => {
        if (coords && isMounted) setOriginCoords(coords);
      });
    }

    if (shippingData.destinationCoords) {
      setDestCoords(shippingData.destinationCoords);
    } else if (shippingData.receiverCountry && shippingData.receiverCountry !== 'N/A') {
      fetchCoords(shippingData.receiverCountry).then(coords => {
        if (coords && isMounted) setDestCoords(coords);
      });
    }

    return () => { isMounted = false; };
  }, [shippingData]);

  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const initialId = propId || searchParams.get('id');

  // Viewport manipulation removed — forcing width=1280 broke mobile UX.

  const handleSearch = useCallback(async (trackingNumber: string) => {
    const id = trackingNumber.trim().toUpperCase();
    if (!id) return;

    setLoading(true);
    setError(null);
    setShippingData(null);

    try {
      const data = await getTracking(id);
      if (data) {
        setShippingData(data);
        toast.success(dict.admin?.success || 'Shipment found', {
          icon: '📦',
        });
      } else {
        const errorMsg = dict.common.notFound;
        setError(errorMsg);
        toast.error(errorMsg, {
          icon: '🔍',
        });
      }
    } catch {
      const errorMsg = dict.common.error;
      setError(errorMsg);
      toast.error(errorMsg, {
        icon: '⚠️',
      });
    } finally {
      setLoading(false);
    }
  }, [dict]);

  useEffect(() => {
    if (initialId) {
      handleSearch(initialId);
    }
  }, [initialId, handleSearch]);

  // Supabase Realtime Subscription
  useEffect(() => {
    if (!shippingData?.trackingNumber) return;

    const supabase = createClient();
    const channel = supabase
      .channel(`tracking-${shippingData.trackingNumber}`)
      .on(
        'postgres_changes',
        {
          event: 'UPDATE',
          schema: 'public',
          table: 'shipment',
          filter: `tracking_id=eq.${shippingData.trackingNumber}`,
        },
        () => {
          // Re-fetch the data when an update occurs
          handleSearch(shippingData.trackingNumber);
        }
      )
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, [shippingData?.trackingNumber, handleSearch]);


  if (!mounted) return null;

  return (
    <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative">
      <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden">
        <div className="absolute inset-0 bg-dot-grid opacity-[0.1]" />
        <div className="bg-stars-layer opacity-[0.4]" />
        <div className="absolute inset-0 bg-topography opacity-[0.2]" />
        <div className="shooting-star" style={{ top: '10%', left: '80%', animationDelay: '2s' }} />
        <div className="shooting-star" style={{ top: '30%', left: '40%', animationDelay: '7s' }} />
        <div className="shooting-star" style={{ top: '50%', left: '90%', animationDelay: '15s' }} />
        <div className="absolute top-0 right-0 w-[600px] h-[600px] bg-accent/5 blur-[120px] rounded-full" />
        <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-primary/5 blur-[100px] rounded-full" />
      </div>

      <div className="w-full max-w-7xl flex flex-col flex-1 px-3 md:px-6 pt-32 md:pt-40 relative z-10">

        {!shippingData && (
          <div className="py-12 md:py-32">
            <div className="max-w-2xl mx-auto text-center mb-12">
              <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-xs font-bold uppercase tracking-widest mb-6 animate-fade-in">
                <ShieldCheck size={14} />
                {dict.common.safeLogistics}
              </div>
              <TrackingSearch onSearch={handleSearch} isLoading={loading} />
            </div>
          </div>
        )}

        {error && (
          <div className="w-full max-w-xl mx-auto p-5 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-4 text-error animate-fade-in mb-12">
            <AlertCircle />
            <p className="font-medium">{error}</p>
          </div>
        )}

        {shippingData && shippingData.status === 'OUT_FOR_DELIVERY' && (
          <div className="w-full max-w-2xl mx-auto mb-8 animate-fade-in">
            <div className="glass-panel p-6 md:p-8 border-accent/30 bg-accent/5 relative overflow-hidden">
              <div className="absolute top-0 right-0 w-64 h-64 bg-accent/10 rounded-full -mr-32 -mt-32 blur-3xl" />
              <div className="relative z-10 flex flex-col md:flex-row items-start md:items-center gap-4">
                <div className="w-16 h-16 bg-accent rounded-2xl flex items-center justify-center shrink-0">
                  <Package size={32} className="text-white" />
                </div>
                <div className="flex-1">
                  <h3 className="text-xl md:text-2xl font-black text-accent mb-2 uppercase tracking-tight">
                    {dict.shipment.outForDeliveryTitle || '🚚 Out for Delivery'}
                  </h3>
                  <p className="text-text-muted font-bold mb-3">
                    {dict.shipment.outForDeliveryDesc || 'Our local agent is en route to your destination.'}
                  </p>
                  <div className="flex items-center gap-2 text-sm font-black text-accent/80">
                    <span>{dict.shipment.expectedDelivery || 'Expected delivery:'} {shippingData.expectedDeliveryTime ? new Date(shippingData.expectedDeliveryTime).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }) : (dict.shipment.soon || 'Soon')}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {shippingData && (
          <div className="mb-24 animate-scale-in">
            <div className="glass-panel p-2 md:p-14 shadow-3xl border-border/50 overflow-hidden relative">
              <div className="absolute top-0 right-0 w-96 h-96 bg-accent/5 rounded-full -mr-48 -mt-48 blur-3xl pointer-events-none" />

              <ShipmentStatusHeader shippingData={shippingData} dict={dict} />

              {shippingData.status === 'CANCELED' ? (
                <ShipmentTerminalState type="canceled" dict={dict} />
              ) : shippingData.isArchived ? (
                <ShipmentTerminalState type="delivered" dict={dict} />
              ) : (
                <>
                  <ShipmentMapBar 
                    shippingData={shippingData} 
                    originCoords={originCoords} 
                    destCoords={destCoords} 
                    liveProgress={liveProgress} 
                    dict={dict} 
                  />
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-20 relative z-10">
                    <ShipmentDetails shippingData={shippingData} dict={dict} />
                    <ShipmentTimeline shippingData={shippingData} dict={dict} />
                  </div>
                </>
              )}
            </div>
          </div>
        )}

        <Footer />
      </div>
    </main>
  );
}
