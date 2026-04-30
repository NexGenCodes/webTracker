/**
 * Shared backend fetch helper.
 * Constructs the full backend URL and forwards the JWT cookie as a Bearer token.
 */
export function getBackendUrl(): string {
    return process.env.BACKEND_URL || 'http://localhost:5000';
}

import { cookies } from 'next/headers';

export async function backendHeaders(extra: Record<string, string> = {}): Promise<Record<string, string>> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...extra,
    };

    try {
        const cookieStore = await cookies();
        const jwt = cookieStore.get('jwt')?.value;
        if (jwt) {
            headers['Authorization'] = `Bearer ${jwt}`;
        }
    } catch {
        // Fallback for edge cases outside Next.js request context
    }

    return headers;
}
