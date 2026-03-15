-- 001_initial_schema.sql
-- Consolidated database logic for WebTracker

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
BEGIN
    IF (TG_OP = 'INSERT') OR (NEW.origin IS DISTINCT FROM OLD.origin) OR (NEW.destination IS DISTINCT FROM OLD.destination) THEN
        
        IF NEW.tracking_id IS NULL OR NEW.tracking_id = '' THEN
            NEW.tracking_id := generate_tracking_id();
        END IF;

        SELECT zone_name INTO dest_tz FROM public.country_timezones WHERE country_name = lower(trim(NEW.destination));
        IF dest_tz IS NULL THEN dest_tz := 'UTC'; END IF;

        now_lagos := CURRENT_TIMESTAMP AT TIME ZONE 'Africa/Lagos';
        
        IF extract(hour from now_lagos) >= 22 THEN
            departure_utc := (date_trunc('day', now_lagos + interval '1 day') + interval '8 hours') AT TIME ZONE 'Africa/Lagos';
        ELSE
            departure_utc := CURRENT_TIMESTAMP + interval '1 hour';
        END IF;

        NEW.scheduled_transit_time := departure_utc;

        IF lower(trim(NEW.origin)) IN ('nigeria', 'ghana', 'benin', 'togo', 'niger', 'cameroon') 
           AND lower(trim(NEW.destination)) IN ('nigeria', 'ghana', 'benin', 'togo', 'niger', 'cameroon') THEN
            transit_hours := 4;
        END IF;

        earliest_arrival_utc := departure_utc + (transit_hours * interval '1 hour');

        arrival_local := earliest_arrival_utc AT TIME ZONE dest_tz;
        
        IF dest_tz = 'UTC' THEN snap_hour := 12; END IF;

        IF extract(hour from arrival_local) >= 17 THEN
            arrival_local := date_trunc('day', arrival_local + interval '1 day') + (snap_hour * interval '1 hour');
        ELSIF extract(hour from arrival_local) < snap_hour THEN
            arrival_local := date_trunc('day', arrival_local) + (snap_hour * interval '1 hour');
        END IF;

        final_arrival_utc := arrival_local AT TIME ZONE dest_tz;
        
        NEW.expected_delivery_time := final_arrival_utc;
        NEW.outfordelivery_time := final_arrival_utc - interval '2 hours';
        NEW.updated_at := CURRENT_TIMESTAMP;

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
    r_user_jid TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    WITH updated_transit AS (
        UPDATE Shipment
        SET status = 'intransit', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'pending' AND scheduled_transit_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid
    ),
    updated_out AS (
        UPDATE Shipment
        SET status = 'outfordelivery', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'intransit' AND outfordelivery_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid
    ),
    updated_delivered AS (
        UPDATE Shipment
        SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
        WHERE status = 'outfordelivery' AND expected_delivery_time <= now_utc
        RETURNING tracking_id, status AS new_status, user_jid
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
