import { CreateShipmentDto, ShipmentData, ServiceResult, ShipmentStatus, DashboardStats } from '@/types/shipment';
import { logger } from '@/lib/logger';
import { getBaseUrl } from '@/lib/utils';


const REQUEST_TIMEOUT = 10000;

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

    // Ensure absolute URL for server-side fetches
    const baseUrl = getBaseUrl();
    const fullUrl = url.startsWith('/') ? `${baseUrl}${url}` : url;

    try {
        const response = await fetch(fullUrl, {
            ...options,
            signal: controller.signal
        });
        clearTimeout(timeoutId);
        return response;
    } catch (error: unknown) {
        clearTimeout(timeoutId);
        if (error instanceof Error && error.name === 'AbortError') {
            throw new Error('Request timeout');
        }
        throw error;
    }
}

/**
 * Categorize and format errors for better user feedback
 */
function handleApiError(error: unknown, context: string): ApiError {
    const message = error instanceof Error ? error.message : String(error);
    const name = error instanceof Error ? error.name : '';

    logger.error(`[ShipmentService] ${context}`, error);

    // Network/Connection errors
    if (message.includes('fetch failed') || message.includes('Failed to fetch')) {
        return {
            type: ApiErrorType.NETWORK,
            message: message,
            userMessage: 'Cannot connect to server. Please check your internet connection or try again later.'
        };
    }

    // Timeout errors
    if (message.includes('timeout') || name === 'AbortError') {
        return {
            type: ApiErrorType.TIMEOUT,
            message: 'Request timed out',
            userMessage: 'Request took too long. The server might be busy, please try again.'
        };
    }

    // Server errors (5xx)
    if (message.includes('500') || message.includes('502') || message.includes('503')) {
        return {
            type: ApiErrorType.SERVER,
            message: message,
            userMessage: 'Server error. Our team has been notified. Please try again in a few minutes.'
        };
    }

    // Unauthorized (401/403)
    if (message.includes('401') || message.includes('403') || message.includes('Unauthorized')) {
        return {
            type: ApiErrorType.UNAUTHORIZED,
            message: message,
            userMessage: 'Session expired. Please sign in again.'
        };
    }

    // Not found (404)
    if (message.includes('404')) {
        return {
            type: ApiErrorType.NOT_FOUND,
            message: message,
            userMessage: 'The requested resource was not found.'
        };
    }

    // Validation errors (400)
    if (message.includes('400') || message.includes('Bad Request')) {
        return {
            type: ApiErrorType.VALIDATION,
            message: message,
            userMessage: 'Invalid data provided. Please check your input and try again.'
        };
    }

    return {
        type: ApiErrorType.UNKNOWN,
        message: message || 'Unknown error',
        userMessage: 'Something went wrong. Please try again or contact support if the issue persists.'
    };
}

const normalizeStatus = (s: string): string => {
    const upper = s.toUpperCase();
    if (upper === 'INTRANSIT') return 'IN_TRANSIT';
    if (upper === 'OUTFORDELIVERY') return 'OUT_FOR_DELIVERY';
    if (upper === 'CANCELLED') return 'CANCELED';
    return upper;
};

export class ShipmentService {
    /**
     * Create a new shipment via Next.js API
     */
    static async create(data: CreateShipmentDto): Promise<ServiceResult<{ trackingNumber: string }>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
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
     * Fetch tracking details via Next.js API
     */
    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        try {
            const response = await fetchWithTimeout(`/api/track/${trackingNumber}`, {
                next: { revalidate: 0 }
            });

            if (!response.ok) {
                if (response.status === 404) return null;
                throw new Error(`API error: ${response.statusText} (${response.status})`);
            }

            const data = await response.json();

            // Normalize status
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
                status: normalizeStatus(data.status as string) as ShipmentStatus,
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
            handleApiError(error, 'Fetch tracking');
            return null;
        }
    }

    /**
     * Admin: Update status via Next.js API
     */
    static async updateStatus(trackingNumber: string, status: string, location: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/${trackingNumber}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json'
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
     * Admin: Delete shipment via Next.js API
     */
    static async delete(trackingNumber: string): Promise<ServiceResult<void>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/${trackingNumber}`, {
                method: 'DELETE'
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
     * Admin: Dashboard data via Next.js API
     */
    static async getDashboardData(): Promise<ServiceResult<{ shipments: ShipmentData[], stats: DashboardStats }>> {
        try {
            // Fetch List
            const listRes = await fetchWithTimeout(`/api/admin/shipments`, {
                next: { revalidate: 0 }
            });

            if (!listRes.ok) {
                throw new Error(`Failed to fetch shipments: ${listRes.status}`);
            }

            const apiShipments = await listRes.json();
            
            interface ApiShipment {
                tracking_id: string;
                status: string;
                sender_name: string;
                recipient_name: string;
                recipient_phone?: string;
                recipient_email?: string;
                recipient_address?: string;
                destination: string;
                weight: number;
                origin: string;
                created_at: string;
            }

            const shipments = apiShipments.map((s: ApiShipment) => ({
                id: s.tracking_id,
                trackingNumber: s.tracking_id,
                status: normalizeStatus(s.status) as ShipmentStatus,
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
            const statsRes = await fetchWithTimeout(`/api/admin/stats`, {
                next: { revalidate: 0 }
            });

            if (!statsRes.ok) {
                throw new Error(`Failed to fetch stats: ${statsRes.status}`);
            }

            const apiStats = await statsRes.json();

            const stats = {
                total: parseInt(apiStats.total),
                inTransit: parseInt(apiStats.intransit),
                outForDelivery: parseInt(apiStats.outfordelivery),
                delivered: parseInt(apiStats.delivered),
                pending: parseInt(apiStats.pending),
                canceled: parseInt(apiStats.canceled),
            };

            return { success: true, data: { shipments: shipments as ShipmentData[], stats } };
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
            const response = await fetchWithTimeout(`/api/admin/shipments/cleanup`, {
                method: 'DELETE'
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

    /**
     * Admin: Prune stale shipments (internal maintenance)
     */
    static async pruneStale(): Promise<ServiceResult<void>> {
        return this.bulkDeleteDelivered();
    }
}
