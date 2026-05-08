import { redirect } from 'next/navigation';
import { getServerSession } from '@/lib/auth';
import { createClient } from '@/lib/supabase/server';
import { getJwtTokenAction } from '@/app/actions/auth';
import DashboardClient from '@/components/dashboard/DashboardClient';

export default async function DashboardPage() {
    const { user } = await getServerSession();

    if (!user || !user.company_id) {
        redirect('/auth');
    }

    const supabase = await createClient();
    const jwt = await getJwtTokenAction();

    // Fetch company data
    const { data: companyData, error: companyError } = await supabase
        .from('companies')
        .select('name, admin_email, subscription_status, subscription_expiry, plan_type, auth_status, whatsapp_phone, brand_color, logo_url, tracking_prefix')
        .eq('id', user.company_id)
        .single();

    if (companyError && companyError.code === 'PGRST116') {
        redirect('/auth');
    }

    // Fetch initial stats using COUNT to avoid memory/performance issues
    const [
        { count: totalCount },
        { count: activeCount },
        { count: deliveredCount }
    ] = await Promise.all([
        supabase.from('shipment').select('*', { count: 'exact', head: true }).eq('company_id', user.company_id),
        supabase.from('shipment').select('*', { count: 'exact', head: true }).eq('company_id', user.company_id).in('status', ['pending', 'intransit', 'outfordelivery']),
        supabase.from('shipment').select('*', { count: 'exact', head: true }).eq('company_id', user.company_id).eq('status', 'delivered')
    ]);

    const stats = {
        total: totalCount || 0,
        active: activeCount || 0,
        delivered: deliveredCount || 0
    };

    return (
        <DashboardClient 
            initialCompanyData={companyData} 
            initialStats={stats} 
            user={user} 
            companyId={user.company_id}
            jwt={jwt}
        />
    );
}
