import { CreateShipmentDto, ShipmentData, ServiceResult } from '@/types/shipment';
import { logger } from '@/lib/logger';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_TOKEN = process.env.API_AUTH_TOKEN || '';
const REQUEST_TIMEOUT = 10000; // 10 seconds

// Enhanced error categorization
enum ApiErrorType {
    NETWORK = 'NETWORK',
    TIMEOUT = 'TIMEOUT',
    SERVER = 'SERVER',
    NOT_FOUND = 'NOT_FOUND',
    UNAUTHORIZED = 'UNAUTHORIZED',
    VALIDATION = 'VALIDATION',
    UNKNOWN = 'UNKNOWN'
}

interface ApiError {
    type: ApiErrorType;
    message: string;
    userMessage: string;
}

/**
 * Enhanced fetch with timeout and better error handling
 */
async function fetchWithTimeout(url: string, options: RequestInit = {}): Promise<Response> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT);

    try {
        const response = await fetch(url, {
            ...options,
            signal: controller.signal
        });
        clearTimeout(timeoutId);
        return response;
    } catch (error: any) {
        clearTimeout(timeoutId);
        if (error.name === 'AbortError') {
            throw new Error('Request timeout');
        }
        throw error;
    }
}

/**
 * Categorize and format errors for better user feedback
 */
function handleApiError(error: any, context: string): ApiError {
    logger.error(`[ShipmentService] ${context}`, error);

    // Network/Connection errors
    if (error.message?.includes('fetch failed') || error.message?.includes('Failed to fetch')) {
        return {
            type: ApiErrorType.NETWORK,
            message: error.message,
            userMessage: 'Cannot connect to server. Please check your internet connection or try again later.'
        };
    }

    // Timeout errors
    if (error.message?.includes('timeout') || error.name === 'AbortError') {
        return {
            type: ApiErrorType.TIMEOUT,
            message: 'Request timed out',
            userMessage: 'Request took too long. The server might be busy, please try again.'
        };
    }

    // Server errors (5xx)
    if (error.message?.includes('500') || error.message?.includes('502') || error.message?.includes('503')) {
        return {
            type: ApiErrorType.SERVER,
            message: error.message,
            userMessage: 'Server error. Our team has been notified. Please try again in a few minutes.'
        };
    }

    // Unauthorized (401/403)
    if (error.message?.includes('401') || error.message?.includes('403') || error.message?.includes('Unauthorized')) {
        return {
            type: ApiErrorType.UNAUTHORIZED,
            message: error.message,
            userMessage: 'Session expired. Please sign in again.'
        };
    }

    // Not found (404)
    if (error.message?.includes('404')) {
        return {
            type: ApiErrorType.NOT_FOUND,
            message: error.message,
            userMessage: 'The requested resource was not found.'
        };
    }

    // Validation errors (400)
    if (error.message?.includes('400') || error.message?.includes('Bad Request')) {
        return {
            type: ApiErrorType.VALIDATION,
            message: error.message,
            userMessage: 'Invalid data provided. Please check your input and try again.'
        };
    }

    // Unknown/Generic errors
    return {
        type: ApiErrorType.UNKNOWN,
        message: error.message || 'Unknown error',
        userMessage: 'Something went wrong. Please try again or contact support if the issue persists.'
    };
}

export class ShipmentService {
    /**
     * Create a new shipment via Go API
     */
    static async create(data: CreateShipmentDto): Promise<ServiceResult<{ trackingNumber: string }>> {
        try {
            const response = await fetchWithTimeout(`${API_URL}/api/shipments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                },
                body: JSON.stringify(data),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            const result = await response.json();
            return { success: true, data: { trackingNumber: result.tracking_id } };
        } catch (error) {
            const apiError = handleApiError(error, 'Create shipment');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Fetch tracking details from Go Backend API
     */
    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        try {
            const response = await fetchWithTimeout(`${API_URL}/api/track/${trackingNumber}`, {
                next: { revalidate: 0 }
            });

            if (!response.ok) {
                if (response.status === 404) return null;
                throw new Error(`API error: ${response.statusText} (${response.status})`);
            }

            const data = await response.json();

            // Map backend simple status to frontend typed status
            const normalizeStatus = (s: string): string => {
                const upper = s.toUpperCase();
                if (upper === 'INTRANSIT') return 'IN_TRANSIT';
                if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
                if (upper === 'CANCELLED') return 'CANCELED';
                return upper;
            };

            const shipment: ShipmentData = {
                id: data.tracking_id,
                trackingNumber: data.tracking_id,
                status: normalizeStatus(data.status) as any,
                senderName: data.sender_name || 'N/A',
                receiverName: data.recipient_name || 'N/A',
                receiverPhone: data.recipient_phone || null,
                receiverEmail: data.recipient_email || null,
                receiverAddress: data.recipient_address || null,
                receiverCountry: data.recipient_country || data.destination || 'N/A',
                weight: data.weight || 0,
                senderCountry: data.origin || 'N/A',
                timeline: data.timeline || [],
                events: [],
                isArchived: data.status === 'delivered',
            };
            return shipment;
        } catch (error) {
            const apiError = handleApiError(error, 'Fetch tracking');
            // For tracking, we return null but still log the detailed error
            return null;
        }
    }

    /**
     * Admin: Update status via Go API
     */
    static async updateStatus(trackingNumber: string, status: string, location: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetchWithTimeout(`${API_URL}/api/shipments/${trackingNumber}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                },
                body: JSON.stringify({ status: status.toLowerCase(), destination: location }),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            return { success: true };
        } catch (error) {
            const apiError = handleApiError(error, 'Update status');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Mark as delivered
     */
    static async markDelivered(trackingNumber: string): Promise<ServiceResult<void>> {
        return this.updateStatus(trackingNumber, 'delivered', 'Destination');
    }

    /**
     * Admin: Delete shipment via Go API
     */
    static async delete(trackingNumber: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetchWithTimeout(`${API_URL}/api/shipments/${trackingNumber}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                }
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            return { success: true };
        } catch (error) {
            const apiError = handleApiError(error, 'Delete shipment');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Dashboard data via Go API
     */
    static async getDashboardData(): Promise<ServiceResult<{ shipments: any[], stats: any }>> {
        try {
            const headers = { 'Authorization': `Bearer ${AUTH_TOKEN}` };

            // Fetch List
            const listRes = await fetchWithTimeout(`${API_URL}/api/shipments`, {
                headers,
                next: { revalidate: 0 }
            });

            if (!listRes.ok) {
                throw new Error(`Failed to fetch shipments: ${listRes.status}`);
            }

            const apiShipments = await listRes.json();

            const normalizeStatus = (s: string): string => {
                const upper = s.toUpperCase();
                if (upper === 'INTRANSIT') return 'IN_TRANSIT';
                if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
                if (upper === 'CANCELLED') return 'CANCELED';
                return upper;
            };

            // Map to frontend expected keys (camelCase)
            const shipments = apiShipments.map((s: any) => ({
                id: s.tracking_id,
                trackingNumber: s.tracking_id,
                status: normalizeStatus(s.status),
                senderName: s.sender_name,
                receiverName: s.recipient_name,
                receiverPhone: s.recipient_phone,
                receiverEmail: s.recipient_email,
                receiverAddress: s.recipient_address,
                receiverCountry: s.destination,
                weight: s.weight,
                senderCountry: s.origin,
                createdAt: s.created_at,
                isArchived: s.status === 'delivered',
            }));

            // Fetch Stats
            const statsRes = await fetchWithTimeout(`${API_URL}/api/stats`, {
                headers,
                next: { revalidate: 0 }
            });

            if (!statsRes.ok) {
                throw new Error(`Failed to fetch stats: ${statsRes.status}`);
            }

            const apiStats = await statsRes.json();

            const stats = {
                total: shipments.length,
                inTransit: shipments.filter((s: any) => s.status === 'IN_TRANSIT').length,
                outForDelivery: shipments.filter((s: any) => s.status === 'OUT_FOR_DELIVERY').length,
                delivered: shipments.filter((s: any) => s.status === 'DELIVERED').length,
                pending: shipments.filter((s: any) => s.status === 'PENDING').length,
                canceled: shipments.filter((s: any) => s.status === 'CANCELED').length,
            };

            return { success: true, data: { shipments, stats } };
        } catch (error) {
            const apiError = handleApiError(error, 'Dashboard data');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Bulk cleanup of delivered shipments
     */
    static async bulkDeleteDelivered(): Promise<ServiceResult<void>> {
        try {
            const response = await fetchWithTimeout(`${API_URL}/api/shipments/cleanup`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${AUTH_TOKEN}`
                }
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            return { success: true };
        } catch (error) {
            const apiError = handleApiError(error, 'Bulk delete');
            return { success: false, error: apiError.userMessage };
        }
    }

    static async pruneStale(): Promise<void> { }
}
