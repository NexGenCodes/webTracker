import pg from 'pg';
import { readFileSync } from 'fs';
import { join } from 'path';

// Parse the backend .env to get DATABASE_URL
const envContent = readFileSync(join(process.cwd(), '..', 'backend', '.env'), 'utf-8');
const dbUrlMatch = envContent.match(/DATABASE_URL="([^"]+)"/);

if (!dbUrlMatch) {
  console.error("DATABASE_URL not found in backend/.env");
  process.exit(1);
}

const dbUrl = dbUrlMatch[1];
const pool = new pg.Pool({ connectionString: dbUrl });

const migrationSql = `
-- Drop existing problematic policies if they exist
DROP POLICY IF EXISTS "tenant_read_own_company" ON companies;
DROP POLICY IF EXISTS "tenant_read_own_shipments" ON shipment;
DROP POLICY IF EXISTS "Allow anonymous read access" ON companies;
DROP POLICY IF EXISTS "Allow anonymous read access" ON shipment;

-- Enable RLS (just in case)
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE shipment ENABLE ROW LEVEL SECURITY;

-- Create secure policies using the 'company_id' from our custom JWT
CREATE POLICY "tenant_read_own_company" ON companies 
  FOR SELECT TO authenticated 
  USING (id = (auth.jwt() ->> 'company_id')::uuid);

CREATE POLICY "tenant_read_own_shipments" ON shipment 
  FOR SELECT TO authenticated 
  USING (company_id = (auth.jwt() ->> 'company_id')::uuid);

-- Configure Realtime
ALTER TABLE shipment REPLICA IDENTITY FULL;
ALTER TABLE companies REPLICA IDENTITY FULL;

-- Try adding to publication (might fail if already added, we catch it)
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
`;

async function run() {
  console.log("Running Supabase RLS and Realtime Migration...");
  try {
    await pool.query(migrationSql);
    console.log("✅ Migration successful!");
  } catch (err) {
    console.error("❌ Migration failed:", err.message);
  } finally {
    pool.end();
  }
}

run();
