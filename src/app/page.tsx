"use client";

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { TrackingSearch } from '@/components/TrackingSearch';
import { getTracking } from './actions/shipment';
import { Package, Truck, CheckCircle, MapPin, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

// Dynamic import for Map to avoid SSR issues
const ShipmentMap = dynamic(() => import('@/components/ShipmentMap'), {
  ssr: false,
  loading: () => <div className="h-[300px] w-full bg-gray-900/50 animate-pulse rounded-xl mt-8"></div>
});

export default function Home() {
  const [loading, setLoading] = useState(false);
  const [shippingData, setShippingData] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async (trackingNumber: string) => {
    setLoading(true);
    setError(null);
    setShippingData(null);

    try {
      const data = await getTracking(trackingNumber);
      if (data) {
        setShippingData(data);
      } else {
        setError("Shipment not found. Please check your tracking number.");
      }
    } catch (err) {
      setError("System error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex flex-col items-center p-4 pb-20 overflow-x-hidden">
      <div className="w-full max-w-5xl flex flex-col flex-1">

        {/* Header/Nav Area */}
        <header className="py-6 flex justify-between items-center mb-8 md:mb-12">
          <div className="flex items-center gap-2 font-bold text-xl text-gradient">
            <Package className="text-blue-400" />
            WebTracker
          </div>
        </header>

        {/* Search Component */}
        <div className={cn("transition-all duration-500", shippingData ? "mb-12" : "mt-20 md:mt-32")}>
          <TrackingSearch onSearch={handleSearch} isLoading={loading} />
        </div>

        {/* Error Message */}
        {error && (
          <div className="w-full max-w-xl mx-auto p-4 bg-red-500/10 border border-red-500/50 rounded-xl flex items-center gap-3 text-red-300 animate-fade-in">
            <AlertCircle />
            {error}
          </div>
        )}

        {/* Results View */}
        {shippingData && (
          <div className="w-full animate-fade-in space-y-8">

            {/* Status Card */}
            <div className="glass-panel p-6 md:p-8">
              <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-white/5 pb-6 mb-6">
                <div>
                  <p className="text-gray-400 text-sm uppercase tracking-wider mb-1">Current Status</p>
                  <h2 className={cn("text-3xl md:text-4xl font-bold",
                    shippingData.status === 'DELIVERED' ? "text-green-400" : "text-blue-400"
                  )}>
                    {shippingData.status.replace(/_/g, ' ')}
                  </h2>
                </div>
                <div className="text-left md:text-right">
                  <p className="text-gray-400 text-sm uppercase tracking-wider mb-1">Destination</p>
                  <h3 className="text-xl font-semibold text-white">
                    {shippingData.isArchived ? "Hidden (Delivered)" : shippingData.receiverCountry}
                  </h3>
                </div>
              </div>

              {/* If Archived/Delivered, show limited view */}
              {shippingData.isArchived ? (
                <div className="flex flex-col items-center justify-center p-8 text-center space-y-4">
                  <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center">
                    <CheckCircle className="w-10 h-10 text-green-500" />
                  </div>
                  <div>
                    <h3 className="text-2xl font-bold text-white mb-2">Package Delivered</h3>
                    <p className="text-gray-400 max-w-md mx-auto">
                      This shipment has been successfully delivered.
                      In accordance with our privacy policy, personal details have been anonymized.
                    </p>
                  </div>
                </div>
              ) : (
                /* Active Tracking View */
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 md:gap-12">

                  {/* Details Column */}
                  <div className="space-y-6">
                    <div>
                      <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
                        <MapPin className="text-blue-400" size={18} />
                        Shipment Details
                      </h4>
                      <div className="bg-gray-900/40 rounded-xl p-4 space-y-3 text-sm border border-white/5">
                        <div className="flex justify-between">
                          <span className="text-gray-500">Tracking ID</span>
                          <span className="font-mono text-white">{shippingData.trackingNumber}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-500">Receiver</span>
                          <span className="text-white">{shippingData.receiverName}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-500">From</span>
                          <span className="text-white">{shippingData.senderName}</span>
                        </div>
                      </div>
                    </div>

                    {/* Timeline */}
                    <div>
                      <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
                        <Truck className="text-blue-400" size={18} />
                        Latest Updates
                      </h4>
                      <div className="space-y-0 relative border-l-2 border-gray-800 ml-3 pl-6 py-2">
                        {shippingData.events?.map((event: any, i: number) => (
                          <div key={event.id} className="relative mb-8 last:mb-0">
                            <div className={cn(
                              "absolute left-[-31px] top-1 w-4 h-4 rounded-full border-2",
                              i === 0 ? "bg-blue-500 border-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" : "bg-gray-900 border-gray-600"
                            )}></div>
                            <p className="text-white font-medium text-lg">{event.status.replace(/_/g, ' ')}</p>
                            <span className="text-xs text-blue-300 bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20">
                              {new Date(event.timestamp).toLocaleString()}
                            </span>
                            <p className="text-sm text-gray-400 mt-2">{event.location}</p>
                            {event.notes && <p className="text-sm text-gray-500 mt-1 italic">"{event.notes}"</p>}
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Map Column */}
                  <div>
                    <h4 className="text-white font-semibold mb-4">Live Location</h4>
                    <ShipmentMap locationName={shippingData.events?.[0]?.location || 'Default'} />
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
