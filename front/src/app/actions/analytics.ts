import { AnalyticsService } from '@/services/analytics.service';
import { isAdmin } from './shipment';

export async function getAnalyticsDashboardData() {
    if (!(await isAdmin())) return { success: false, error: 'Unauthorized' };
    return await AnalyticsService.getAnalyticsData();
}
