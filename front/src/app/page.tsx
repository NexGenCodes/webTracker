"use client";

import { useState, useCallback, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { TrackingSearch } from '@/components/TrackingSearch';
import { getTracking } from './actions/shipment';
import { CheckCircle, MapPin, AlertCircle, ShieldCheck, Globe, Zap, Copy, Package, Share2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { ShipmentData } from '@/types/shipment';
import { useI18n } from '@/components/I18nContext';
import { Footer } from '@/components/Footer';
import { FeatureCard } from '@/components/FeatureCard';
import { toast } from 'react-hot-toast';
import DynamicMap from '@/components/DynamicMap';
import { useShipmentProgress } from '@/hooks/useShipmentProgress';

interface HomeProps {
  initialId?: string | null;
}

function HomeContent({ initialId: propId }: HomeProps) {
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

  useEffect(() => {
    // Dynamic Viewport logic to simulate "Desktop Mode" on mobile when tracking results are active
    if (typeof window !== 'undefined') {
      const viewport = document.querySelector('meta[name="viewport"]');
      if (viewport) {
        if (shippingData) {
          // Force 1280px width allows md: and lg: styles to trigger on mobile
          // initial-scale=0.3 fits the 1280px width onto the mobile screen
          viewport.setAttribute('content', 'width=1280, initial-scale=0.3, user-scalable=yes');
        } else {
          // Normal responsive behavior for the landing/search hero
          viewport.setAttribute('content', 'width=device-width, initial-scale=1, user-scalable=yes');
        }
      }
    }
    
    // Cleanup on unmount
    return () => {
      const viewport = document.querySelector('meta[name="viewport"]');
      if (viewport) viewport.setAttribute('content', 'width=device-width, initial-scale=1, user-scalable=yes');
    };
  }, [shippingData]);

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

  // Deep linking effect
  useEffect(() => {
    if (initialId) {
      handleSearch(initialId);
    }
  }, [initialId, handleSearch]);


  if (!mounted) return null;

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

      <div className="w-full max-w-7xl flex flex-col flex-1 px-3 md:px-6 pt-32 md:pt-40 relative z-10">


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
            <div className="glass-panel p-2 md:p-14 shadow-3xl border-border/50 overflow-hidden relative">
              <div className="absolute top-0 right-0 w-96 h-96 bg-accent/5 rounded-full -mr-48 -mt-48 blur-3xl pointer-events-none" />

              {/* Status Header - Horizontal on mobile */}
              <div className="flex flex-row justify-between items-center mb-4 md:mb-16 pb-3 md:pb-12 border-b border-border relative z-10 gap-x-2">
                <div className="flex items-center gap-2 md:gap-6">
                  <div className="relative">
                    <div className="absolute inset-0 bg-accent blur-3xl opacity-20 animate-pulse" />
                    <div className="relative w-10 h-10 md:w-24 md:h-24 bg-accent rounded-xl md:rounded-[2.5rem] flex items-center justify-center text-white shadow-2xl shadow-accent/40">
                      <Package size={18} strokeWidth={2.5} className="md:hidden" />
                      <Package size={36} strokeWidth={2.5} className="hidden md:block" />
                    </div>
                  </div>
                  <div className="flex-1">
                    <span className="text-accent text-[8px] md:text-[11px] font-black uppercase tracking-[0.2em] md:tracking-[0.4em] mb-0.5 md:mb-2 block">{dict.shipment.status}</span>
                    <h2 className="text-xs md:text-3xl lg:text-4xl font-black text-text-main tracking-tighter uppercase leading-none">
                      {shippingData.isArchived ? dict.shipment.finalized : (dict.statuses?.[shippingData.status] || shippingData.status.replace(/_/g, ' '))}
                    </h2>
                  </div>
                </div>

                <div className="flex flex-col items-end md:items-end w-auto">
                  <span className="text-text-muted text-[7px] md:text-[11px] font-black uppercase tracking-[0.2em] md:tracking-[0.4em] mb-0.5 md:mb-3">{dict.shipment.trackingId}</span>
                  <div className="flex items-center gap-1 md:gap-4 bg-surface-muted px-1.5 md:px-6 py-1 md:py-4 rounded-lg md:rounded-3xl border border-border group/copy transition-all hover:border-accent/30 shadow-inner w-auto max-w-[120px] md:max-w-none">
                    <span className="font-mono text-[9px] md:text-lg lg:text-2xl font-black tracking-tighter md:tracking-widest text-text-main group-hover:text-accent transition-colors truncate">{shippingData.trackingNumber}</span>
                    <div className="flex items-center gap-1 ml-auto">
                      <button
                        onClick={() => {
                          navigator.clipboard.writeText(shippingData.trackingNumber);
                          toast.success(dict.admin?.copied || "Copied!");
                        }}
                        className="p-1 md:p-2 rounded-xl md:rounded-2xl transition-all bg-surface hover:bg-surface-muted text-text-muted hover:text-accent border border-border flex items-center justify-center active:scale-90"
                        title={dict.admin?.copy || "Copy"}
                      >
                        <Copy size={14} className="md:w-[18px] md:h-[18px]" />
                      </button>
                      <button
                        onClick={() => {
                          const shareData = {
                            title: `Shipment ${shippingData.trackingNumber}`,
                            text: `Tracking status for ${shippingData.trackingNumber}: ${shippingData.status}`,
                            url: `${window.location.origin}/api/receipt/${shippingData.trackingNumber}?status=${shippingData.status}&origin=${shippingData.senderCountry}&dest=${shippingData.receiverCountry}&sender=${shippingData.senderName}&receiver=${shippingData.receiverName}&weight=${shippingData.weight}%20KGS&content=${shippingData.cargoType}`,
                          };
                          if (navigator.share) {
                            navigator.share(shareData);
                          } else {
                            navigator.clipboard.writeText(shareData.url);
                            toast.success("Link copied to share!");
                          }
                        }}
                        className="p-1 md:p-2 rounded-xl md:rounded-2xl transition-all bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 flex items-center justify-center active:scale-95"
                        title="Share Receipt"
                      >
                        <Share2 size={14} className="md:w-[18px] md:h-[18px]" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              {shippingData.status === 'CANCELED' ? (
                <div className="py-24 flex flex-col items-center text-center">
                  <div className="w-32 h-32 bg-error/10 rounded-[3rem] flex items-center justify-center mb-10 shadow-inner">
                    <AlertCircle className="w-16 h-16 text-error" />
                  </div>
                  <div className="max-w-2xl">
                    <h3 className="text-2xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-error">Shipment Canceled</h3>
                    <p className="text-text-muted text-lg md:text-xl leading-relaxed font-bold opacity-80">
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
                    <h3 className="text-2xl md:text-5xl font-black mb-6 tracking-tighter uppercase text-gradient">{dict.shipment.deliveredTitle}</h3>
                    <p className="text-text-muted text-lg md:text-xl leading-relaxed font-bold opacity-80">
                      {dict.shipment.deliveredDesc}
                    </p>
                  </div>
                </div>
              ) : (
                <>
                  {/* Map Info Bar - 3 Columns on Mobile */}
                  <div className="grid grid-cols-3 gap-1.5 md:gap-4 mb-4 md:mb-6 w-full max-w-6xl mx-auto animate-fade-in md:px-0">
                    <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-surface/50 text-center md:text-left">
                      <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-text-muted/10 flex items-center justify-center shrink-0">
                        <MapPin size={12} className="text-text-muted md:w-5 md:h-5" />
                      </div>
                      <div className="min-w-0">
                        <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-text-muted block mb-0.5 truncate">{dict.shipment.from || 'Origin'}</span>
                        <span className="text-[8px] md:text-sm font-black text-text-main truncate block">{shippingData.senderCountry}</span>
                      </div>
                    </div>

                    <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-accent/5 border-accent/20 text-center md:text-left">
                      <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-accent/10 flex items-center justify-center shrink-0">
                        <MapPin size={12} className="text-accent md:w-5 md:h-5" />
                      </div>
                      <div className="min-w-0">
                        <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-accent block mb-0.5 truncate">{dict.shipment.destination || 'Destination'}</span>
                        <span className="text-[8px] md:text-sm font-black text-text-main truncate block">{shippingData.receiverCountry}</span>
                      </div>
                    </div>

                    <div className="glass-panel p-1.5 md:p-4 flex flex-col md:flex-row items-center md:items-center gap-1 md:gap-4 bg-surface/50 text-center md:text-left">
                      <div className="w-6 h-6 md:w-10 md:h-10 rounded-md md:rounded-xl bg-surface/10 flex items-center justify-center shrink-0">
                        <div className="relative flex h-1.5 w-1.5 md:h-3 md:w-3">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-1.5 w-1.5 md:h-3 md:w-3 bg-accent"></span>
                        </div>
                      </div>
                      <div className="min-w-0">
                        <span className="text-[6px] md:text-[10px] font-black uppercase tracking-tight md:tracking-widest text-accent block mb-0.5 truncate">{dict.shipment.live || 'Live'} </span>
                        <span className="text-[8px] md:text-sm font-black text-text-main uppercase tracking-tighter truncate block">
                          {dict.statuses?.[shippingData.status] || shippingData.status.replace(/_/g, ' ')}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="mb-16 relative z-10 animate-fade-in delay-300 w-full max-w-6xl mx-auto shadow-2xl rounded-[2rem] overflow-hidden">
                    <DynamicMap 
                      origin={originCoords} 
                      destination={destCoords} 
                      progress={liveProgress}
                      shipment={shippingData}
                      dict={dict}
                    />
                  </div>
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
                            <div key={idx} className="flex justify-between items-center py-2 md:py-6 border-b border-border last:border-0 group/item">
                              <span className="text-text-muted font-black text-[7px] md:text-sm uppercase tracking-widest">{detail.label}</span>
                              <span className={cn("font-black text-text-main text-[10px] md:text-xl group-hover:text-accent transition-colors", detail.italic && "italic")}>{detail.value}</span>
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
                      <div className="space-y-4 md:space-y-12 relative border-l-2 border-border ml-2 md:ml-3 pl-6 md:pl-10 py-2">
                        {shippingData.timeline ? (
                          shippingData.timeline.map((event, i) => (
                            <motion.div
                              key={i}
                              className="relative group/event"
                              initial={{ opacity: 0, x: -20 }}
                              animate={{ opacity: 1, x: 0 }}
                              transition={{ duration: 0.5, delay: i * 0.15, ease: "easeOut" }}
                            >
                              <div className={cn(
                                "absolute left-[-32px] md:left-[-52px] top-1.5 w-4 h-4 md:w-8 md:h-8 rounded-full border-[3px] md:border-4 border-bg transition-all duration-500",
                                event.is_completed ? "bg-accent scale-125 shadow-lg shadow-accent/40" : "bg-border opacity-50"
                              )}>
                                {/* Pulse ring on the latest active event */}
                                {event.is_completed && !shippingData.timeline?.[i + 1]?.is_completed && (
                                  <motion.div
                                    className="absolute inset-0 rounded-full border-2 border-accent"
                                    animate={{ scale: [1, 2], opacity: [0.6, 0] }}
                                    transition={{ duration: 1.5, repeat: Infinity, ease: "easeOut" }}
                                  />
                                )}
                              </div>
                              <div className="flex flex-wrap items-center gap-2 md:gap-4 mb-2">
                                <p className={cn("font-black text-sm md:text-lg lg:text-xl tracking-tight leading-none uppercase", event.is_completed ? "text-text-main" : "text-text-muted opacity-50")}>
                                  {dict.statuses?.[event.status as keyof typeof dict.statuses] || event.status}
                                </p>
                                {event.is_completed && !shippingData.timeline?.[i + 1]?.is_completed && (
                                  <span className="bg-accent text-white text-[8px] md:text-[9px] font-black uppercase px-2 md:px-2 py-0.5 md:py-1 rounded-lg animate-pulse tracking-wider md:tracking-widest">{dict.shipment.live}</span>
                                )}
                              </div>
                              <motion.span
                                className="text-[7px] md:text-[10px] font-black text-accent/60 uppercase tracking-[0.15em] md:tracking-[0.2em] block mb-4"
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                transition={{ duration: 0.4, delay: i * 0.15 + 0.2 }}
                              >
                                {new Date(event.timestamp).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
                              </motion.span>
                              <motion.div
                                className="mt-4 md:mt-4 flex items-start gap-2 md:gap-4 text-xs md:text-sm font-bold text-text-muted bg-surface-muted/50 p-4 md:p-4 rounded-xl md:rounded-2xl border border-border group-hover/event:border-accent/20 transition-colors"
                                initial={{ opacity: 0, y: 8 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ duration: 0.4, delay: i * 0.15 + 0.3 }}
                              >
                                <Package size={14} className="text-accent/60 mt-0.5 md:w-4 md:h-4 shrink-0" />
                                <span className="leading-relaxed">{event.description}</span>
                              </motion.div>
                              {/* Additional Timeline Detail Layers */}
                              {event.is_completed && (
                                <>
                                  {/* First Detail Layer: Location & Date */}
                                  <motion.div
                                    className="mt-2 md:mt-3 flex items-center gap-2 text-[10px] md:text-xs font-bold text-text-muted/70 bg-surface/30 px-3 py-2 rounded-lg border border-border/50"
                                    initial={{ opacity: 0, y: 6 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ duration: 0.4, delay: i * 0.15 + 0.4 }}
                                  >
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
                                  </motion.div>

                                  {/* Second Detail Layer: Carrier & Handler Info */}
                                  <motion.div
                                    className="mt-2 flex items-center justify-between gap-2 text-[9px] md:text-[10px] text-text-muted/60 bg-surface/20 px-3 py-1.5 rounded-md border border-border/30"
                                    initial={{ opacity: 0, y: 4 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ duration: 0.4, delay: i * 0.15 + 0.5 }}
                                  >
                                    <div className="flex items-center gap-1.5">
                                      <span className="w-1.5 h-1.5 rounded-full bg-accent/60"></span>
                                      <span className="font-bold uppercase tracking-wider">Air Freight</span>
                                    </div>
                                    <span className="opacity-50">|</span>
                                    <div className="flex items-center gap-1">
                                      <span className="font-mono">ID: #{shippingData.trackingNumber.slice(-6)}</span>
                                    </div>
                                  </motion.div>
                                </>
                              )}
                            </motion.div>
                          ))
                        ) : (
                          shippingData.events?.map((event, i: number) => (
                            <div key={event.id} className="relative group/event">
                              <div className={cn(
                                "absolute left-[-33px] md:left-[-51px] top-1.5 w-6 h-6 rounded-full border-4 border-bg transition-all duration-500",
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
                </>
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

export default function Home({ initialId }: HomeProps) {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
      </div>
    }>
      <HomeContent initialId={initialId} />
    </Suspense>
  );
}
