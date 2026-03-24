CREATE TABLE SystemConfig (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE UserPreference (
    jid TEXT PRIMARY KEY,
    language TEXT NOT NULL DEFAULT 'en',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE GroupAuthority (
    jid TEXT PRIMARY KEY,
    is_authorized BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Shipment (
    tracking_id TEXT PRIMARY KEY,
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
