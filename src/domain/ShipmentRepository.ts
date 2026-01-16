import type { Shipment } from '../domain/Shipment';

export interface ShipmentRepository {
    save(shipment: Shipment): Promise<void>;
    findById(id: string): Promise<Shipment | null>;
    findByTrackingNumber(trackingNumber: string): Promise<Shipment | null>;
    getAll(): Promise<Shipment[]>;
    update(shipment: Shipment): Promise<void>;
}
