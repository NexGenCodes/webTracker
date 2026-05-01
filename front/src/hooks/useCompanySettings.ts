"use client";

import { useState, useEffect } from 'react';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';
import { createClient } from '@/lib/supabase/client';
import { 
    CONTACT_EMAIL, 
    CONTACT_PHONE, 
    CONTACT_HQ 
} from '@/constants';

export function useCompanySettings() {
    const { companyId } = useMultiTenant();
    const [settings, setSettings] = useState({
        companyName: '',
        contactEmail: CONTACT_EMAIL,
        contactPhone: CONTACT_PHONE,
        contactHq: CONTACT_HQ,
        logoUrl: '',
        trackingPrefix: '',
        brandColor: '#0066FF'
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (!companyId) {
            setLoading(false);
            return;
        }

        const supabase = createClient();

        const fetchSettings = async () => {
            try {
                const { data, error } = await supabase
                    .from('companies')
                    .select('name, admin_email, whatsapp_phone, logo_url, tracking_prefix, brand_color')
                    .eq('id', companyId)
                    .single();

                if (data && !error) {
                    setSettings({
                        companyName: data.name || '',
                        contactEmail: data.admin_email || CONTACT_EMAIL,
                        contactPhone: data.whatsapp_phone || CONTACT_PHONE,
                        contactHq: CONTACT_HQ,
                        logoUrl: data.logo_url || '',
                        trackingPrefix: data.tracking_prefix || '',
                        brandColor: data.brand_color || '#0066FF'
                    });
                }
            } catch (error) {
                console.error("Failed to fetch company settings:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchSettings();
    }, [companyId]);

    return { settings, loading };
}
