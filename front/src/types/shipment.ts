export type ShipmentStatus = 'PENDING' | 'IN_TRANSIT' | 'OUT_FOR_DELIVERY' | 'DELIVERED' | 'CANCELED';

export interface ShipmentEvent {
    id: string;
    timestamp: string | Date;
    status: ShipmentStatus;
    location: string;
    notes?: string | null;
}

export interface TimelineEvent {
    status: string;
    timestamp: string;
    description: string;
    is_completed: boolean;
}

export interface ShipmentData {
    id: string;
    trackingNumber: string;
    status: ShipmentStatus;
    senderName: string | null;
    senderCountry?: string | null;
    cargoType?: string | null;
    receiverName: string | null;
    receiverAddress: string | null;
    receiverCountry: string | null;
    receiverPhone: string | null;
    weight?: number;
    receiverEmail: string | null;
    isArchived: boolean;
    events: ShipmentEvent[];
    originCoords?: [number, number];
    destinationCoords?: [number, number];
    createdAt?: string | Date;
    scheduledTransitTime?: string | Date;
    outfordeliveryTime?: string | Date;
    expectedDeliveryTime?: string | Date;
    estimatedDelivery?: string | Date;
    timeline?: TimelineEvent[];
    // Added for WhatsApp context in services
    whatsappMessageId?: string | null;
    whatsappFrom?: string | null;
}

export interface CreateShipmentDto {
    receiverName: string;
    receiverAddress: string;
    receiverCountry: string;
    receiverPhone: string;
    receiverEmail: string;
    senderName: string;
    senderCountry: string;
    weight?: number;
    cargoType?: string;
}

export interface ServiceResult<T = unknown> {
    success: boolean;
    data?: T;
    error?: string;
    count?: number;
}

export interface DashboardStats {
    total: number;
    inTransit: number;
    outForDelivery: number;
    delivered: number;
    pending: number;
    canceled: number;
}

export interface Pagination {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
}

export interface PaginatedResult<T> {
    data: T[];
    pagination: Pagination;
}

export interface Dictionary {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [key: string]: any;
}

export interface ParseResult {
    success: boolean;
    data?: Partial<CreateShipmentDto>;
    error?: string;
    correction?: string;
}

