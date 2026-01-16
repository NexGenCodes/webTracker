import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Shipment } from '../domain/Shipment';

interface ShipmentStore {
    shipments: Shipment[];
    saveShipment: (shipment: Shipment) => void;
    updateShipment: (shipment: Shipment) => void;
    getShipment: (id: string) => Shipment | undefined;
}

export const useShipmentStore = create<ShipmentStore>()(
    persist(
        (set, get) => ({
            shipments: [],
            saveShipment: (shipment) => set((state) => {
                const exists = state.shipments.some(s => s.trackingNumber === shipment.trackingNumber);
                if (exists) return state; // Prevent duplicates if any
                return { shipments: [...state.shipments, shipment] };
            }),
            updateShipment: (updated) => set((state) => ({
                shipments: state.shipments.map(s =>
                    s.trackingNumber === updated.trackingNumber ? updated : s
                )
            })),
            getShipment: (id) => get().shipments.find(s => s.trackingNumber === id),
        }),
        {
            name: 'shipment-storage', // name of the item in the storage (must be unique)
        }
    )
);
