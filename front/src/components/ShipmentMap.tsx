"use client";

import React, { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, Polyline, useMap } from 'react-leaflet';
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

// Helper to generate a curved path (arc) between two points
const generateArc = (start: [number, number], end: [number, number], steps: number = 100): [number, number][] => {
    const points: [number, number][] = [];
    const midLat = (start[0] + end[0]) / 2;
    const midLng = (start[1] + end[1]) / 2;

    // Calculate distance to determine arc height
    const distance = Math.sqrt(Math.pow(end[0] - start[0], 2) + Math.pow(end[1] - start[1], 2));
    const arcHeight = distance * 0.2; // 20% of distance as height

    for (let i = 0; i <= steps; i++) {
        const t = i / steps;
        // Basic coordinates
        let lat = start[0] + (end[0] - start[0]) * t;
        let lng = start[1] + (end[1] - start[1]) * t;

        // Add arc height (offset latitude based on a parabola)
        // This is a simple visual arc, not true geodesic but looks great
        lat += Math.sin(t * Math.PI) * arcHeight;

        points.push([lat, lng]);
    }
    return points;
};

const getShipmentProgress = (status: string, isArchived: boolean): number => {
    if (isArchived || status === 'DELIVERED') return 1;
    if (status === 'OUT_FOR_DELIVERY') return 0.85;
    if (status === 'IN_TRANSIT') return 0.55;
    if (status === 'PENDING') return 0.15;
    return 0.5;
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
    const [isMounted, setIsMounted] = useState(false);
    const [animatedProgress, setAnimatedProgress] = useState(0);

    const origin = shipmentData.originCoords || [20.0, 0.0];
    const destination = shipmentData.destinationCoords || [40.0, -74.0];
    const targetProgress = getShipmentProgress(shipmentData.status, shipmentData.isArchived);

    // Generate the full arc path
    const arcPath = generateArc(origin, destination);

    // Calculate current position along the arc based on progress
    const getCurrentPosition = (progress: number): [number, number] => {
        const index = Math.floor(progress * (arcPath.length - 1));
        return arcPath[index];
    };

    // Animate marker position
    useEffect(() => {
        setIsMounted(true);

        const duration = 2500; // Smoother 2.5s animation
        const steps = 120; // More steps for smoothness
        const stepDuration = duration / steps;
        let currentStep = 0;

        const interval = setInterval(() => {
            currentStep++;
            const easedT = 1 - Math.pow(1 - (currentStep / steps), 3); // Ease out cubic
            const progress = Math.min(easedT * targetProgress, targetProgress);
            setAnimatedProgress(progress);

            if (currentStep >= steps) {
                clearInterval(interval);
            }
        }, stepDuration);

        return () => clearInterval(interval);
    }, [targetProgress]);

    // Custom icons
    const originIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-origin-icon',
        html: `<div class="origin-marker"></div>`,
        iconSize: [24, 24],
        iconAnchor: [12, 12]
    }) : null;

    const destinationIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-destination-icon',
        html: `<div class="destination-marker"></div>`,
        iconSize: [24, 24],
        iconAnchor: [12, 12]
    }) : null;

    const movingIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-moving-icon',
        html: `
            <div class="moving-marker-glow">
                <div class="moving-marker-core"></div>
                <div class="moving-marker-trail"></div>
            </div>
        `,
        iconSize: [32, 32],
        iconAnchor: [16, 16]
    }) : null;

    const currentPosition = getCurrentPosition(animatedProgress);
    const mapCenter: [number, number] = [
        (origin[0] + destination[0]) / 2 + (targetProgress > 0 ? 5 : 0), // Slight offset for arc view
        (origin[1] + destination[1]) / 2
    ];

    const distance = Math.sqrt(
        Math.pow(destination[0] - origin[0], 2) + Math.pow(destination[1] - origin[1], 2)
    );
    const zoom = distance > 80 ? 2 : distance > 40 ? 3 : 4;

    if (!isMounted) {
        return (
            <div className="h-[300px] md:h-[450px] w-full rounded-3xl overflow-hidden border border-border bg-surface-muted animate-pulse" />
        );
    }

    const estimatedDate = shipmentData.estimatedDelivery
        ? new Date(shipmentData.estimatedDelivery).toLocaleDateString(undefined, {
            month: 'short',
            day: 'numeric',
            year: 'numeric'
        })
        : 'N/A';

    return (
        <div className="relative w-full space-y-6">
            {/* Header Telemetry */}
            <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 glass-panel p-6 border-accent/20">
                <div className="space-y-1">
                    <p className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] mb-1">{dict.map.orbitalTracking}</p>
                    <div className="flex items-center gap-3">
                        <div className="h-2 w-48 bg-surface-muted rounded-full overflow-hidden border border-border/50 shadow-inner">
                            <div
                                className="h-full bg-linear-to-r from-accent via-primary to-accent bg-size-[200%_100%] animate-pulse transition-all duration-1000 ease-out"
                                style={{ width: `${animatedProgress * 100}%` }}
                            />
                        </div>
                        <span className="text-sm font-black text-accent font-mono">
                            {Math.round(animatedProgress * 100)}%
                        </span>
                    </div>
                </div>
                <div className="flex gap-8">
                    <div className="text-right">
                        <p className="text-[10px] font-black text-text-muted uppercase tracking-widest leading-none mb-1">{dict.map.statusLabel}</p>
                        <p className="text-xs font-black text-text-main uppercase">{shipmentData.isArchived ? dict.map.signalTerminated : dict.map.uplinkActive}</p>
                    </div>
                    <div className="text-right">
                        <p className="text-[10px] font-black text-text-muted uppercase tracking-widest leading-none mb-1">{dict.map.etaWindow}</p>
                        <p className="text-xs font-black text-accent uppercase">{estimatedDate}</p>
                    </div>
                </div>
            </div>

            {/* Map Container */}
            <div className="h-[300px] md:h-[450px] w-full rounded-3xl overflow-hidden border border-border/50 shadow-2xl relative group">
                {/* Visual Noise Overlay */}
                <div className="absolute inset-0 pointer-events-none bg-radial-gradient from-transparent via-transparent to-bg/20 z-10" />
                <div className="absolute top-6 left-6 z-10 flex gap-2">
                    <div className="bg-bg/80 backdrop-blur-md px-3 py-1.5 rounded-lg border border-border flex items-center gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-success animate-ping" />
                        <span className="text-[10px] font-black text-text-main uppercase tracking-widest">{dict.map.liveTelemetry}</span>
                    </div>
                </div>

                <MapContainer
                    center={mapCenter}
                    zoom={zoom}
                    style={{ height: '100%', width: '100%' }}
                    zoomControl={false}
                    attributionControl={false}
                    className="grayscale-[0.2] contrast-[1.1] brightness-[0.8]"
                >
                    <TileLayer
                        url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
                    />

                    {/* Background Static Route */}
                    <Polyline
                        positions={arcPath}
                        pathOptions={{
                            color: 'hsl(var(--accent))',
                            weight: 1,
                            opacity: 0.2,
                            dashArray: '4, 8',
                        }}
                    />

                    {/* Active Pulsing Route */}
                    <Polyline
                        positions={arcPath.slice(0, Math.floor(animatedProgress * (arcPath.length - 1)) + 1)}
                        pathOptions={{
                            color: 'hsl(var(--accent))',
                            weight: 3,
                            opacity: 0.8,
                            lineCap: 'round',
                        }}
                    />

                    {/* Origin Marker */}
                    {originIcon && (
                        <Marker position={origin} icon={originIcon} />
                    )}

                    {/* Destination Marker */}
                    {destinationIcon && (
                        <Marker position={destination} icon={destinationIcon} />
                    )}

                    {/* Moving Progress Indicator */}
                    {movingIcon && (
                        <Marker position={currentPosition} icon={movingIcon} />
                    )}

                    <MapUpdater center={mapCenter} zoom={zoom} />
                </MapContainer>
            </div>

            {/* Geographic Coordinates */}
            <div className="grid grid-cols-2 gap-4">
                <div className="glass-panel p-4 bg-primary/5 border-primary/20">
                    <p className="text-[10px] font-black text-primary uppercase tracking-widest mb-1">{dict.map.signalOrigin}</p>
                    <p className="text-sm font-black text-text-main truncate uppercase">{shipmentData.senderCountry || dict.map.deepSpace}</p>
                </div>
                <div className="glass-panel p-4 bg-accent/5 border-accent/20 text-right">
                    <p className="text-[10px] font-black text-accent uppercase tracking-widest mb-1">{dict.map.signalTarget}</p>
                    <p className="text-sm font-black text-text-main truncate uppercase">{shipmentData.receiverCountry || dict.map.undisclosed}</p>
                </div>
            </div>
        </div>
    );
};

export default ShipmentMap;
