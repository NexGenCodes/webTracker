import { z } from 'zod';

export const ShipmentStatusSchema = z.enum([
    'PENDING',
    'PICKED_UP',
    'IN_TRANSIT',
    'ARRIVED_IN_COUNTRY',
    'OUT_FOR_DELIVERY',
    'DELIVERED'
]);

export type ShipmentStatus = z.infer<typeof ShipmentStatusSchema>;

export const TrackingEventSchema = z.object({
    id: z.string(),
    status: ShipmentStatusSchema,
    location: z.string(),
    timestamp: z.string(),
    notes: z.string().optional()
});

export type TrackingEvent = z.infer<typeof TrackingEventSchema>;

export const ShipmentSchema = z.object({
    id: z.string(),
    trackingNumber: z.string(), // The HASH ID
    senderName: z.string().min(1, "Sender Name is required"),
    receiverName: z.string().min(1, "Receiver Name is required"),
    receiverEmail: z.string().email().optional(), // In case we add it later
    receiverAddress: z.string().min(1, "Address is required"),
    receiverCountry: z.string().min(1, "Country is required"),
    receiverPhone: z.string().min(1, "Phone is required"),
    status: ShipmentStatusSchema,
    events: z.array(TrackingEventSchema),
    createdAt: z.string(),
    updatedAt: z.string()
});

export type Shipment = z.infer<typeof ShipmentSchema>;

// Input DTO for creating a shipment from parsed data
export const CreateShipmentDtoSchema = ShipmentSchema.pick({
    senderName: true,
    receiverName: true,
    receiverAddress: true,
    receiverCountry: true,
    receiverPhone: true
});

export type CreateShipmentDto = z.infer<typeof CreateShipmentDtoSchema>;
