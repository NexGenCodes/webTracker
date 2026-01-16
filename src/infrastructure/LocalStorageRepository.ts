import type { Shipment } from '../domain/Shipment';
import type { ShipmentRepository } from '../domain/ShipmentRepository';

const STORAGE_KEY = 'airwaybill_shipments';

export class LocalStorageRepository implements ShipmentRepository {
    async save(shipment: Shipment): Promise<void> {
        const shipments = await this.getAll();
        shipments.push(shipment);
        localStorage.setItem(STORAGE_KEY, JSON.stringify(shipments));
    }

    async findById(id: string): Promise<Shipment | null> {
        const shipments = await this.getAll();
        return shipments.find(s => s.id === id) || null;
    }

    async findByTrackingNumber(trackingNumber: string): Promise<Shipment | null> {
        const shipments = await this.getAll();
        return shipments.find(s => s.trackingNumber === trackingNumber) || null;
    }

    async getAll(): Promise<Shipment[]> {
        const data = localStorage.getItem(STORAGE_KEY);
        return data ? JSON.parse(data) : [];
    }

    async update(shipment: Shipment): Promise<void> {
        const shipments = await this.getAll();
        const index = shipments.findIndex(s => s.id === shipment.id);
        if (index !== -1) {
            shipments[index] = shipment;
            localStorage.setItem(STORAGE_KEY, JSON.stringify(shipments));
        }
    }
}
