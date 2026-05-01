import { redirect } from 'next/navigation';
import { getServerSession } from '@/lib/auth';
import { createClient } from '@/lib/supabase/server';
import DashboardClient from '@/components/dashboard/DashboardClient';

export default async function DashboardPage() {
    const { user } = await getServerSession();

    if (!user || !user.company_id) {
        redirect('/auth');
    }

    const supabase = await createClient();

    // Fetch company data
    const { data: companyData, error: companyError } = await supabase
        .from('companies')
        .select('name, admin_email, subscription_status, subscription_expiry, plan_type, auth_status, whatsapp_phone, brand_color, logo_url, tracking_prefix')
        .eq('id', user.company_id)
        .single();

    if (companyError && companyError.code === 'PGRST116') {
        redirect('/auth');
    }

    // Fetch initial stats
    const { data: shipments } = await supabase
        .from('shipment')
        .select('status')
        .eq('company_id', user.company_id);

    let stats = { total: 0, active: 0, delivered: 0 };
    if (shipments) {
        const active = shipments.filter(s => s.status !== 'DELIVERED' && s.status !== 'CANCELED').length;
        const delivered = shipments.filter(s => s.status === 'DELIVERED').length;
        stats = { total: shipments.length, active, delivered };
    }

    return (
        <DashboardClient 
            initialCompanyData={companyData} 
            initialStats={stats} 
            user={user} 
            companyId={user.company_id} 
        />
    );
}
