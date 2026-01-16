import type { Shipment } from '../domain/Shipment';
import type { ShipmentRepository } from '../domain/ShipmentRepository';
import { useShipmentStore } from './store';

export class ZustandRepository implements ShipmentRepository {
    async save(shipment: Shipment): Promise<void> {
        const { saveShipment } = useShipmentStore.getState();
        saveShipment(shipment);
        return Promise.resolve();
    }

    async findByTrackingNumber(trackingNumber: string): Promise<Shipment | null> {
        const { getShipment } = useShipmentStore.getState();
        const shipment = getShipment(trackingNumber);
        return Promise.resolve(shipment || null);
    }

    async findById(id: string): Promise<Shipment | null> {
        // For this app, trackingNumber is functionally the ID, or we search by underlying UUID if needed.
        // Assuming findByTrackingNumber is the primary lookup.
        // If we need strict ID lookup:
        const { shipments } = useShipmentStore.getState();
        const shipment = shipments.find(s => s.id === id);
        return Promise.resolve(shipment || null);
    }

    async getAll(): Promise<Shipment[]> {
        const { shipments } = useShipmentStore.getState();
        return Promise.resolve(shipments);
    }

    async update(shipment: Shipment): Promise<void> {
        const { updateShipment } = useShipmentStore.getState();
        updateShipment(shipment);
        return Promise.resolve();
    }
}
