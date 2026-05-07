CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    admin_email TEXT NOT NULL UNIQUE,
    admin_password_hash TEXT,
    whatsapp_phone TEXT,
    logo_url TEXT,
    brand_color TEXT DEFAULT '#0066FF',
    auth_status TEXT DEFAULT 'pending_linking',
    subscription_status TEXT DEFAULT 'active',
    subscription_expiry TIMESTAMP,
    plan_type TEXT DEFAULT 'trial',
    setup_token TEXT UNIQUE,
    tracking_prefix TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS systemconfig (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    key TEXT,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, key)
);

CREATE TABLE IF NOT EXISTS userpreference (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    jid TEXT,
    language TEXT NOT NULL DEFAULT 'en',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, jid)
);

CREATE TABLE IF NOT EXISTS groupauthority (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    jid TEXT,
    is_authorized BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, jid)
);

CREATE TABLE IF NOT EXISTS shipment (
    tracking_id TEXT PRIMARY KEY,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    user_jid TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    scheduled_transit_time TIMESTAMP,
    outfordelivery_time TIMESTAMP,
    expected_delivery_time TIMESTAMP,
    
    sender_timezone TEXT,
    recipient_timezone TEXT,

    sender_name TEXT,
    sender_phone TEXT,
    origin TEXT,
    recipient_name TEXT,
    recipient_phone TEXT,
    recipient_email TEXT,
    recipient_id TEXT,
    recipient_address TEXT,
    destination TEXT,
    cargo_type TEXT,
    weight DOUBLE PRECISION,
    cost DOUBLE PRECISION,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_telemetry_event_type ON telemetry(event_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_created_at ON telemetry(created_at);

-- Performance indexes for multi-tenant scalability
CREATE INDEX IF NOT EXISTS idx_shipment_company_created ON shipment(company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipment_company_status ON shipment(company_id, status);
CREATE INDEX IF NOT EXISTS idx_shipment_company_user ON shipment(company_id, user_jid);
CREATE INDEX IF NOT EXISTS idx_telemetry_company_created ON telemetry(company_id, created_at DESC);

-- Enable RLS
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE shipment ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;

-- Companies: each tenant reads only their own row
DROP POLICY IF EXISTS "tenant_read_own_company" ON companies;
CREATE POLICY "tenant_read_own_company" ON companies 
  FOR SELECT TO authenticated 
  USING (id = (auth.jwt() ->> 'company_id')::uuid);

-- Shipments: each tenant reads only their own shipments  
DROP POLICY IF EXISTS "tenant_read_own_shipments" ON shipment;
CREATE POLICY "tenant_read_own_shipments" ON shipment 
  FOR SELECT TO authenticated 
  USING (company_id = (auth.jwt() ->> 'company_id')::uuid);
  
-- Payments: each tenant reads only their own payments
DROP POLICY IF EXISTS "tenant_read_own_payments" ON payments;
CREATE POLICY "tenant_read_own_payments" ON payments
  FOR SELECT TO authenticated
  USING (company_id = (auth.jwt() ->> 'company_id')::uuid);

-- Shipments: anon users cannot read shipments directly to prevent enumeration
DROP POLICY IF EXISTS "anon_read_shipment_by_tracking" ON shipment;

-- RPC function for public tracking to force tracking_id requirement
CREATE OR REPLACE FUNCTION get_public_shipment(p_tracking_id TEXT)
RETURNS SETOF shipment
LANGUAGE sql SECURITY DEFINER
AS $$
  SELECT * FROM shipment WHERE tracking_id = p_tracking_id LIMIT 1;
$$;

-- Add shipment and companies to Realtime publication
ALTER TABLE shipment REPLICA IDENTITY FULL;
ALTER TABLE companies REPLICA IDENTITY FULL;
ALTER TABLE payments REPLICA IDENTITY FULL;

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

  IF NOT EXISTS (
    SELECT 1 FROM pg_publication_tables 
    WHERE pubname = 'supabase_realtime' AND tablename = 'payments'
  ) THEN
    ALTER PUBLICATION supabase_realtime ADD TABLE payments;
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    reference TEXT UNIQUE NOT NULL,
    amount DOUBLE PRECISION,
    status TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    desc_key TEXT NOT NULL,
    base_price INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'NGN',
    interval_key TEXT NOT NULL DEFAULT 'monthlyInterval',
    popular BOOLEAN DEFAULT FALSE,
    trial_key TEXT,
    btn_key TEXT NOT NULL,
    features JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
