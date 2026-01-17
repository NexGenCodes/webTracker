import { NextResponse } from 'next/server';
import { GoogleGenerativeAI } from '@google/generative-ai';

const genAI = new GoogleGenerativeAI(process.env.GEMINI_API_KEY!);

export async function POST(request: Request) {
    try {
        const { message } = await request.json();

        if (!message) {
            return NextResponse.json({ error: 'Message is required' }, { status: 400 });
        }

        const model = genAI.getGenerativeModel({
            model: "gemini-2.0-flash-exp",
            generationConfig: {
                temperature: 0.1,
                responseMimeType: "application/json"
            }
        });

        const systemPrompt = `You are a logistics data extraction assistant. Extract shipping information from user messages and return ONLY valid JSON.

REQUIRED FIELDS:
- Sender: name, country
- Receiver: name, phone, address, country

RULES:
1. If ALL required fields are present, set status to "SUCCESS"
2. If ANY field is missing, set status to "INCOMPLETE" and provide a helpful "correction" message listing the missing fields
3. Phone numbers should include country code if possible
4. Always return valid JSON matching this exact schema

RESPONSE SCHEMA:
{
  "status": "SUCCESS" | "INCOMPLETE",
  "sender": {
    "name": "string or null",
    "country": "string or null"
  },
  "receiver": {
    "name": "string or null",
    "phone": "string or null",
    "address": "string or null",
    "country": "string or null"
  },
  "correction": "string or null (only if INCOMPLETE)"
}

EXAMPLES:

Input: "Send package from John in USA to Mary at +234123456789, 123 Lagos St, Nigeria"
Output: {
  "status": "SUCCESS",
  "sender": { "name": "John", "country": "USA" },
  "receiver": { "name": "Mary", "phone": "+234123456789", "address": "123 Lagos St", "country": "Nigeria" },
  "correction": null
}

Input: "Ship from Jane in UK to Bob in Nigeria"
Output: {
  "status": "INCOMPLETE",
  "sender": { "name": "Jane", "country": "UK" },
  "receiver": { "name": "Bob", "phone": null, "address": null, "country": "Nigeria" },
  "correction": "Missing required information: receiver's phone number and address. Please provide the complete delivery address and contact number."
}`;

        const result = await model.generateContent([
            { text: systemPrompt },
            { text: `Extract shipping data from this message:\n\n${message}` }
        ]);

        const responseText = result.response.text();
        const parsedData = JSON.parse(responseText);

        // Validate the response structure
        if (!parsedData.status || !parsedData.sender || !parsedData.receiver) {
            throw new Error('Invalid response structure from AI');
        }

        return NextResponse.json(parsedData);

    } catch (error) {
        console.error('Gemini parsing error:', error);
        return NextResponse.json({
            error: 'Failed to parse message',
            details: error instanceof Error ? error.message : 'Unknown error'
        }, { status: 500 });
    }
}
