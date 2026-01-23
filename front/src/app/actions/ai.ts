'use server';

import { ParseResult } from '@/types/shipment';
import { logger } from '@/lib/logger';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_TOKEN = process.env.API_AUTH_TOKEN || '';

export async function parseShipmentAI(text: string): Promise<ParseResult> {
    if (!text || text.trim().length < 5) {
        return { success: false, error: 'Please provide more details to parse.' };
    }

    try {
        const response = await fetch(`${API_URL}/api/parse`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${AUTH_TOKEN}`
            },
            body: JSON.stringify({ text }),
        });

        if (!response.ok) {
            const errData = await response.json().catch(() => ({}));
            return { success: false, error: errData.error || `Server error: ${response.statusText}` };
        }

        const data = await response.json();
        return { success: true, data };

    } catch (error: any) {
        logger.error('AI Parsing Bridge Error', error);
        return {
            success: false,
            error: 'Failed to connect to parsing engine.',
        };
    }
}
