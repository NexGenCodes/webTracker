'use server';

import { GoogleGenerativeAI } from '@google/generative-ai';
import { CreateShipmentDto, ParseResult } from '@/types/shipment';
import { logger } from '@/lib/logger';

const genAI = new GoogleGenerativeAI(process.env.GEMINI_API_KEY || '');

// --- Regex Handling ---
const rxReceiverName = /(?:Receiver|Receivers?)\s*Name:\s*(.*)/i;
const rxReceiverAddress = /(?:Receiver|Receivers?)\s*Address:\s*(.*)/i;
const rxReceiverPhone = /(?:Receiver|Receivers?)\s*Phone:\s*(.*)/i;
const rxReceiverCountry = /(?:Receiver|Receivers?)\s*Country:\s*(.*)|Destination:\s*(.*)/i;
const rxSenderName = /(?:Sender|Senders?)\s*Name:\s*(.*)|Sender:\s*(.*)/i;
const rxSenderCountry = /(?:Sender|Senders?)\s*Country:\s*(.*)|Origin:\s*(.*)/i;

function extract(rx: RegExp, text: string): string {
    const match = text.match(rx);
    if (match && match.length > 1) {
        // Return first non-undefined, non-empty capture group (skipping the full match at [0])
        for (let i = 1; i < match.length; i++) {
            if (match[i] && match[i].trim()) return match[i].trim();
        }
    }
    return "";
}

function parseWithRegex(text: string): CreateShipmentDto {
    return {
        receiverName: extract(rxReceiverName, text),
        receiverAddress: extract(rxReceiverAddress, text),
        receiverPhone: extract(rxReceiverPhone, text),
        receiverCountry: extract(rxReceiverCountry, text),
        senderName: extract(rxSenderName, text),
        senderCountry: extract(rxSenderCountry, text)
    };
}

export async function parseShipmentAI(text: string): Promise<ParseResult> {
    if (!text || text.trim().length < 5) {
        return { success: false, error: 'Please provide more details to parse.' };
    }

    // 1. Try Regex First (Fast & Free)
    logger.info("Starting extraction", { textLength: text.length });
    const regexData = parseWithRegex(text);

    const missingFields = [];
    if (!regexData.receiverName) missingFields.push('receiverName');
    if (!regexData.receiverAddress) missingFields.push('receiverAddress');
    if (!regexData.receiverPhone) missingFields.push('receiverPhone');
    if (!regexData.receiverCountry) missingFields.push('receiverCountry');
    if (!regexData.senderName) missingFields.push('senderName');
    if (!regexData.senderCountry) missingFields.push('senderCountry');

    if (missingFields.length === 0) {
        // Perfect match!
        return { success: true, data: regexData };
    }

    // 2. AI Fallback (Smart but slower)
    if (!process.env.GEMINI_API_KEY) {
        // If no API key, return what we have from Regex
        return {
            success: Object.values(regexData).some(v => !!v),
            data: regexData,
            error: missingFields.length > 0 ? `Missing: ${missingFields.join(', ')}` : undefined
        };
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
        const aiData: CreateShipmentDto = JSON.parse(responseText);

        // 3. Merge Results (Prefer Regex for known matches, use AI for gaps/typos)
        // Actually, AI is better at handling typos, so if Regex failed significantly, AI might be better overall.
        // But let's fill gaps strictly to preserve "Hybrid" nature.    
        const mergedData = { ...regexData };
        if (!mergedData.receiverName) mergedData.receiverName = aiData.receiverName;
        if (!mergedData.receiverAddress) mergedData.receiverAddress = aiData.receiverAddress;
        if (!mergedData.receiverPhone) mergedData.receiverPhone = aiData.receiverPhone;
        if (!mergedData.receiverCountry) mergedData.receiverCountry = aiData.receiverCountry;
        if (!mergedData.senderName) mergedData.senderName = aiData.senderName;
        if (!mergedData.senderCountry) mergedData.senderCountry = aiData.senderCountry;

        const hasData = Object.values(mergedData).some(v => v && v.length > 0);
        if (!hasData) {
            logger.warn('AI extraction returned no data', { text });
            return { success: false, error: 'Could not extract valid shipping information.' };
        }

        logger.info('AI extraction successful');
        return { success: true, data: mergedData };

    } catch (error: any) {
        logger.error('AI Parsing Error', error);
        // Fallback to whatever regex found if AI fails
        if (Object.values(regexData).some(v => !!v)) {
            return { success: true, data: regexData, error: 'AI enhancement failed, returned raw extraction.' };
        }
        return {
            success: false,
            error: 'Failed to process with AI. Please fill manually.',
            correction: error.message
        };
    }
}
