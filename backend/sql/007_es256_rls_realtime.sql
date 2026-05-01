-- Enable RLS
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE shipment ENABLE ROW LEVEL SECURITY;

-- Clean up older policies to prevent conflicts
DROP POLICY IF EXISTS "Companies can read their own data" ON companies;
DROP POLICY IF EXISTS "Companies can read their own shipments" ON shipment;
DROP POLICY IF EXISTS "tenant_read_own_company" ON companies;
DROP POLICY IF EXISTS "tenant_read_own_shipments" ON shipment;

-- Companies: each tenant reads only their own row using standard Supabase auth.jwt()
CREATE POLICY "tenant_read_own_company" ON companies 
  FOR SELECT TO authenticated 
  USING (id = (auth.jwt() ->> 'company_id')::uuid);

-- Shipments: each tenant reads only their own shipments using standard Supabase auth.jwt()
CREATE POLICY "tenant_read_own_shipments" ON shipment 
  FOR SELECT TO authenticated 
  USING (company_id = (auth.jwt() ->> 'company_id')::uuid);

-- Add shipment and companies to Realtime publication
ALTER TABLE shipment REPLICA IDENTITY FULL;
ALTER TABLE companies REPLICA IDENTITY FULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_publication_tables 
    WHERE pubname = 'supabase_realtime' AND tablename = 'shipment'
  ) THEN
    ALTER PUBLICATION supabase_realtime ADD TABLE shipment;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_publication_tables 
    WHERE pubname = 'supabase_realtime' AND tablename = 'companies'
  ) THEN
    ALTER PUBLICATION supabase_realtime ADD TABLE companies;
  END IF;
END
$$;
