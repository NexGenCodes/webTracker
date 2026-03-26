"use client";

import { useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Polyline, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { ShipmentData, Dictionary } from '@/types/shipment';

// Client-side fix for default Leaflet marker icons in Next.js
if (typeof window !== 'undefined') {
  const DefaultIcon = L.icon({
    iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
    shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
    iconSize: [25, 41],
    iconAnchor: [12, 41],
    popupAnchor: [1, -34],
  });
  L.Marker.prototype.options.icon = DefaultIcon;
}

interface MapProps {
  origin: [number, number];
  destination: [number, number];
  progress?: number;
  shipment?: ShipmentData | null;
  dict?: Dictionary;
}

function MapUpdater({ origin, destination }: { origin: [number, number]; destination: [number, number] }) {
  const map = useMap();
  useEffect(() => {
    const bounds = L.latLngBounds([origin, destination]);
    map.fitBounds(bounds, { padding: [50, 50] });
  }, [map, origin, destination]);
  return null;
}

export default function MapComponent({ origin, destination, progress = 0, shipment, dict }: MapProps) {
  // Interpolate current location based on progress (0-100)
  const currentLat = origin[0] + (destination[0] - origin[0]) * (progress / 100);
  const currentLng = origin[1] + (destination[1] - origin[1]) * (progress / 100);
  const currentLocation: [number, number] | null = progress > 0 && progress < 100
    ? [currentLat, currentLng]
    : null;

  // Calculate the geographic heading to properly pivot the plane marker toward destination
  const deltaLat = destination[0] - origin[0];
  const deltaLng = destination[1] - origin[1];
  const mathAngle = Math.atan2(deltaLat, deltaLng) * (180 / Math.PI);
  const planeRotation = 45 - mathAngle; // Convert from purely geographic to CSS (0deg = plane's native NorthEast pointing)
  return (
    <div className="w-full h-[500px] sm:h-[600px] rounded-[2rem] overflow-hidden glass-panel border border-border/50 relative z-10 shadow-2xl">
      <MapContainer
        center={origin}
        zoom={4}
        scrollWheelZoom={false}
        className="w-full h-full z-0"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          className="map-tiles"
        />

        <Marker position={origin}>
          <Popup className="font-bold text-sm tracking-wider uppercase">Origin Facility</Popup>
        </Marker>

        <Marker position={destination}>
          <Popup className="font-bold text-sm tracking-wider uppercase">Destination</Popup>
        </Marker>

        {currentLocation && (
          <Marker
            position={currentLocation}
            zIndexOffset={1000}
            icon={L.divIcon({
              className: 'live-radar-marker',
              html: `<div class="relative flex flex-col items-center justify-center">
                       <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-plane text-yellow-400 fill-yellow-400 drop-shadow-[0_0_12px_rgba(250,204,21,0.8)] z-10 relative transition-transform duration-1000" style="transform: rotate(${planeRotation}deg);"><path d="M17.8 19.2 16 11l3.5-3.5C21 6 21.5 4 21 3c-1-.5-3 0-4.5 1.5L13 8 4.8 6.2c-.5-.1-.9.2-1.1.6L3 8l6 5.8L6.8 16l-3.2-.8c-.4-.1-.8.2-1 .6L2 17l4.5 1 1 4.5.8-.6c.4-.2.7-.6.6-1l-.8-3.2 2.2-2.2 5.8 6c.4.2.8 0 1-.3l1.2-1c.3-.3.4-.7.3-1.2z"/></svg>
                       <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                         <span class="animate-ping absolute inline-flex h-14 w-14 rounded-full bg-yellow-400 opacity-60"></span>
                         <span class="relative inline-flex rounded-full h-3 w-3 bg-yellow-400 shadow-[0_0_15px_rgba(250,204,21,1)]"></span>
                       </div>
                     </div>`,
              iconSize: [48, 48],
              iconAnchor: [24, 24],
              popupAnchor: [0, -20],
            })}
          >
            <Popup className="font-bold text-xs tracking-wider uppercase text-yellow-500">Live Position</Popup>
          </Marker>
        )}

        <Polyline
          positions={[origin, destination]}
          color="var(--color-accent)"
          weight={4}
          dashArray="10, 15"
          className="opacity-80 drop-shadow-[0_0_8px_rgba(var(--color-accent-rgb),0.5)]"
        />

        <MapUpdater origin={origin} destination={destination} />
      </MapContainer>

      {/* Floating Overlays */}
      <div className="absolute top-4 left-4 sm:top-6 sm:left-6 z-[400] flex flex-col gap-3 pointer-events-none">
        <div className="px-3 py-2 sm:px-5 sm:py-3 rounded-2xl flex items-center gap-3 shadow-[0_4px_20px_rgba(0,0,0,0.1)] backdrop-blur-xl bg-surface/80 border border-border/60">
          <span className="w-2.5 h-2.5 rounded-full bg-text-muted animate-pulse"></span>
          <span className="text-[10px] font-bold uppercase tracking-[0.2em] text-text-muted">{dict?.shipment.from || 'Origin'}</span>
          <span className="text-xs sm:text-sm font-black text-text-main">{shipment?.senderCountry || dict?.shipment.originFacility || 'Origin Facility'}</span>
        </div>
        <div className="px-3 py-2 sm:px-5 sm:py-3 rounded-2xl flex items-center gap-3 shadow-[0_4px_20px_rgba(rgba(var(--color-accent-rgb),0.15))] backdrop-blur-xl bg-accent/10 border border-accent/30">
          <span className="w-2.5 h-2.5 rounded-full bg-accent shadow-[0_0_10px_rgba(var(--color-accent-rgb),1)] animate-pulse"></span>
          <span className="text-[10px] font-bold uppercase tracking-[0.2em] text-accent">{dict?.shipment.destination || 'Dest'}</span>
          <span className="text-xs sm:text-sm font-black text-text-main">{shipment?.receiverCountry || dict?.shipment.destination || 'Destination'}</span>
        </div>
      </div>

      {shipment && (
        <div className="absolute top-4 right-4 sm:top-6 sm:right-6 z-[400] pointer-events-none">
          <div className="px-4 py-3 sm:px-6 sm:py-4 rounded-3xl shadow-[0_10px_30px_rgba(var(--color-accent-rgb),0.2)] backdrop-blur-xl bg-surface/90 border border-accent/40 flex flex-col items-end text-right">
            <span className="text-[9px] sm:text-[10px] font-black uppercase tracking-[0.3em] text-accent mb-1 flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-accent shadow-[0_0_6px_rgba(var(--color-accent-rgb),1)]"></span>
              </span>
              {dict?.shipment.liveStatus || 'Live Status'}
            </span>
            <span className="text-sm sm:text-lg font-black text-text-main capitalize tracking-tight">
              {dict?.admin?.[shipment.status.toLowerCase()] || shipment.status.replace(/_/g, ' ')}
            </span>
          </div>
        </div>
      )}

      <style jsx global>{`
        .leaflet-container {
           background: var(--color-surface-muted);
           font-family: inherit;
        }
        .map-tiles {
           filter: var(--map-filter, none);
        }
        .dark .map-tiles {
           filter: invert(1) hue-rotate(180deg) brightness(95%) contrast(90%);
        }
        .leaflet-popup-content-wrapper {
           background: var(--color-surface);
           color: var(--color-text-main);
           border-radius: 12px;
           border: 1px solid var(--color-border);
           box-shadow: 0 10px 25px rgba(0,0,0,0.1);
        }
        .leaflet-popup-tip {
           background: var(--color-surface);
        }
      `}</style>
    </div>
  );
}
