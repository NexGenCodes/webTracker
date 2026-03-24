import { z } from "zod";

export const shipmentSchema = z.object({
  senderName: z.string().min(2, "Sender name is required"),
  senderCountry: z.string().min(2, "Sender country is required"),
  receiverName: z.string().min(2, "Receiver name is required"),
  receiverPhone: z.string().min(5, "Valid phone number is required"),
  receiverEmail: z.string().email("Invalid email address"),
  receiverAddress: z.string().min(5, "Receiver address is required"),
  receiverCountry: z.string().min(2, "Receiver country is required"),
  cargoType: z.string().optional(),
  weight: z.number().min(0.1, "Weight must be at least 0.1"),
});

export type ShipmentFormData = z.infer<typeof shipmentSchema>;
