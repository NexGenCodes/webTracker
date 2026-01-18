"use client";

import React, { useEffect, useState, useMemo } from 'react';
import { MapContainer, TileLayer, Marker, Polyline, useMap } from 'react-leaflet';
import { useTheme } from 'next-themes';
import { useI18n } from './I18nContext';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import { ShipmentData } from '@/types/shipment';

// Leaflet SSR fix
if (typeof window !== 'undefined') {
    delete (L.Icon.Default.prototype as any)._getIconUrl;
    L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
        iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
        shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
    });
}

interface ShipmentMapProps {
    shipmentData: ShipmentData;
}

// --- Icons SVGs ---
const TruckIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white drop-shadow-md"><rect x="1" y="3" width="15" height="13"></rect><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"></polygon><circle cx="5.5" cy="18.5" r="2.5"></circle><circle cx="18.5" cy="18.5" r="2.5"></circle></svg>`;
const PlaneIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white drop-shadow-md"><path d="M2 12h20"></path><path d="M13 2l5.36 8.04"></path><path d="M6 12l2.36 7.09"></path><path d="M19 12l-4-8"></path><path d="M16 22l-4-8"></path></svg>`; // Simplified plane

// Helper to generate a curved path (arc) between two points
const generateArc = (start: [number, number], end: [number, number], steps: number = 100): [number, number][] => {
    const points: [number, number][] = [];
    const distance = Math.sqrt(Math.pow(end[0] - start[0], 2) + Math.pow(end[1] - start[1], 2));
    const arcHeight = distance * 0.15; // Slightly shallower arc for realism

    for (let i = 0; i <= steps; i++) {
        const t = i / steps;
        let lat = start[0] + (end[0] - start[0]) * t;
        let lng = start[1] + (end[1] - start[1]) * t;
        lat += Math.sin(t * Math.PI) * arcHeight;
        points.push([lat, lng]);
    }
    return points;
};

// Calculate bearing between two points
const calculateBearing = (start: [number, number], end: [number, number]): number => {
    const startLat = start[0] * Math.PI / 180;
    const startLng = start[1] * Math.PI / 180;
    const endLat = end[0] * Math.PI / 180;
    const endLng = end[1] * Math.PI / 180;

    const y = Math.sin(endLng - startLng) * Math.cos(endLat);
    const x = Math.cos(startLat) * Math.sin(endLat) -
        Math.sin(startLat) * Math.cos(endLat) * Math.cos(endLng - startLng);
    const θ = Math.atan2(y, x);
    const brng = (θ * 180 / Math.PI + 360) % 360; // in degrees
    return brng;
};

const getShipmentProgress = (status: string, isArchived: boolean): number => {
    if (isArchived || status === 'DELIVERED') return 1;
    if (status === 'OUT_FOR_DELIVERY') return 0.85;
    if (status === 'IN_TRANSIT') return 0.55;
    return 0;
};

const MapUpdater: React.FC<{ center: [number, number]; zoom: number }> = ({ center, zoom }) => {
    const map = useMap();
    useEffect(() => {
        map.setView(center, zoom);
    }, [center, zoom, map]);
    return null;
};

const ShipmentMap: React.FC<ShipmentMapProps> = ({ shipmentData }) => {
    const { dict } = useI18n();
    const { resolvedTheme } = useTheme();
    const [isMounted, setIsMounted] = useState(false);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [progress, setProgress] = useState(0);

    const origin = shipmentData.originCoords || [20.0, 0.0];
    const destination = shipmentData.destinationCoords || [40.0, -74.0];
    const arcPath = useMemo(() => generateArc(origin, destination, 200), [origin, destination]);

    // Derived values
    const distance = Math.sqrt(Math.pow(destination[0] - origin[0], 2) + Math.pow(destination[1] - origin[1], 2));
    const isLongJorney = distance > 50; // Threshold for plane
    // @ts-ignore
    const VehicleSVG = isLongJorney ? PlaneIconSVG : TruckIconSVG;

    const isActive = shipmentData.status === 'IN_TRANSIT' || shipmentData.status === 'OUT_FOR_DELIVERY';

    useEffect(() => {
        setIsMounted(true);
        const calculateProgress = () => {
            if (shipmentData.status === 'PENDING') return 0;
            if (shipmentData.status === 'DELIVERED') return 1;
            if (shipmentData.isArchived) return 1;
            if (shipmentData.status === 'CANCELED') return 0;

            if (!shipmentData.createdAt || !shipmentData.estimatedDelivery) {
                return 0.1;
            }

            const start = new Date(shipmentData.createdAt).getTime();
            const end = new Date(shipmentData.estimatedDelivery).getTime();
            const now = new Date().getTime();

            if (end <= start) return 0.5;
            const percentage = (now - start) / (end - start);
            return Math.max(0.05, Math.min(percentage, 0.95));
        };

        setProgress(calculateProgress());
        const interval = setInterval(() => setProgress(calculateProgress()), 60000);
        return () => clearInterval(interval);
    }, [shipmentData]);

    const pathIndex = Math.floor(progress * (arcPath.length - 1));
    const currentPosition = arcPath[pathIndex];
    const nextPosition = arcPath[Math.min(pathIndex + 1, arcPath.length - 1)];
    const bearing = calculateBearing(currentPosition, nextPosition);

    const traveledPath = arcPath.slice(0, pathIndex + 1);
    const remainingPath = arcPath.slice(pathIndex);

    const mapCenter: [number, number] = [
        (origin[0] + destination[0]) / 2 + 5,
        (origin[1] + destination[1]) / 2
    ];
    const zoom = distance > 100 ? 2 : distance > 50 ? 3 : 4;


    const tileLayerUrl = resolvedTheme === 'dark'
        ? 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png'
        : 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png';

    const rotationAdjustment = isLongJorney ? 45 : 0;

    // --- Custom Leaflet Icons ---
    const originIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-origin-icon',
        html: `<div class="fedex-origin-marker"></div>`,
        iconSize: [20, 20],
        iconAnchor: [10, 10]
    }) : null;

    const destinationIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-destination-icon',
        html: `<div class="fedex-destination-marker"></div>`,
        iconSize: [20, 20],
        iconAnchor: [10, 10]
    }) : null;

    const vehicleIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-vehicle-icon',
        html: `
            <div style="transform: rotate(${bearing - 45 + rotationAdjustment}deg);" class="transition-transform duration-500 ease-linear">
                <div class="p-1.5 bg-accent rounded-full shadow-lg border-2 border-white relative z-50">
                    <div class="w-5 h-5 text-white flex items-center justify-center">
                        ${VehicleSVG}
                    </div>
                </div>
                <div class="absolute inset-0 bg-accent rounded-full animate-ping opacity-50 z-0"></div>
            </div>
        `,
        iconSize: [32, 32],
        iconAnchor: [16, 16]
    }) : null;


    if (!isMounted) return <div className="h-[300px] w-full bg-surface-muted animate-pulse rounded-2xl" />;

    const toggleFullscreen = () => setIsFullscreen(!isFullscreen);
    const estimatedDate = shipmentData.estimatedDelivery
        ? new Date(shipmentData.estimatedDelivery).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
        : 'N/A';

    return (
        <div className={`relative transition-all duration-500 ease-in-out ${isFullscreen ? 'fixed inset-0 z-50 h-screen w-screen bg-slate-900/90 backdrop-blur-sm p-4' : 'w-full space-y-4'}`}>

            {!isFullscreen && (
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white dark:bg-surface-muted p-5 rounded-2xl border border-border shadow-sm">
                    <div className="space-y-2">
                        <p className="text-xs font-semibold text-text-muted uppercase tracking-wide">{dict.map.orbitalTracking}</p>
                        <div className="flex items-center gap-3">
                            <div className="h-2 w-48 bg-gray-200 dark:bg-surface-muted rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-accent transition-all duration-1000 ease-out"
                                    style={{ width: `${progress * 100}%` }}
                                />
                            </div>
                            <span className="text-sm font-bold text-accent">
                                {Math.round(progress * 100)}%
                            </span>
                        </div>
                    </div>
                    <div className="flex gap-6">
                        <div className="text-right">
                            <p className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-1">{dict.map.statusLabel}</p>
                            <p className="text-sm font-bold text-text-main">{shipmentData.status}</p>
                        </div>
                        <div className="text-right">
                            <p className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-1">{dict.map.etaWindow}</p>
                            <p className="text-sm font-bold text-accent">{estimatedDate}</p>
                        </div>
                    </div>
                </div>
            )}

            <div className={`
                ${isFullscreen
                    ? 'h-full w-full rounded-2xl shadow-2xl border-2 border-accent/20'
                    : 'h-[300px] md:h-[450px] w-full rounded-2xl border border-border shadow-lg'
                } 
                relative overflow-hidden bg-white dark:bg-gray-900 transition-all duration-500
            `}>
                <button
                    onClick={toggleFullscreen}
                    className="absolute top-4 right-4 z-10 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 p-2 rounded-lg shadow-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors border border-border"
                    title={isFullscreen ? "Exit Fullscreen" : "Enter Fullscreen"}
                >
                    {isFullscreen ? (
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M8 3v3a2 2 0 0 1-2 2H3" /><path d="M21 8h-3a2 2 0 0 1-2-2V3" /><path d="M3 16h3a2 2 0 0 1 2 2v3" /><path d="M16 21v-3a2 2 0 0 1 2-2h3" /></svg>
                    ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M15 3h6v6" /><path d="M9 21H3v-6" /><path d="M21 3l-7 7" /><path d="M3 21l7-7" /></svg>
                    )}
                </button>

                <div className="absolute top-4 left-4 z-10 flex flex-col gap-2">
                    <div className="bg-white/90 dark:bg-gray-900/90 backdrop-blur-md px-4 py-2 rounded-xl border border-white/20 shadow-lg flex items-center gap-3">
                        <div className="relative">
                            <div className={`w-3 h-3 rounded-full ${isActive ? 'bg-success shadow-[0_0_10px_rgba(34,197,94,0.5)]' : 'bg-gray-400'}`} />
                            {isActive && <div className="absolute inset-0 rounded-full bg-success animate-ping opacity-75"></div>}
                        </div>
                        <div>
                            <p className="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-widest leading-none mb-0.5">Telemetry</p>
                            <p className="text-xs font-bold text-gray-800 dark:text-gray-100 uppercase tracking-wide leading-none">{isActive ? dict.map.liveTelemetry : 'Offline'}</p>
                        </div>
                    </div>
                </div>

                {isFullscreen && (
                    <div className="absolute bottom-8 left-8 z-10 w-64 bg-slate-900/80 backdrop-blur-xl border border-white/10 p-4 rounded-2xl text-white shadow-2xl">
                        <div className="flex justify-between items-end mb-4">
                            <div>
                                <p className="text-xs text-slate-400 uppercase tracking-wider font-semibold mb-1">Package Status</p>
                                <p className="text-lg font-bold text-white leading-none">{shipmentData.status.replace(/_/g, ' ')}</p>
                            </div>
                            <div className="text-right">
                                <p className="text-xs text-slate-400 uppercase tracking-wider font-semibold mb-1">Speed</p>
                                <p className="text-lg font-bold text-accent leading-none">
                                    {isActive ? (isLongJorney ? '~860 km/h' : '~95 km/h') : '0 km/h'}
                                </p>
                            </div>
                        </div>
                        <div className="h-1.5 w-full bg-slate-700 rounded-full overflow-hidden mb-2">
                            <div className="h-full bg-linear-to-r from-blue-500 to-accent" style={{ width: `${progress * 100}%` }} />
                        </div>
                        <div className="flex justify-between text-[10px] text-slate-400 font-mono uppercase">
                            <span>{Math.round(bearing)}° HEAD</span>
                            <span>{Math.round(progress * 100)}% Complete</span>
                        </div>
                    </div>
                )}


                <MapContainer
                    center={mapCenter}
                    zoom={zoom}
                    style={{ height: '100%', width: '100%' }}
                    zoomControl={false}
                    attributionControl={false}
                >
                    <TileLayer url={tileLayerUrl} />

                    {/* Base Track (The full planned route) - Muted & Dashed */}
                    <Polyline
                        positions={arcPath}
                        pathOptions={{
                            color: resolvedTheme === 'dark' ? '#334155' : '#64748b', // Slate-700 (dark) vs Slate-500 (light)
                            weight: 3,
                            opacity: 0.8,
                            dashArray: '6, 8',
                            lineCap: 'round'
                        }}
                    />

                    {/* Progress Track (Traveled distance) - Solid Accent */}
                    <Polyline
                        positions={traveledPath}
                        pathOptions={{
                            color: 'hsl(var(--accent))',
                            weight: 4,
                            opacity: 1,
                            lineCap: 'round',
                            // shadowBlur: 10, // Not a standard Leaflet Polyline option
                            // shadowColor: 'hsl(var(--accent))' // Not a standard Leaflet Polyline option
                        }}
                    />

                    {originIcon && <Marker position={origin} icon={originIcon} />}
                    {destinationIcon && <Marker position={destination} icon={destinationIcon} />}

                    {vehicleIcon && isActive && progress > 0 && progress < 1 && (
                        <Marker position={currentPosition} icon={vehicleIcon} />
                    )}

                    <MapUpdater center={mapCenter} zoom={zoom} />
                </MapContainer>
            </div>

            {!isFullscreen && (
                <div className="grid grid-cols-2 gap-4">
                    <div className="bg-white dark:bg-surface-muted p-4 rounded-xl border border-border">
                        <p className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-1">{dict.map.signalOrigin}</p>
                        <p className="text-sm font-bold text-text-main truncate">{shipmentData.senderCountry || 'Unknown'}</p>
                    </div>
                    <div className="bg-white dark:bg-surface-muted p-4 rounded-xl border border-border text-right">
                        <p className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-1">{dict.map.signalTarget}</p>
                        <p className="text-sm font-bold text-text-main truncate">{shipmentData.receiverCountry || 'Unknown'}</p>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ShipmentMap;
