import type { Shipment, CreateShipmentDto, TrackingEvent } from '../domain/Shipment';
import type { ShipmentRepository } from '../domain/ShipmentRepository';

export class ShipmentService {
    private repository: ShipmentRepository;

    constructor(repository: ShipmentRepository) {
        this.repository = repository;
    }

    async createShipment(dto: CreateShipmentDto): Promise<Shipment> {
        const id = crypto.randomUUID();
        const trackingNumber = this.generateTrackingNumber();
        const now = new Date().toISOString();

        const shipment: Shipment = {
            id,
            trackingNumber,
            ...dto,
            status: 'PENDING',
            events: [
                {
                    id: crypto.randomUUID(),
                    status: 'PENDING',
                    location: 'Origin Processing Center',
                    timestamp: now,
                    notes: 'Shipment created'
                }
            ],
            createdAt: now,
            updatedAt: now
        };

        await this.repository.save(shipment);
        return shipment;
    }

    async getTracking(trackingNumber: string): Promise<Shipment | null> {
        return this.repository.findByTrackingNumber(trackingNumber);
    }

    async addEvent(shipmentId: string, location: string, status?: any, notes?: string): Promise<Shipment> {
        const shipment = await this.repository.findById(shipmentId);
        if (!shipment) throw new Error("Shipment not found");

        const now = new Date().toISOString();

        // Auto-Arrival Logic
        let newStatus = status || shipment.status;
        if (location.trim().toLowerCase() === shipment.receiverCountry.toLowerCase()) {
            newStatus = 'ARRIVED_IN_COUNTRY';
        }

        const event: TrackingEvent = {
            id: crypto.randomUUID(),
            status: newStatus,
            location,
            timestamp: now,
            notes
        };

        shipment.events.push(event);
        shipment.status = newStatus;
        shipment.updatedAt = now;

        // Delivery Cleanup Logic
        if (newStatus === 'DELIVERED') {
            this.anonymizeShipment(shipment);
        }

        await this.repository.update(shipment);
        return shipment;
    }

    private generateTrackingNumber(): string {
        return 'AWB-' + Math.random().toString(36).substring(2, 9).toUpperCase();
    }

    private anonymizeShipment(shipment: Shipment) {
        shipment.senderName = "GDPR-REDACTED";
        shipment.receiverName = "GDPR-REDACTED";
        shipment.receiverAddress = "REDACTED";
        shipment.receiverPhone = "REDACTED";
        // Keep Country and Tracking Number
    }
}
