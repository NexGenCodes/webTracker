'use server';

import { ParseResult } from '@/types/shipment';
import { logger } from '@/lib/logger';

function getNextJsBaseUrl() {
  if (typeof window !== 'undefined') return '';
  if (process.env.API_URL) return process.env.API_URL;
  if (process.env.VERCEL_URL) return `https://${process.env.VERCEL_URL}`;
  return `http://localhost:${process.env.PORT ?? 3000}`;
}

export async function parseShipmentAI(text: string): Promise<ParseResult> {
    if (!text || text.trim().length < 5) {
        return { success: false, error: 'Please provide more details to parse.' };
    }

    try {
        // Uses the local Next.js API route (no VPS dependency)
        const baseUrl = getNextJsBaseUrl();
        const response = await fetch(`${baseUrl}/api/parse-shipment`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ text }),
        });

        if (!response.ok) {
            const errData = await response.json().catch(() => ({}));
            return { success: false, error: errData.error || `Server error: ${response.statusText}` };
        }

        const data = await response.json();
        return { success: true, data };

    } catch (error: unknown) {
        logger.error('AI Parsing Bridge Error', error);
        return {
            success: false,
            error: 'Failed to connect to parsing engine.',
        };
    }
}
