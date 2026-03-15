-- 001_initial_schema.sql
-- Consolidated database logic for WebTracker

-- 0. Country Timezones (lookup table used by scheduling trigger)
CREATE TABLE IF NOT EXISTS public.country_timezones (
    country_name TEXT PRIMARY KEY,
    zone_name TEXT NOT NULL
);

-- Seed timezone data (idempotent via ON CONFLICT)
INSERT INTO public.country_timezones (country_name, zone_name) VALUES
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
CREATE OR REPLACE FUNCTION public.generate_tracking_id()
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
CREATE OR REPLACE FUNCTION public.fn_shipment_auto_schedule()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    dest_tz TEXT;
    now_lagos TIMESTAMP;
    departure_utc TIMESTAMP;
    transit_hours INT := 10;
    earliest_arrival_utc TIMESTAMP;
    arrival_local TIMESTAMP;
    snap_hour INT := 10;
    final_arrival_utc TIMESTAMP;
    NEW.updated_at := CURRENT_TIMESTAMP;

    IF (TG_OP = 'INSERT') 
       OR (NEW.origin IS DISTINCT FROM OLD.origin) 
       OR (NEW.destination IS DISTINCT FROM OLD.destination) 
       OR (NEW.scheduled_transit_time IS DISTINCT FROM OLD.scheduled_transit_time)
       OR (NEW.expected_delivery_time IS DISTINCT FROM OLD.expected_delivery_time) THEN
        
        IF NEW.tracking_id IS NULL OR NEW.tracking_id = '' THEN
            NEW.tracking_id := generate_tracking_id();
        END IF;

        SELECT zone_name INTO dest_tz FROM public.country_timezones WHERE country_name = lower(trim(NEW.destination));
        IF dest_tz IS NULL THEN dest_tz := 'UTC'; END IF;

        -- 1. Manual Arrival Override: If the expected_delivery_time itself was edited
        IF (TG_OP = 'UPDATE') AND (NEW.expected_delivery_time IS DISTINCT FROM OLD.expected_delivery_time) THEN
             NEW.outfordelivery_time := NEW.expected_delivery_time - interval '2 hours';
        
        -- 2. Auto-Scheduling: For new shipments or edits to Location/Departure
        ELSE
            IF (TG_OP = 'INSERT') THEN
                -- ORIGINAL CREATION LOGIC: 1h delay, 10 PM cap logic, and 10 AM snap arrival
                now_lagos := CURRENT_TIMESTAMP AT TIME ZONE 'Africa/Lagos';
                IF extract(hour from now_lagos) >= 22 THEN
                    departure_utc := (date_trunc('day', now_lagos + interval '1 day') + interval '8 hours') AT TIME ZONE 'Africa/Lagos';
                ELSE
                    departure_utc := CURRENT_TIMESTAMP + interval '1 hour';
                END IF;
                NEW.scheduled_transit_time := departure_utc;

                -- Snapping Arrival to 10 AM Local
                IF dest_tz = 'UTC' THEN snap_hour := 12; END IF;
                arrival_local := date_trunc('day', (departure_utc AT TIME ZONE dest_tz) + interval '1 day') + (snap_hour * interval '1 hour');
                final_arrival_utc := arrival_local AT TIME ZONE dest_tz;
            ELSE
                -- EDIT LOGIC (UPDATE)
                departure_utc := NEW.scheduled_transit_time;
                
                -- Branch: Past/Today vs Future
                IF (departure_utc <= CURRENT_TIMESTAMP + interval '10 minutes') THEN
                    -- Strict Rules for Past/Today: Arrival = Departure + 1 day (Strict 24h)
                    final_arrival_utc := departure_utc + interval '1 day';
                ELSE
                    -- Future Rules: Snapping Arrival to 10 AM Local
                    IF dest_tz = 'UTC' THEN snap_hour := 12; END IF;
                    arrival_local := date_trunc('day', (departure_utc AT TIME ZONE dest_tz) + interval '1 day') + (snap_hour * interval '1 hour');
                    final_arrival_utc := arrival_local AT TIME ZONE dest_tz;
                END IF;
            END IF;
            
            NEW.expected_delivery_time := final_arrival_utc;
            NEW.outfordelivery_time := final_arrival_utc - interval '2 hours';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

DROP TRIGGER IF EXISTS trg_shipment_init ON public.Shipment;
CREATE TRIGGER trg_shipment_init
BEFORE INSERT OR UPDATE ON public.Shipment
FOR EACH ROW EXECUTE FUNCTION public.fn_shipment_auto_schedule();

-- 3. Atomic Status Transitions
CREATE OR REPLACE FUNCTION public.fn_process_status_transitions(now_utc TIMESTAMP)
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
CREATE OR REPLACE FUNCTION public.fn_prune_aged_shipments()
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
