-- ==========================================
-- UPDATE ROW LEVEL SECURITY (RLS) POLICIES
-- Security fix: Secure tables against anonymous Supabase JS client queries
-- ==========================================

-- 1. Secure the companies table
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies 
        WHERE tablename = 'companies' AND policyname = 'Companies can read their own data'
    ) THEN
        CREATE POLICY "Companies can read their own data" 
        ON companies FOR SELECT 
        USING (
            id::text = current_setting('request.jwt.claims', true)::json->>'company_id'
        );
    END IF;
END $$;

-- 2. Secure the shipment table
-- Drop the overly permissive public policy that allowed anyone to read all shipments
DROP POLICY IF EXISTS "Public can read shipments by tracking_id" ON shipment;

-- Restrict shipments to only the authenticated company
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies 
        WHERE tablename = 'shipment' AND policyname = 'Companies can read their own shipments'
    ) THEN
        CREATE POLICY "Companies can read their own shipments" 
        ON shipment FOR SELECT 
        USING (
            company_id::text = current_setting('request.jwt.claims', true)::json->>'company_id'
        );
    END IF;
END $$;
