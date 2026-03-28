'use server';

import { ParseResult } from '@/types/shipment';
import { logger } from '@/lib/logger';
import { ShipmentService } from '@/services/shipment.service';
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

export async function parseShipmentAI(text: string): Promise<ParseResult> {
    if (!text || text.trim().length < 5) {
        return { success: false, error: 'Please provide more details to parse.' };
    }
    
    // Protect this route
    const session = await getServerSession(authOptions);
    if (!session) return { success: false, error: 'Unauthorized' };

    try {
        const result = await ShipmentService.parseText(text);
        if (!result.success || !result.data) {
            return { success: false, error: result.error || 'Failed to parse text.' };
        }

        const d = result.data as Record<string, unknown>;
        
        // Map the Go PascalCase struct strings back to standard frontend camelCase
        const mappedData = {
            receiverName: String(d.ReceiverName || d.receiverName || ''),
            receiverAddress: String(d.ReceiverAddress || d.receiverAddress || ''),
            receiverCountry: String(d.ReceiverCountry || d.receiverCountry || ''),
            receiverPhone: String(d.ReceiverPhone || d.receiverPhone || ''),
            receiverEmail: String(d.ReceiverEmail || d.receiverEmail || ''),
            senderName: String(d.SenderName || d.senderName || ''),
            senderCountry: String(d.SenderCountry || d.senderCountry || ''),
            cargoType: String(d.CargoType || d.cargoType || 'consignment box'),
            weight: typeof d.Weight === 'number' ? d.Weight : 15.0
        };

        return { success: true, data: mappedData };

    } catch (error: unknown) {
        logger.error('AI Parsing Bridge Error', error);
        return {
            success: false,
            error: 'Failed to connect to backend parsing engine.',
        };
    }
}
