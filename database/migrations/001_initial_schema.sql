-- 0. Core Tables
CREATE TABLE IF NOT EXISTS SystemConfig (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS UserPreference (
    jid TEXT PRIMARY KEY,
    language TEXT NOT NULL DEFAULT 'en',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS GroupAuthority (
    jid TEXT PRIMARY KEY,
    is_authorized BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS Shipment (
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

CREATE INDEX IF NOT EXISTS idx_shipment_triggers_pending ON Shipment(status, scheduled_transit_time) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_shipment_triggers_transit ON Shipment(status, outfordelivery_time) WHERE status = 'intransit';
CREATE INDEX IF NOT EXISTS idx_shipment_triggers_outfordelivery ON Shipment(status, expected_delivery_time) WHERE status = 'outfordelivery';
CREATE INDEX IF NOT EXISTS idx_shipment_user_jid ON Shipment(user_jid);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_phone ON Shipment(recipient_phone) WHERE recipient_phone IS NOT NULL AND recipient_phone != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_email ON Shipment(recipient_email) WHERE recipient_email IS NOT NULL AND recipient_email != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_unique_id ON Shipment(recipient_id) WHERE recipient_id IS NOT NULL AND recipient_id != '';

-- 1. Country Timezones
CREATE TABLE IF NOT EXISTS country_timezones (
    country_name TEXT PRIMARY KEY,
    zone_name TEXT NOT NULL
);

-- Seed timezone data
INSERT INTO country_timezones (country_name, zone_name) VALUES
    ('nigeria', 'Africa/Lagos'),
    ('ghana', 'Africa/Accra'),
    ('kenya', 'Africa/Nairobi'),
    ('south africa', 'Africa/Johannesburg'),
    ('egypt', 'Africa/Cairo'),
    ('morocco', 'Africa/Casablanca'),
    ('tanzania', 'Africa/Dar_es_Salaam'),
    ('ethiopia', 'Africa/Addis_Ababa'),
    ('uganda', 'Africa/Kampala'),
    ('cameroon', 'Africa/Douala'),
    ('senegal', 'Africa/Dakar'),
    ('ivory coast', 'Africa/Abidjan'),
    ('angola', 'Africa/Luanda'),
    ('mozambique', 'Africa/Maputo'),
    ('zambia', 'Africa/Lusaka'),
    ('zimbabwe', 'Africa/Harare'),
    ('united states', 'America/New_York'),
    ('usa', 'America/New_York'),
    ('canada', 'America/Toronto'),
    ('mexico', 'America/Mexico_City'),
    ('brazil', 'America/Sao_Paulo'),
    ('argentina', 'America/Argentina/Buenos_Aires'),
    ('colombia', 'America/Bogota'),
    ('chile', 'America/Santiago'),
    ('peru', 'America/Lima'),
    ('united kingdom', 'Europe/London'),
    ('uk', 'Europe/London'),
    ('france', 'Europe/Paris'),
    ('germany', 'Europe/Berlin'),
    ('italy', 'Europe/Rome'),
    ('spain', 'Europe/Madrid'),
    ('netherlands', 'Europe/Amsterdam'),
    ('belgium', 'Europe/Brussels'),
    ('portugal', 'Europe/Lisbon'),
    ('switzerland', 'Europe/Zurich'),
    ('sweden', 'Europe/Stockholm'),
    ('norway', 'Europe/Oslo'),
    ('denmark', 'Europe/Copenhagen'),
    ('poland', 'Europe/Warsaw'),
    ('turkey', 'Europe/Istanbul'),
    ('russia', 'Europe/Moscow'),
    ('china', 'Asia/Shanghai'),
    ('japan', 'Asia/Tokyo'),
    ('south korea', 'Asia/Seoul'),
    ('india', 'Asia/Kolkata'),
    ('indonesia', 'Asia/Jakarta'),
    ('malaysia', 'Asia/Kuala_Lumpur'),
    ('singapore', 'Asia/Singapore'),
    ('thailand', 'Asia/Bangkok'),
    ('vietnam', 'Asia/Ho_Chi_Minh'),
    ('philippines', 'Asia/Manila'),
    ('pakistan', 'Asia/Karachi'),
    ('bangladesh', 'Asia/Dhaka'),
    ('saudi arabia', 'Asia/Riyadh'),
    ('uae', 'Asia/Dubai'),
    ('united arab emirates', 'Asia/Dubai'),
    ('qatar', 'Asia/Qatar'),
    ('israel', 'Asia/Jerusalem'),
    ('australia', 'Australia/Sydney'),
    ('new zealand', 'Pacific/Auckland')
ON CONFLICT (country_name) DO NOTHING;

-- 1. Tracking ID Generation
DROP FUNCTION IF EXISTS generate_tracking_id() CASCADE;
CREATE OR REPLACE FUNCTION generate_tracking_id()
 RETURNS text
 LANGUAGE plpgsql
AS $function$
DECLARE
    prefix TEXT := 'AWB';
    digits TEXT := '';
    i INT;
BEGIN
    FOR i IN 1..9 LOOP
        digits := digits || floor(random() * 10)::text;
    END LOOP;
    RETURN prefix || '-' || digits;
END;
$function$;

-- 2. Automated Scheduling Logic
DROP FUNCTION IF EXISTS fn_shipment_auto_schedule() CASCADE;
CREATE OR REPLACE FUNCTION fn_shipment_auto_schedule()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_dest_tz TEXT;
    v_now_lagos TIMESTAMP;
    v_departure_utc TIMESTAMP;
    v_arrival_local TIMESTAMP;
    v_snap_hour INT := 10;
    v_final_arrival_utc TIMESTAMP;
BEGIN
    NEW.updated_at := CURRENT_TIMESTAMP;

    IF (TG_OP = 'INSERT') 
       OR (NEW.origin IS DISTINCT FROM OLD.origin) 
       OR (NEW.destination IS DISTINCT FROM OLD.destination) 
       OR (NEW.scheduled_transit_time IS DISTINCT FROM OLD.scheduled_transit_time)
       OR (NEW.expected_delivery_time IS DISTINCT FROM OLD.expected_delivery_time) THEN
        
        IF NEW.tracking_id IS NULL OR NEW.tracking_id = '' THEN
            NEW.tracking_id := generate_tracking_id();
        END IF;

        SELECT zone_name INTO v_dest_tz FROM country_timezones WHERE country_name = lower(trim(NEW.destination));
        IF v_dest_tz IS NULL THEN v_dest_tz := 'UTC'; END IF;

        -- 1. Manual Arrival Override
        IF (TG_OP = 'UPDATE') AND (NEW.expected_delivery_time IS DISTINCT FROM OLD.expected_delivery_time) THEN
             NEW.outfordelivery_time := NEW.expected_delivery_time - interval '2 hours';
        
        -- 2. Auto-Scheduling
        ELSE
            IF (TG_OP = 'INSERT') THEN
                v_now_lagos := CURRENT_TIMESTAMP AT TIME ZONE 'Africa/Lagos';
                IF extract(hour from v_now_lagos) >= 22 THEN
                    v_departure_utc := (date_trunc('day', v_now_lagos + interval '1 day') + interval '8 hours') AT TIME ZONE 'Africa/Lagos';
                ELSE
                    v_departure_utc := CURRENT_TIMESTAMP + interval '1 hour';
                END IF;
                NEW.scheduled_transit_time := v_departure_utc;

                IF v_dest_tz = 'UTC' THEN v_snap_hour := 12; END IF;
                v_arrival_local := date_trunc('day', (v_departure_utc AT TIME ZONE v_dest_tz) + interval '1 day') + (v_snap_hour * interval '1 hour');
                v_final_arrival_utc := v_arrival_local AT TIME ZONE v_dest_tz;
            ELSE
                v_departure_utc := NEW.scheduled_transit_time;
                IF (v_departure_utc <= CURRENT_TIMESTAMP + interval '10 minutes') THEN
                    v_final_arrival_utc := v_departure_utc + interval '1 day';
                ELSE
                    IF v_dest_tz = 'UTC' THEN v_snap_hour := 12; END IF;
                    v_arrival_local := date_trunc('day', (v_departure_utc AT TIME ZONE v_dest_tz) + interval '1 day') + (v_snap_hour * interval '1 hour');
                    v_final_arrival_utc := v_arrival_local AT TIME ZONE v_dest_tz;
                END IF;
            END IF;
            
            NEW.expected_delivery_time := v_final_arrival_utc;
            NEW.outfordelivery_time := v_final_arrival_utc - interval '2 hours';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

DROP TRIGGER IF EXISTS trg_shipment_init ON Shipment;
CREATE TRIGGER trg_shipment_init
BEFORE INSERT OR UPDATE ON Shipment
FOR EACH ROW EXECUTE FUNCTION fn_shipment_auto_schedule();

-- 3. Atomic Status Transitions
DROP FUNCTION IF EXISTS fn_process_status_transitions(TIMESTAMP) CASCADE;
CREATE OR REPLACE FUNCTION fn_process_status_transitions(now_utc TIMESTAMP)
RETURNS TABLE (
    r_tracking_id TEXT,
    new_status TEXT,
    r_user_jid TEXT,
    r_recipient_email TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    WITH updated_transit AS (
        UPDATE Shipment
        SET status = 'intransit', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'pending' AND scheduled_transit_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid, recipient_email
    ),
    updated_out AS (
        UPDATE Shipment
        SET status = 'outfordelivery', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'intransit' AND outfordelivery_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid, recipient_email
    ),
    updated_delivered AS (
        UPDATE Shipment
        SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'outfordelivery' AND expected_delivery_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid, recipient_email
    )
    SELECT * FROM updated_transit
    UNION ALL
    SELECT * FROM updated_out
    UNION ALL
    SELECT * FROM updated_delivered;
END;
$$;

-- 4. Aged Data Pruning
DROP FUNCTION IF EXISTS fn_prune_aged_shipments() CASCADE;
CREATE OR REPLACE FUNCTION fn_prune_aged_shipments()
RETURNS TABLE (deleted_count BIGINT)
LANGUAGE plpgsql
AS $$
DECLARE
    two_days_ago TIMESTAMP := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 days';
    seven_days_ago TIMESTAMP := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '7 days';
    count1 BIGINT;
    count2 BIGINT;
BEGIN
    WITH del1 AS (
        DELETE FROM Shipment 
        WHERE status = 'delivered' AND updated_at < two_days_ago
        RETURNING 1
    ) SELECT count(*) INTO count1 FROM del1;

    WITH del2 AS (
        DELETE FROM Shipment 
        WHERE created_at < seven_days_ago
        RETURNING 1
    ) SELECT count(*) INTO count2 FROM del2;

    RETURN QUERY SELECT (count1 + count2);
END;
$$;
