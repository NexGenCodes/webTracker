"use client";

import { useState, useCallback } from 'react';
import dynamic from 'next/dynamic';
import { TrackingSearch } from '@/components/TrackingSearch';
import { getTracking } from './actions/shipment';
import { CheckCircle, MapPin, AlertCircle, ShieldCheck, Globe, Zap, Copy, Check, Package } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ShipmentData } from '@/types/shipment';
import { useI18n } from '@/components/I18nContext';
import { Header } from '@/components/Header';
import { Footer } from '@/components/Footer';
import { FeatureCard } from '@/components/FeatureCard';

const ShipmentMap = dynamic(() => import('@/components/ShipmentMap'), {
  ssr: false,
  loading: () => <div className="h-[300px] w-full bg-surface-muted animate-pulse rounded-2xl mt-8"></div>
});

export default function Home() {
  const { dict } = useI18n();
  const [loading, setLoading] = useState(false);
  const [shippingData, setShippingData] = useState<ShipmentData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const copyToClipboard = useCallback((text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  const handleSearch = async (trackingNumber: string) => {
    setLoading(true);
    setError(null);
    setShippingData(null);

    try {
      const data = await getTracking(trackingNumber);
      if (data) {
        setShippingData(data);
      } else {
        setError(dict.common.notFound);
      }
    } catch (err) {
      setError(dict.common.error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative">
      {/* Background Flair */}
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

      <div className="w-full max-w-7xl flex flex-col flex-1 px-6 pt-32 md:pt-40 relative z-10">

        <Header />

        {/* Hero Section */}
        <div className={cn(
          "transition-all duration-1000 cubic-bezier(0.16, 1, 0.3, 1)",
          shippingData ? "py-8 opacity-50 blur-sm scale-95" : "py-24 md:py-32"
        )}>
          <div className="max-w-2xl mx-auto text-center mb-12">
            {!shippingData && (
              <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-accent/10 border border-accent/20 text-accent text-xs font-bold uppercase tracking-widest mb-6 animate-fade-in">
                <ShieldCheck size={14} />
                {dict.common.safeLogistics}
              </div>
            )}
            <TrackingSearch onSearch={handleSearch} isLoading={loading} />
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="w-full max-w-xl mx-auto p-5 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-4 text-error animate-fade-in mb-12">
            <AlertCircle />
            <p className="font-medium">{error}</p>
          </div>
        )}

        {/* Search Results */}
        {shippingData && (
          <div className="mb-24 animate-scale-in">
            <div className="glass-panel p-8 md:p-14 shadow-3xl border-border/50 overflow-hidden relative">
              <div className="absolute top-0 right-0 w-96 h-96 bg-accent/5 rounded-full -mr-48 -mt-48 blur-3xl pointer-events-none" />

              {/* Status Header */}
              <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-10 mb-16 pb-12 border-b border-border relative z-10">
                <div className="flex items-center gap-6">
                  <div className="relative">
                    <div className="absolute inset-0 bg-accent blur-3xl opacity-20 animate-pulse" />
                    <div className="relative w-24 h-24 bg-accent rounded-[2.5rem] flex items-center justify-center text-white shadow-2xl shadow-accent/40">
                      <Package size={36} strokeWidth={2.5} />
                    </div>
                  </div>
                  <div>
                    <span className="text-accent text-[11px] font-black uppercase tracking-[0.4em] mb-2 block">{dict.shipment.status}</span>
                    <h2 className="text-4xl md:text-6xl font-black text-text-main tracking-tighter uppercase leading-none">
                      {shippingData.isArchived ? dict.shipment.finalized : shippingData.status.replace(/_/g, ' ')}
                    </h2>
                  </div>
                </div>

                <div className="flex flex-col items-start md:items-end">
                  <span className="text-text-muted text-[11px] font-black uppercase tracking-[0.4em] mb-3">{dict.shipment.trackingId}</span>
                  <div className="flex items-center gap-4 bg-surface-muted px-6 py-4 rounded-3xl border border-border group/copy transition-all hover:border-accent/30 shadow-inner">
                    <span className="font-mono text-2xl font-black tracking-widest text-text-main group-hover:text-accent transition-colors">{shippingData.trackingNumber}</span>
                    <button
                      onClick={() => copyToClipboard(shippingData.trackingNumber)}
                      className="p-2.5 hover:bg-accent/10 rounded-2xl text-text-muted hover:text-accent transition-all active:scale-90"
                    >
                      {copied ? <Check size={20} className="text-success" /> : <Copy size={20} strokeWidth={2.5} />}
                    </button>
                  </div>
                </div>
              </div>

              {shippingData.isArchived ? (
                <div className="py-24 flex flex-col items-center text-center">
                  <div className="w-32 h-32 bg-success/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner rotate-3 transition-transform hover:rotate-0 duration-500">
                    <CheckCircle className="w-16 h-16 text-success" />
                  </div>
                  <div className="max-w-2xl">
                    <h3 className="text-5xl md:text-7xl font-black mb-6 tracking-tighter uppercase text-gradient">{dict.shipment.deliveredTitle}</h3>
                    <p className="text-text-muted text-xl leading-relaxed font-bold opacity-80">
                      {dict.shipment.deliveredDesc}
                    </p>
                  </div>
                </div>
              ) : (
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-20 relative z-10">
                  <div className="lg:col-span-5 space-y-16">
                    <div className="space-y-8">
                      <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
                        {dict.shipment.details}
                        <span className="h-px flex-1 bg-accent/20" />
                      </h4>
                      <div className="space-y-2">
                        {[
                          { label: dict.shipment.receiver, value: shippingData.receiverName },
                          { label: dict.shipment.from, value: shippingData.senderName },
                          { label: dict.shipment.destination, value: shippingData.receiverCountry, italic: true }
                        ].map((detail, idx) => (
                          <div key={idx} className="flex justify-between items-center py-6 border-b border-border last:border-0 group/item">
                            <span className="text-text-muted font-black text-xs uppercase tracking-widest">{detail.label}</span>
                            <span className={cn("font-black text-text-main text-xl group-hover:text-accent transition-colors", detail.italic && "italic")}>{detail.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="space-y-8">
                      <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
                        {dict.shipment.latestUpdates}
                        <span className="h-px flex-1 bg-accent/20" />
                      </h4>
                      <div className="space-y-12 relative border-l-2 border-border ml-3 pl-10 py-2">
                        {shippingData.events?.map((event: any, i: number) => (
                          <div key={event.id} className="relative group/event">
                            <div className={cn(
                              "absolute left-[-47px] top-1.5 w-6 h-6 rounded-full border-4 border-bg transition-all duration-500",
                              i === 0 ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border group-hover/event:bg-accent/40"
                            )} />
                            <div className="flex items-center gap-4 mb-2">
                              <p className={cn("font-black text-xl tracking-tight leading-none uppercase", i === 0 ? "text-text-main" : "text-text-muted")}>
                                {event.status.replace(/_/g, ' ')}
                              </p>
                              {i === 0 && (
                                <span className="bg-accent text-white text-[9px] font-black uppercase px-2.5 py-1 rounded-lg animate-pulse tracking-widest">{dict.shipment.live}</span>
                              )}
                            </div>
                            <span className="text-[10px] font-black text-accent/60 uppercase tracking-[0.2em]">
                              {new Date(event.timestamp).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
                            </span>
                            <div className="mt-4 flex items-center gap-3 text-sm font-bold text-text-muted bg-surface-muted/50 p-4 rounded-2xl border border-border group-hover/event:border-accent/20 transition-colors">
                              <MapPin size={16} className="text-accent/60" />
                              {event.location}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div className="lg:col-span-7 flex flex-col h-full">
                    <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent mb-8 flex items-center gap-4">
                      {dict.shipment.liveLocation}
                      <span className="h-px flex-1 bg-accent/20" />
                    </h4>
                    <div className="flex-1 rounded-[2.5rem] overflow-hidden border border-border shadow-3xl relative z-0 min-h-[500px]">
                      <ShipmentMap locationName={shippingData.events?.[0]?.location || dict.common.default} />
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Placeholder Features */}
        {!shippingData && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-24 animate-fade-in delay-200">
            <FeatureCard
              icon={<ShieldCheck />}
              title={dict.hero.feature1Title}
              description={dict.hero.feature1Desc}
            />
            <FeatureCard
              icon={<Globe />}
              title={dict.hero.feature2Title}
              description={dict.hero.feature2Desc}
            />
            <FeatureCard
              icon={<Zap />}
              title={dict.hero.feature3Title}
              description={dict.hero.feature3Desc}
            />
          </div>
        )}

        <Footer />
      </div>
    </main>
  );
}
