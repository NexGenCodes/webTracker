import { NextRequest, NextResponse } from 'next/server';
import pdf from 'pdf-parse';

/**
 * Server-side API route for PDF text extraction.
 * This runs in a Node.js environment, which natively supports 
 * modules like 'fs' and 'path', avoiding the browser build errors.
 */
export async function POST(req: NextRequest) {
    try {
        const formData = await req.formData();
        const file = formData.get('file') as Blob;
        
        if (!file) {
            return NextResponse.json({ error: 'No file provided' }, { status: 400 });
        }

        const arrayBuffer = await file.arrayBuffer();
        const buffer = Buffer.from(arrayBuffer);
        
        const data = await pdf(buffer);
        
        // Return extracted text to be processed by AI
        return NextResponse.json({ text: data.text });
    } catch (error: unknown) {
        console.error('PDF Server-Side Parse Error:', error);
        return NextResponse.json({ 
            error: error instanceof Error ? error.message : 'Failed to parse PDF on server' 
        }, { status: 500 });
    }
}
