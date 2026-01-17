export type ShipmentStatus = 'PENDING' | 'IN_TRANSIT' | 'OUT_FOR_DELIVERY' | 'DELIVERED' | 'CANCELED';

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
    senderCountry?: string | null;
    receiverName: string | null;
    receiverAddress: string | null;
    receiverCountry: string | null;
    receiverPhone: string | null;
    isArchived: boolean;
    events: ShipmentEvent[];
    originCoords?: [number, number];
    destinationCoords?: [number, number];
    estimatedDelivery?: string | Date;
    // Added for WhatsApp context in services
    whatsappMessageId?: string | null;
    whatsappFrom?: string | null;
}

export interface CreateShipmentDto {
    receiverName: string;
    receiverAddress: string;
    receiverCountry: string;
    receiverPhone: string;
    senderName: string;
    senderCountry: string;
}

export interface ServiceResult<T = any> {
    success: boolean;
    data?: T;
    error?: string;
    count?: number;
}

export interface ParseResult {
    success: boolean;
    data?: CreateShipmentDto;
    error?: string;
    correction?: string;
}

