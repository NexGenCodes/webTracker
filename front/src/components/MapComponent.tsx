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
  // Quadratic Bezier Curve calculation for flight path
  const getCurvePoint = (t: number, p0: [number, number], p1: [number, number], p2: [number, number]): [number, number] => {
    const lat = Math.pow(1 - t, 2) * p0[0] + 2 * (1 - t) * t * p1[0] + Math.pow(t, 2) * p2[0];
    const lng = Math.pow(1 - t, 2) * p0[1] + 2 * (1 - t) * t * p1[1] + Math.pow(t, 2) * p2[1];
    return [lat, lng];
  };

  // Determine control point for the arch
  const midX = (origin[0] + destination[0]) / 2;
  const midY = (origin[1] + destination[1]) / 2;
  const dx = destination[0] - origin[0];
  const dy = destination[1] - origin[1];
  const dist = Math.sqrt(dx * dx + dy * dy);
  
  // Offset control point perpendicularly to the line
  const offsetScale = 0.15; // Lower for subtle arch
  const p1: [number, number] = [midX - dy * offsetScale, midY + dx * offsetScale];

  // Generate curve points for Polyline
  const curvePoints: [number, number][] = [];
  const steps = 50;
  for (let i = 0; i <= steps; i++) {
    curvePoints.push(getCurvePoint(i / steps, origin, p1, destination));
  }

  // Interpolate current location along the CURVE based on progress
  const t = Math.max(0, Math.min(1, progress / 100)); // Clamp between 0 and 1
  const currentPos = getCurvePoint(t, origin, p1, destination);
  
  // Calculate orientation based on a small delta
  // If at the very end, look backwards to maintain arrival heading
  const delta = 0.01;
  const refPos = t < 0.99 
    ? getCurvePoint(t + delta, origin, p1, destination)
    : currentPos;
  const prevPos = t >= 0.99
    ? getCurvePoint(t - delta, origin, p1, destination)
    : currentPos;

  const angleRad = t < 0.99
    ? Math.atan2(refPos[0] - currentPos[0], refPos[1] - currentPos[1])
    : Math.atan2(currentPos[0] - prevPos[0], currentPos[1] - prevPos[1]);

  const mathAngle = angleRad * (180 / Math.PI);
  const planeRotation = 45 - mathAngle;

  // Plane visibility: Always visible once started, or if arrived
  const currentLocation: [number, number] | null = progress >= 0 ? currentPos : null;
  return (
    <div className="w-full h-[600px] md:h-[750px] rounded-[2rem] overflow-hidden glass-panel border border-border/50 relative z-10 shadow-2xl">
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
                       <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-plane text-yellow-400 fill-yellow-400 drop-shadow-[0_0_12px_rgba(250,204,21,0.8)] z-10 relative transition-transform duration-1000" style="transform: rotate(${planeRotation}deg);"><path d="M17.8 19.2 16 11l3.5-3.5C21 6 21.5 4 21 3c-1-.5-3 0-4.5 1.5L13 8 4.8 6.2c-.5-.1-.9.2-1.1.6L3 8l6 5.8L6.8 16l-3.2-.8c-.4-.1-.8.2-1 .6L2 17l4.5 1 1 4.5.8-.6c.4-.2.7-.6.6-1l-.8-3.2 2.2-2.2 5.8 6.c.4.2.8 0 1-.3l1.2-1c.3-.3.4-.7.3-1.2z"/></svg>
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
          positions={curvePoints}
          color="var(--color-accent)"
          weight={4}
          dashArray="10, 15"
          className="opacity-80 drop-shadow-[0_0_8px_rgba(var(--color-accent-rgb),0.5)]"
        />

        <MapUpdater origin={origin} destination={destination} />
      </MapContainer>

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
