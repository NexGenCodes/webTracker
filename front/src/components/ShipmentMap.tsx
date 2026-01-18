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

    // Custom icons - Clean FedEx-style markers
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

    const movingIcon = typeof window !== 'undefined' ? L.divIcon({
        className: 'custom-moving-icon',
        html: `<div class="fedex-moving-marker"></div>`,
        iconSize: [16, 16],
        iconAnchor: [8, 8]
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
        <div className="relative w-full space-y-4">
            {/* Clean Progress Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white dark:bg-surface-muted p-5 rounded-2xl border border-border shadow-sm">
                <div className="space-y-2">
                    <p className="text-xs font-semibold text-text-muted uppercase tracking-wide">{dict.map.orbitalTracking}</p>
                    <div className="flex items-center gap-3">
                        <div className="h-2 w-48 bg-gray-200 dark:bg-surface-muted rounded-full overflow-hidden">
                            <div
                                className="h-full bg-accent transition-all duration-1000 ease-out"
                                style={{ width: `${animatedProgress * 100}%` }}
                            />
                        </div>
                        <span className="text-sm font-bold text-accent">
                            {Math.round(animatedProgress * 100)}%
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

            {/* Map Container */}
            <div className="h-[300px] md:h-[450px] w-full rounded-2xl overflow-hidden border border-border shadow-lg relative bg-white">
                {/* Simple status badge */}
                <div className="absolute top-4 left-4 z-10">
                    <div className="bg-white/95 dark:bg-surface/95 backdrop-blur-sm px-3 py-1.5 rounded-lg border border-border shadow-sm flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-success" />
                        <span className="text-xs font-semibold text-text-main uppercase tracking-wide">{dict.map.liveTelemetry}</span>
                    </div>
                </div>

                <MapContainer
                    center={mapCenter}
                    zoom={zoom}
                    style={{ height: '100%', width: '100%' }}
                    zoomControl={false}
                    attributionControl={false}
                >
                    {/* Light, clean map tiles (FedEx style) */}
                    <TileLayer
                        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                    />

                    {/* Simple route line */}
                    <Polyline
                        positions={arcPath.slice(0, Math.floor(animatedProgress * (arcPath.length - 1)) + 1)}
                        pathOptions={{
                            color: 'hsl(var(--accent))',
                            weight: 4,
                            opacity: 0.7,
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

            {/* Origin and Destination Labels */}
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
        </div>
    );
};

export default ShipmentMap;
