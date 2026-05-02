-- ==============================================================================
-- ZERO-DOWNTIME SAAS MIGRATION
-- This script creates the multi-tenant structure and uses a Legacy Default 
-- to ensure the live bot doesn't break while we upgrade the codebase.
-- ==============================================================================

-- 1. Create the Companies table
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    admin_email TEXT NOT NULL UNIQUE,
    admin_password_hash TEXT,
    whatsapp_phone TEXT,
    logo_url TEXT,
    brand_color TEXT DEFAULT '#0066FF',
    auth_status TEXT DEFAULT 'pending_linking', -- pending_linking | active | disconnected
    subscription_status TEXT DEFAULT 'active',  -- active | suspended | cancelled
    subscription_expiry TIMESTAMP,
    setup_token TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Insert the "Legacy" Company (Your current Airwaybill setup)
-- We capture its UUID into a variable for the next steps
DO $$ 
DECLARE
    legacy_id UUID;
BEGIN
    -- Check if it already exists to be safe
    SELECT id INTO legacy_id FROM companies WHERE admin_email = 'airwaybill61@gmail.com';
    
    IF legacy_id IS NULL THEN
        INSERT INTO companies (name, admin_email, whatsapp_phone, auth_status, setup_token)
        VALUES (
            'Airwaybill', 
            'airwaybill61@gmail.com', 
            '2349077584528', 
            'active', 
            gen_random_uuid()::text
        ) RETURNING id INTO legacy_id;
    END IF;

    -- 3. Safely alter existing tables to add company_id
    -- We set the DEFAULT to the legacy_id so the LIVE bot keeps working perfectly!
    
    -- Shipments
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='shipment' AND column_name='company_id') THEN
        EXECUTE format('ALTER TABLE shipment ADD COLUMN company_id UUID REFERENCES companies(id) DEFAULT %L', legacy_id);
    END IF;

    -- Group Authority
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='groupauthority' AND column_name='company_id') THEN
        EXECUTE format('ALTER TABLE groupauthority ADD COLUMN company_id UUID REFERENCES companies(id) DEFAULT %L', legacy_id);
        
        -- Drop old primary key and add composite
        ALTER TABLE groupauthority DROP CONSTRAINT IF EXISTS groupauthority_pkey;
        ALTER TABLE groupauthority ADD PRIMARY KEY (company_id, jid);
    END IF;

    -- User Preference
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='userpreference' AND column_name='company_id') THEN
        EXECUTE format('ALTER TABLE userpreference ADD COLUMN company_id UUID REFERENCES companies(id) DEFAULT %L', legacy_id);
        
        -- Drop old primary key and add composite
        ALTER TABLE userpreference DROP CONSTRAINT IF EXISTS userpreference_pkey;
        ALTER TABLE userpreference ADD PRIMARY KEY (company_id, jid);
    END IF;

    -- System Config
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='systemconfig' AND column_name='company_id') THEN
        EXECUTE format('ALTER TABLE systemconfig ADD COLUMN company_id UUID REFERENCES companies(id) DEFAULT %L', legacy_id);
        
        ALTER TABLE systemconfig DROP CONSTRAINT IF EXISTS systemconfig_pkey;
        ALTER TABLE systemconfig ADD PRIMARY KEY (company_id, key);
    END IF;
    
    -- Telemetry
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='telemetry' AND column_name='company_id') THEN
        EXECUTE format('ALTER TABLE telemetry ADD COLUMN company_id UUID REFERENCES companies(id) DEFAULT %L', legacy_id);
    END IF;

END $$;
