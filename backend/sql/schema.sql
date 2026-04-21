CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    admin_email TEXT NOT NULL UNIQUE,
    admin_password_hash TEXT,
    whatsapp_phone TEXT,
    logo_url TEXT,
    brand_color TEXT DEFAULT '#0066FF',
    auth_status TEXT DEFAULT 'pending_linking',
    subscription_status TEXT DEFAULT 'active',
    subscription_expiry TIMESTAMP,
    plan_type TEXT DEFAULT 'pro',
    setup_token TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS SystemConfig (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    key TEXT,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, key)
);

CREATE TABLE IF NOT EXISTS UserPreference (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    jid TEXT,
    language TEXT NOT NULL DEFAULT 'en',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, jid)
);

CREATE TABLE IF NOT EXISTS GroupAuthority (
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    jid TEXT,
    is_authorized BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, jid)
);

CREATE TABLE IF NOT EXISTS Shipment (
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

CREATE TABLE IF NOT EXISTS Telemetry (
    id SERIAL PRIMARY KEY,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_telemetry_event_type ON Telemetry(event_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_created_at ON Telemetry(created_at);
