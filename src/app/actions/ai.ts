'use server';

import { GoogleGenerativeAI } from '@google/generative-ai';
import { CreateShipmentDto } from '@/types/shipment';

const genAI = new GoogleGenerativeAI(process.env.GEMINI_API_KEY || '');

export interface ParseResult {
    success: boolean;
    data?: CreateShipmentDto;
    error?: string;
    correction?: string;
}

export async function parseShipmentAI(text: string): Promise<ParseResult> {
    if (!process.env.GEMINI_API_KEY) {
        return { success: false, error: 'AI capabilities are not configured (Missing API Key)' };
    }

    if (!text || text.trim().length < 10) {
        return { success: false, error: 'Please provide more details to parse.' };
    }

    try {
        const model = genAI.getGenerativeModel({
            model: "gemini-2.0-flash-exp",
            generationConfig: {
                temperature: 0.1,
                responseMimeType: "application/json"
            }
        });

        const systemPrompt = `You are a logistics data extraction assistant. Extract shipping information from user text and return JSON matching the schema below.
        
        TARGET SCHEMA:
        {
            "receiverName": string,
            "receiverAddress": string,
            "receiverCountry": string,
            "receiverPhone": string,
            "senderName": string,
            "senderCountry": string
        }

        RULES:
        1. Extract the fields from the input text.
        2. If a field is missing, use an empty string "" - DO NOT return null.
        3. Infer countries if city names are well-known (e.g. "Paris" -> "France").
        4. Phone numbers: Extract as is.
        `;

        const result = await model.generateContent([
            { text: systemPrompt },
            { text: `Extract shipping data from this:\n\n${text}` }
        ]);

        const responseText = result.response.text();
        const parsed: CreateShipmentDto = JSON.parse(responseText);

        // Basic validation - check if we got at least some data
        const hasData = Object.values(parsed).some(v => v && v.length > 0);

        if (!hasData) {
            return { success: false, error: 'Could not extract valid shipping information.' };
        }

        return { success: true, data: parsed };

    } catch (error: any) {
        console.error('AI Parsing Error:', error);
        return {
            success: false,
            error: 'Failed to process with AI. Please fill manually.',
            correction: error.message
        };
    }
}
