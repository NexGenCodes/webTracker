"use client";

import React, { useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

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
    locationName: string;
}


const MOCK_COORDS: Record<string, [number, number]> = {
    'USA': [37.0902, -95.7129],
    'China': [35.8617, 104.1954],
    'Germany': [51.1657, 10.4515],
    'Origin': [25.0, -40.0],
    'Destination': [40.7128, -74.0060], // NY
    'Default': [20.0, 0.0]
};

const MapUpdater: React.FC<{ coords: [number, number] }> = ({ coords }) => {
    const map = useMap();
    useEffect(() => {
        map.flyTo(coords, 5);
    }, [coords, map]);
    return null;
};

const ShipmentMap: React.FC<ShipmentMapProps> = ({ locationName }) => {
    const coordKey = Object.keys(MOCK_COORDS).find(k => locationName?.includes(k)) || 'Default';
    const coords = MOCK_COORDS[coordKey];

    return (
        <div className="h-[300px] md:h-[400px] w-full rounded-xl overflow-hidden border border-gray-700/50 mt-8 relative z-0">
            <MapContainer center={coords} zoom={4} style={{ height: '100%', width: '100%' }}>
                <TileLayer
                    url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
                    attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
                />
                <Marker position={coords}>
                    <Popup>
                        Current Location: <br /> {locationName}
                    </Popup>
                </Marker>
                <MapUpdater coords={coords} />
            </MapContainer>
        </div>
    );
};

export default ShipmentMap;
