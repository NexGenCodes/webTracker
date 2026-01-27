"use client";

import { useState, useCallback, useEffect, Suspense } from 'react';
import dynamic from 'next/dynamic';
import { useSearchParams } from 'next/navigation';
import { TrackingSearch } from '@/components/TrackingSearch';
import { getTracking } from './actions/shipment';
import { CheckCircle, MapPin, AlertCircle, ShieldCheck, Globe, Zap, Copy, Check, Package } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ShipmentData } from '@/types/shipment';
import { useI18n } from '@/components/I18nContext';
import { Footer } from '@/components/Footer';
import { FeatureCard } from '@/components/FeatureCard';
import { toast } from 'react-hot-toast';



function HomeContent() {
  const { dict } = useI18n();
  const searchParams = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [shippingData, setShippingData] = useState<ShipmentData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const initialId = searchParams.get('id');

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
    } catch (err) {
      const errorMsg = dict.common.error;
      setError(errorMsg);
      toast.error(errorMsg, {
        icon: '⚠️',
      });
    } finally {
      setLoading(false);
    }
  }, [dict, dict.admin, dict.common]);

  // Deep linking effect
  useEffect(() => {
    if (initialId) {
      handleSearch(initialId);
    }
  }, [initialId, handleSearch]);

  const copyToClipboard = useCallback((text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success(dict.admin?.copied || 'Copied!', {
      duration: 2000,
      icon: '📋',
    });
    setTimeout(() => setCopied(false), 2000);
  }, [dict.admin]);

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


        {/* Hero Section */}
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

        {/* Error */}
        {error && (
          <div className="w-full max-w-xl mx-auto p-5 bg-error/10 border border-error/20 rounded-2xl flex items-center gap-4 text-error animate-fade-in mb-12">
            <AlertCircle />
            <p className="font-medium">{error}</p>
          </div>
        )}

        {/* Arrival Notification */}
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
                    📦 Package Arrived!
                  </h3>
                  <p className="text-text-muted font-bold mb-3">
                    Our local agent will contact you shortly.
                  </p>
                  <div className="flex items-center gap-2 text-sm font-black text-accent/80">
                    <span>Expected delivery: Tomorrow, 8:00 AM - 10:00 AM</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Search Results */}
        {shippingData && (
          <div className="mb-24 animate-scale-in">
            <div className="glass-panel p-5 md:p-14 shadow-3xl border-border/50 overflow-hidden relative">
              <div className="absolute top-0 right-0 w-96 h-96 bg-accent/5 rounded-full -mr-48 -mt-48 blur-3xl pointer-events-none" />

              {/* Status Header - Mobile First */}
              <div className="flex flex-col gap-6 md:flex-row md:justify-between md:items-center mb-8 md:mb-16 pb-8 md:pb-12 border-b border-border relative z-10">
                <div className="flex items-center gap-4 md:gap-6">
                  <div className="relative">
                    <div className="absolute inset-0 bg-accent blur-3xl opacity-20 animate-pulse" />
                    <div className="relative w-16 h-16 md:w-24 md:h-24 bg-accent rounded-[2rem] md:rounded-[2.5rem] flex items-center justify-center text-white shadow-2xl shadow-accent/40">
                      <Package size={28} strokeWidth={2.5} className="md:hidden" />
                      <Package size={36} strokeWidth={2.5} className="hidden md:block" />
                    </div>
                  </div>
                  <div className="flex-1">
                    <span className="text-accent text-[10px] md:text-[11px] font-black uppercase tracking-[0.3em] md:tracking-[0.4em] mb-1 md:mb-2 block">{dict.shipment.status}</span>
                    <h2 className="text-xl md:text-4xl lg:text-6xl font-black text-text-main tracking-tighter uppercase leading-none">
                      {shippingData.isArchived ? dict.shipment.finalized : shippingData.status.replace(/_/g, ' ')}
                    </h2>
                  </div>
                </div>

                <div className="flex flex-col items-start md:items-end w-full md:w-auto">
                  <span className="text-text-muted text-[10px] md:text-[11px] font-black uppercase tracking-[0.3em] md:tracking-[0.4em] mb-2 md:mb-3">{dict.shipment.trackingId}</span>
                  <div className="flex items-center gap-2 md:gap-4 bg-surface-muted px-3 md:px-6 py-3 md:py-4 rounded-2xl md:rounded-3xl border border-border group/copy transition-all hover:border-accent/30 shadow-inner w-full md:w-auto">
                    <span className="font-mono text-sm md:text-lg lg:text-2xl font-black tracking-wide md:tracking-widest text-text-main group-hover:text-accent transition-colors break-all">{shippingData.trackingNumber}</span>
                    <button
                      onClick={() => copyToClipboard(shippingData.trackingNumber)}
                      className="p-2 md:p-2.5 hover:bg-accent/10 rounded-xl md:rounded-2xl text-text-muted hover:text-accent transition-all active:scale-90 shrink-0"
                    >
                      {copied ? <Check size={18} className="text-success md:w-5 md:h-5" /> : <Copy size={18} strokeWidth={2.5} className="md:w-5 md:h-5" />}
                    </button>
                  </div>
                </div>
              </div>

              {shippingData.status === 'CANCELED' ? (
                <div className="py-24 flex flex-col items-center text-center">
                  <div className="w-32 h-32 bg-error/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner">
                    <AlertCircle className="w-16 h-16 text-error" />
                  </div>
                  <div className="max-w-2xl">
                    <h3 className="text-3xl md:text-7xl font-black mb-6 tracking-tighter uppercase text-error">Shipment Canceled</h3>
                    <p className="text-text-muted text-xl leading-relaxed font-bold opacity-80">
                      This shipment has been canceled by the administrator. Please contact support for more details.
                    </p>
                  </div>
                </div>
              ) : shippingData.isArchived ? (
                <div className="py-24 flex flex-col items-center text-center">
                  <div className="w-32 h-32 bg-success/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner rotate-3 transition-transform hover:rotate-0 duration-500">
                    <CheckCircle className="w-16 h-16 text-success" />
                  </div>
                  <div className="max-w-2xl">
                    <h3 className="text-3xl md:text-7xl font-black mb-6 tracking-tighter uppercase text-gradient">{dict.shipment.deliveredTitle}</h3>
                    <p className="text-text-muted text-xl leading-relaxed font-bold opacity-80">
                      {dict.shipment.deliveredDesc}
                    </p>
                  </div>
                </div>
              ) : (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-20 relative z-10">
                  {/* Left Column: Details */}
                  <div className="space-y-16 order-2 lg:order-1">
                    <div className="space-y-8">
                      <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
                        {dict.shipment.details}
                        <span className="h-px flex-1 bg-accent/20" />
                      </h4>
                      <div className="space-y-2">
                        {[
                          { label: dict.shipment.receiver, value: shippingData.receiverName },
                          { label: dict.shipment.destination, value: shippingData.receiverCountry, italic: true },
                          { label: dict.shipment.from, value: shippingData.senderName },
                          { label: dict.shipment.origin, value: shippingData.senderCountry },
                          { label: "Weight", value: `${shippingData.weight || 15} KGS` },
                        ].map((detail, idx) => (
                          <div key={idx} className="flex justify-between items-center py-6 border-b border-border last:border-0 group/item">
                            <span className="text-text-muted font-black text-xs uppercase tracking-widest">{detail.label}</span>
                            <span className={cn("font-black text-text-main text-lg md:text-xl group-hover:text-accent transition-colors", detail.italic && "italic")}>{detail.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Right Column: Timeline */}
                  <div className="space-y-8 order-1 lg:order-2">
                    <h4 className="text-[10px] font-black uppercase tracking-[0.3em] text-accent flex items-center gap-4">
                      {dict.shipment.latestUpdates}
                      <span className="h-px flex-1 bg-accent/20" />
                    </h4>
                    <div className="space-y-8 md:space-y-12 relative border-l-2 border-border ml-2 md:ml-3 pl-6 md:pl-10 py-2">
                      {shippingData.timeline ? (
                        shippingData.timeline.map((event, i) => (
                          <div key={i} className="relative group/event">
                            <div className={cn(
                              "absolute left-[-35px] md:left-[-47px] top-1.5 w-5 h-5 md:w-6 md:h-6 rounded-full border-3 md:border-4 border-bg transition-all duration-500",
                              event.is_completed ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border opacity-50"
                            )} />
                            <div className="flex flex-wrap items-center gap-2 md:gap-4 mb-2">
                              <p className={cn("font-black text-base md:text-lg lg:text-xl tracking-tight leading-none uppercase", event.is_completed ? "text-text-main" : "text-text-muted opacity-50")}>
                                {event.status}
                              </p>
                              {event.is_completed && !shippingData.timeline?.[i + 1]?.is_completed && (
                                <span className="bg-accent text-white text-[8px] md:text-[9px] font-black uppercase px-2 md:px-2.5 py-0.5 md:py-1 rounded-lg animate-pulse tracking-wider md:tracking-widest">{dict.shipment.live}</span>
                              )}
                            </div>
                            <span className="text-[9px] md:text-[10px] font-black text-accent/60 uppercase tracking-[0.15em] md:tracking-[0.2em] block mb-3">
                              {new Date(event.timestamp).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
                            </span>
                            <div className="mt-3 md:mt-4 flex items-start gap-2 md:gap-3 text-xs md:text-sm font-bold text-text-muted bg-surface-muted/50 p-3 md:p-4 rounded-xl md:rounded-2xl border border-border group-hover/event:border-accent/20 transition-colors">
                              <Package size={14} className="text-accent/60 mt-0.5 md:w-4 md:h-4 shrink-0" />
                              <span className="leading-relaxed">{event.description}</span>
                            </div>
                            {/* Additional Timeline Detail Layers */}
                            {event.is_completed && (
                              <>
                                {/* First Detail Layer: Location & Date */}
                                <div className="mt-2 md:mt-3 flex items-center gap-2 text-[10px] md:text-xs font-bold text-text-muted/70 bg-surface/30 px-3 py-2 rounded-lg border border-border/50">
                                  <MapPin size={12} className="text-accent/50 shrink-0 md:w-3.5 md:h-3.5" />
                                  <span className="truncate">
                                    {event.status.toLowerCase().includes('delivered') || event.status.toLowerCase().includes('arrival')
                                      ? `${shippingData.receiverCountry} (Destination)`
                                      : event.status.toLowerCase().includes('depart') || event.status.toLowerCase().includes('origin')
                                        ? `${shippingData.senderCountry} (Origin)`
                                        : 'In Transit'}
                                  </span>
                                  <span className="text-accent/30">•</span>
                                  <span className="text-[9px] md:text-[10px] opacity-60 truncate">{new Date(event.timestamp).toLocaleDateString()}</span>
                                </div>

                                {/* Second Detail Layer: Carrier & Handler Info */}
                                <div className="mt-2 flex items-center justify-between gap-2 text-[9px] md:text-[10px] text-text-muted/60 bg-surface/20 px-3 py-1.5 rounded-md border border-border/30">
                                  <div className="flex items-center gap-1.5">
                                    <span className="w-1.5 h-1.5 rounded-full bg-accent/60"></span>
                                    <span className="font-bold uppercase tracking-wider">Air Freight</span>
                                  </div>
                                  <span className="opacity-50">|</span>
                                  <div className="flex items-center gap-1">
                                    <span className="font-mono">ID: #{shippingData.trackingNumber.slice(-6)}</span>
                                  </div>
                                </div>
                              </>
                            )}
                          </div>
                        ))
                      ) : (
                        shippingData.events?.map((event: any, i: number) => (
                          <div key={event.id} className="relative group/event">
                            <div className={cn(
                              "absolute left-[-47px] top-1.5 w-6 h-6 rounded-full border-4 border-bg transition-all duration-500",
                              i === 0 ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border group-hover/event:bg-accent/40"
                            )} />
                            <div className="flex items-center gap-4 mb-2">
                              <p className={cn("font-black text-lg md:text-xl tracking-tight leading-none uppercase", i === 0 ? "text-text-main" : "text-text-muted")}>
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
                        ))
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

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

export default function Home() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
      </div>
    }>
      <HomeContent />
    </Suspense>
  );
}
