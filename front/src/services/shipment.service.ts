import { CreateShipmentDto, ShipmentData, ServiceResult, ShipmentStatus, DashboardStats, PaginatedResult, Pagination } from '@/types/shipment';
import { logger } from '@/lib/logger';

import { createAdminClient } from '@/lib/supabase/admin';

function getNextJsBaseUrl() {
    if (typeof window !== 'undefined') return '';
    if (process.env.API_URL) return process.env.API_URL;
    if (process.env.VERCEL_URL) return `https://${process.env.VERCEL_URL}`;
    return `http://localhost:${process.env.PORT ?? 3000}`;
}

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
    const baseUrl = getNextJsBaseUrl();
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

const flattenSqlc = (obj: unknown): unknown => {
    if (obj === null || typeof obj !== 'object') return obj;
    if (Array.isArray(obj)) return obj.map(flattenSqlc);
    const objAsRecord = obj as Record<string, unknown>;
    if ('Valid' in objAsRecord && Object.keys(objAsRecord).length === 2) {
        if (!objAsRecord.Valid) return null;
        if ('String' in objAsRecord) return objAsRecord.String;
        if ('Time' in objAsRecord) return objAsRecord.Time;
        if ('Float64' in objAsRecord) return objAsRecord.Float64;
        if ('Int64' in objAsRecord) return objAsRecord.Int64;
        if ('Bool' in objAsRecord) return objAsRecord.Bool;
    }
    const res: Record<string, unknown> = {};
    for (const k in objAsRecord) {
        res[k] = flattenSqlc(objAsRecord[k]);
    }
    return res;
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

            const result = flattenSqlc(await response.json()) as { tracking_id: string };
            return { success: true, data: { trackingNumber: result.tracking_id } };
        } catch (error) {
            const apiError = handleApiError(error, 'Create shipment');
            return { success: false, error: apiError.userMessage };
        }
    }

    static async parseText(text: string): Promise<ServiceResult<Record<string, unknown>>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/parse`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text }),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            const apiError = handleApiError(error, 'Parse AI text');
            return { success: false, error: apiError.userMessage };
        }
    }

    static async getByTracking(trackingNumber: string): Promise<ShipmentData | null> {
        if (!trackingNumber) return null;

        try {
            const supabase = createAdminClient();

            const { data, error } = await supabase
                .from('shipment')
                .select('*')
                .eq('tracking_id', trackingNumber)
                .single();

            if (error || !data) {
                if (error?.code === 'PGRST116') return null;
                throw new Error(`Supabase error: ${error?.message}`);
            }


            const timelineStr = (val: unknown) => typeof val === 'string' ? val : '';
            const statusStr = typeof data.status === 'string' ? data.status.toLowerCase() : '';
            const scheduledTransit = timelineStr(data.scheduled_transit_time);
            const expectedDelivery = timelineStr(data.expected_delivery_time);

            const timeline = [
                {
                    status: 'Order Placed',
                    timestamp: timelineStr(data.created_at),
                    description: `Shipment registered at ${timelineStr(data.origin) || 'origin'}`,
                    is_completed: true
                },
                {
                    status: 'In Transit',
                    timestamp: scheduledTransit,
                    description: 'Package has left the origin facility and is on its way',
                    is_completed: ['intransit', 'outfordelivery', 'delivered'].includes(statusStr)
                },
                {
                    status: 'Out for Delivery',
                    timestamp: timelineStr(data.outfordelivery_time),
                    description: 'Package is with our local agent for final delivery',
                    is_completed: ['outfordelivery', 'delivered'].includes(statusStr)
                },
                {
                    status: 'Delivered',
                    timestamp: expectedDelivery,
                    description: 'Package has been successfully delivered',
                    is_completed: statusStr === 'delivered'
                }
            ];

            const redactName = (name: unknown): string => {
                if (typeof name !== 'string' || !name) return 'N/A';
                const parts = name.split(' ');
                if (parts[0].length <= 2) return parts[0] + '***';
                return parts[0].substring(0, 2) + '******';
            };

            const shipment: ShipmentData = {
                id: timelineStr(data.tracking_id),
                trackingNumber: timelineStr(data.tracking_id),
                status: normalizeStatus(statusStr) as ShipmentStatus,
                senderName: redactName(data.sender_name),
                receiverName: redactName(data.recipient_name),
                receiverPhone: typeof data.recipient_phone === 'string' ? data.recipient_phone : null,
                receiverEmail: typeof data.recipient_email === 'string' ? data.recipient_email : null,
                receiverAddress: typeof data.recipient_address === 'string' ? data.recipient_address : null,
                receiverCountry: timelineStr(data.destination) || 'N/A',
                weight: typeof data.weight === 'number' ? data.weight : (typeof data.weight === 'string' ? parseFloat(data.weight) : 0),
                senderCountry: timelineStr(data.origin) || 'N/A',
                timeline: timeline,
                events: [],
                createdAt: timelineStr(data.created_at),
                scheduledTransitTime: scheduledTransit,
                outfordeliveryTime: timelineStr(data.outfordelivery_time),
                expectedDeliveryTime: expectedDelivery,
                isArchived: statusStr === 'delivered',
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
     * Admin: Paginated shipment list with filtering
     */
    static async getShipments(params: {
        page?: number;
        limit?: number;
        search?: string;
        status?: string;
    } = {}): Promise<ServiceResult<PaginatedResult<ShipmentData>>> {
        try {
            const supabase = createAdminClient();
            const page = params.page || 1;
            const limit = params.limit || 20;
            const offset = (page - 1) * limit;

            // Build the query
            let query = supabase
                .from('shipment')
                .select('*', { count: 'exact' });

            // Apply status filter
            if (params.status && params.status !== 'all') {
                query = query.eq('status', params.status.toUpperCase());
            }

            // Apply search filter (ilike on tracking_id, sender_name, recipient_name)
            if (params.search) {
                query = query.or(
                    `tracking_id.ilike.%${params.search}%,sender_name.ilike.%${params.search}%,recipient_name.ilike.%${params.search}%`
                );
            }

            // Apply pagination and ordering
            const { data: apiShipments, count, error } = await query
                .order('created_at', { ascending: false })
                .range(offset, offset + limit - 1);

            if (error) throw new Error(`Supabase fetch error: ${error.message}`);

            const shipments = (apiShipments || []).map(s => ({
                id: s.tracking_id,
                trackingNumber: s.tracking_id,
                status: normalizeStatus(s.status) as ShipmentStatus,
                senderName: s.sender_name,
                receiverName: s.recipient_name,
                receiverPhone: s.recipient_phone ?? null,
                receiverEmail: s.recipient_email ?? null,
                receiverAddress: s.recipient_address ?? null,
                receiverCountry: s.destination,
                weight: s.weight,
                senderCountry: s.origin,
                createdAt: s.created_at,
                scheduledTransitTime: s.scheduled_transit_time,
                outfordeliveryTime: s.outfordelivery_time,
                expectedDeliveryTime: s.expected_delivery_time,
                isArchived: typeof s.status === 'string' && s.status.toLowerCase() === 'delivered',
                events: [],
                timeline: [],
            }));

            const totalCount = count || 0;
            const pagination: Pagination = {
                page,
                limit,
                total: totalCount,
                totalPages: Math.ceil(totalCount / limit),
            };

            return {
                success: true,
                data: {
                    data: shipments,
                    pagination,
                }
            };
        } catch (error) {
            const apiError = handleApiError(error, 'Paginated shipments');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Dashboard data via Next.js API
     */
    static async getDashboardData(): Promise<ServiceResult<{ shipments: ShipmentData[], stats: DashboardStats }>> {
        try {
            const supabase = createAdminClient();

            // Fetch Recent List
            const { data: apiShipments, error: listError } = await supabase
                .from('shipment')
                .select('*')
                .order('created_at', { ascending: false })
                .limit(10);

            if (listError) throw new Error(`Supabase fetch error: ${listError.message}`);

            const shipments = (apiShipments || []).map(s => ({
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
                scheduledTransitTime: s.scheduled_transit_time,
                outfordeliveryTime: s.outfordelivery_time,
                expectedDeliveryTime: s.expected_delivery_time,
                isArchived: typeof s.status === 'string' && s.status.toLowerCase() === 'delivered',
            }));

            // Fetch Stats using concurrent count queries
            const countQuery = async (statusFilter?: string) => {
                let q = supabase.from('shipment').select('*', { count: 'exact', head: true });
                if (statusFilter) q = q.eq('status', statusFilter);
                const { count, error } = await q;
                if (error) throw new Error(`Supabase count error: ${error.message}`);
                return count || 0;
            };

            const [total, inTransit, outForDelivery, delivered, pending, canceled] = await Promise.all([
                countQuery(),
                countQuery('IN_TRANSIT'),
                countQuery('OUT_FOR_DELIVERY'),
                countQuery('DELIVERED'),
                countQuery('PENDING'),
                countQuery('CANCELED'),
            ]);

            const stats = { total, inTransit, outForDelivery, delivered, pending, canceled };

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

    /**
     * Admin: Bulk update status via labels or selection
     */
    static async bulkUpdateStatus(ids: string[], status: string): Promise<ServiceResult<{ updated: number }>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/bulk_status`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids, status: status.toLowerCase() }),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            const data = await response.json();
            return { success: true, data: { updated: data.updated } };
        } catch (error) {
            const apiError = handleApiError(error, 'Bulk update status');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Bulk delete selected shipments
     */
    static async bulkDeleteShipments(ids: string[]): Promise<ServiceResult<{ deleted: number }>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/bulk_delete`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids }),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            const data = await response.json();
            return { success: true, data: { deleted: data.deleted } };
        } catch (error) {
            const apiError = handleApiError(error, 'Bulk delete shipments');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Bulk create from CSV manifest
     */
    static async bulkCreateCSV(text: string): Promise<ServiceResult<{ created: number; failed: number }>> {
        try {
            const response = await fetchWithTimeout(`/api/admin/shipments/bulk_csv`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text }),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            const data = await response.json();
            return { success: true, data: { created: data.created, failed: data.failed } };
        } catch (error) {
            const apiError = handleApiError(error, 'Bulk CSV create');
            return { success: false, error: apiError.userMessage };
        }
    }

    /**
     * Admin: Fetch system telemetry handles
     */
    static async getTelemetry(): Promise<ServiceResult<{ stats: unknown[]; recent: unknown[] }>> {
        try {
            const supabase = createAdminClient();

            // Fetch recent telemetry events
            const { data: recent, error: recentError } = await supabase
                .from('telemetry')
                .select('*')
                .order('created_at', { ascending: false })
                .limit(50);

            if (recentError) throw new Error(`Supabase telemetry error: ${recentError.message}`);

            // Aggregate stats by event_type
            const { data: stats, error: statsError } = await supabase
                .from('telemetry')
                .select('event_type')
                .order('created_at', { ascending: false });

            if (statsError) throw new Error(`Supabase stats error: ${statsError.message}`);

            // Group by event_type for counts
            const grouped = (stats || []).reduce<Record<string, number>>((acc, row) => {
                const key = typeof row.event_type === 'string' ? row.event_type : 'unknown';
                acc[key] = (acc[key] || 0) + 1;
                return acc;
            }, {});

            const statsSummary = Object.entries(grouped).map(([event_type, count]) => ({
                event_type,
                count,
            }));

            return { success: true, data: { stats: statsSummary, recent: recent || [] } };
        } catch (error) {
            const apiError = handleApiError(error, 'Fetch telemetry');
            return { success: false, error: apiError.userMessage };
        }
    }
}
