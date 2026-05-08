import { redirect } from 'next/navigation';
import { getServerSession } from '@/lib/auth';
import { createClient } from '@/lib/supabase/server';
import SuperAdminClient from '@/components/dashboard/super-admin/SuperAdminClient';

export default async function SuperAdminPage() {
    const { user } = await getServerSession();

    // Secondary server-side protection
    if (!user || user.role !== 'super_admin') {
        redirect('/dashboard');
    }

    const supabase = await createClient();

    // Fetch all companies/tenants
    const { data: companies, error } = await supabase
        .from('companies')
        .select('*')
        .order('created_at', { ascending: false });

    if (error) {
        console.error("Failed to fetch companies for super admin", error);
    }

    // Retrieve the JWT from the cookie store to enable client-side Supabase authentication
    const { cookies } = await import('next/headers');
    const cookieStore = await cookies();
    const jwtToken = cookieStore.get('jwt')?.value;

    return (
        <SuperAdminClient 
            user={user}
            initialCompanies={companies || []}
            jwt={jwtToken}
        />
    );
}
