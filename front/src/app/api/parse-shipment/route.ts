import { NextResponse } from 'next/server';
import { GoogleGenerativeAI } from '@google/generative-ai';
import { logger } from '@/lib/logger';

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
/* ... rest of the prompt ... */
`;

    logger.info('API Parsing request received');
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
    logger.error('Gemini parsing API error', error);
    return NextResponse.json({
      error: 'Failed to parse message',
      details: error instanceof Error ? error.message : 'Unknown error'
    }, { status: 500 });
  }
}
