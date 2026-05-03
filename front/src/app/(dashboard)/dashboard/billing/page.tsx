import { Suspense } from 'react';
import { getPlansAction, getPaymentHistoryAction } from '@/app/actions/billing';
import BillingClient from '@/components/billing/BillingClient';
import { Loader2 } from 'lucide-react';
import { getServerSession } from '@/lib/auth';
import { createClient } from '@/lib/supabase/server';
import { redirect } from 'next/navigation';

export default async function BillingPage() {
    const { user } = await getServerSession();
    if (!user || !user.company_id) redirect('/auth');

    // Fetch data on the server
    const [plansRes, paymentsRes, supabase] = await Promise.all([
        getPlansAction(),
        getPaymentHistoryAction(),
        createClient()
    ]);

    const { data: companyData } = await supabase
        .from('companies')
        .select('subscription_expiry, plan_type, subscription_status')
        .eq('id', user.company_id)
        .single();

    const plans = plansRes.success ? plansRes.data || [] : [];
    const payments = paymentsRes.success ? paymentsRes.data || [] : [];

    return (
        <Suspense fallback={
            <div className="min-h-screen flex items-center justify-center bg-transparent">
                <Loader2 className="w-8 h-8 text-accent animate-spin" />
            </div>
        }>
            <BillingClient 
                initialPlans={plans} 
                initialPayments={payments} 
                companyData={companyData}
            />
        </Suspense>
    );
}
