import { createAdminClient } from '@/lib/supabase/admin';
import { ServiceResult } from '@/types/shipment';

export interface AnalyticsData {
    timeline: { date: string; count: number }[];
    statusDistribution: { name: string; value: number; fill: string }[];
    topRoutes: { name: string; count: number }[];
    telemetry: Record<string, unknown>[];
}

export class AnalyticsService {
    static async getAnalyticsData(): Promise<ServiceResult<AnalyticsData>> {
        try {
            const supabase = createAdminClient();
            
            // 1. Fetch shipments for timeline and routes
            const thirtyDaysAgo = new Date();
            thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
            
            const { data: shipments, error } = await supabase
                .from('shipment')
                .select('created_at, status, origin, destination')
                .gte('created_at', thirtyDaysAgo.toISOString());
                
            if (error) throw new Error(error.message);
            
            // Process timeline
            const timelineMap = new Map<string, number>();
            // Process status
            const statusMap = new Map<string, number>();
            // Process routes
            const routeMap = new Map<string, number>();
            
            shipments?.forEach((s: { created_at: string, status: string, origin: string, destination: string }) => {
                // Timeline
                const date = new Date(s.created_at).toISOString().split('T')[0];
                timelineMap.set(date, (timelineMap.get(date) || 0) + 1);
                
                // Status
                const status = s.status?.toUpperCase() || 'PENDING';
                statusMap.set(status, (statusMap.get(status) || 0) + 1);
                
                // Routes
                if (s.origin && s.destination) {
                    const route = `${s.origin} → ${s.destination}`;
                    routeMap.set(route, (routeMap.get(route) || 0) + 1);
                }
            });
            
            // Format timeline (fill missing dates)
            const timeline = [];
            for (let i = 29; i >= 0; i--) {
                const d = new Date();
                d.setDate(d.getDate() - i);
                const dateStr = d.toISOString().split('T')[0];
                timeline.push({
                    date: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
                    count: timelineMap.get(dateStr) || 0
                });
            }
            
            // Format status
            const statusColors: Record<string, string> = {
                'PENDING': '#f59e0b',
                'IN_TRANSIT': '#3b82f6',
                'OUT_FOR_DELIVERY': '#8b5cf6',
                'DELIVERED': '#10b981',
                'CANCELED': '#ef4444'
            };
            
            const statusDistribution = Array.from(statusMap.entries()).map(([name, value]) => ({
                name: name.replace(/_/g, ' '),
                value,
                fill: statusColors[name] || '#6b7280'
            }));
            
            // Format routes
            const topRoutes = Array.from(routeMap.entries())
                .map(([name, count]) => ({ name, count }))
                .sort((a, b) => b.count - a.count)
                .slice(0, 5);
                
            // Fetch telemetry
            const { data: telemetryData } = await supabase
                .from('Telemetry')
                .select('*')
                .order('created_at', { ascending: false })
                .limit(20);
                
            return {
                success: true,
                data: {
                    timeline,
                    statusDistribution,
                    topRoutes,
                    telemetry: telemetryData || []
                }
            };
        } catch (error: unknown) {
            return { success: false, error: error instanceof Error ? error.message : 'Failed to fetch analytics' };
        }
    }
}
