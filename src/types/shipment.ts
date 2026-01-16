export type ShipmentStatus = 'PENDING' | 'IN_TRANSIT' | 'OUT_FOR_DELIVERY' | 'DELIVERED' | 'CANCELLED';

export interface ShipmentEvent {
    id: string;
    timestamp: string | Date;
    status: ShipmentStatus;
    location: string;
    notes?: string | null;
}

export interface ShipmentData {
    id: string;
    trackingNumber: string;
    status: ShipmentStatus;
    senderName: string | null;
    receiverName: string | null;
    receiverAddress: string | null;
    receiverCountry: string | null;
    receiverPhone: string | null;
    isArchived: boolean;
    events: ShipmentEvent[];
}
