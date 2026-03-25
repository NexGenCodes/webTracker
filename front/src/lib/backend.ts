/**
 * Shared backend fetch helper.
 * Constructs the full backend URL and attaches the X-API-Key header.
 */
export function getBackendUrl(): string {
    return process.env.BACKEND_URL || 'http://localhost:5000';
}

export function backendHeaders(extra: Record<string, string> = {}): Record<string, string> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...extra,
    };

    const apiKey = process.env.API_SECRET_KEY;
    if (apiKey) {
        headers['X-API-Key'] = apiKey;
    }

    return headers;
}
