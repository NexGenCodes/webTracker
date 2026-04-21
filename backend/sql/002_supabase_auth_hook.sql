-- Migration 002: Supabase Custom Access Token Hook
-- Injects the company_id into the JWT claims for multi-tenant backend isolation.

CREATE OR REPLACE FUNCTION public.custom_access_token_hook(event jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    claims jsonb;
    user_email text;
    comp_id uuid;
BEGIN
    -- Extract the user's email from the event payload
    user_email := event->'claims'->>'email';
    
    -- Lookup the company ID where admin_email matches
    SELECT id INTO comp_id FROM public.companies WHERE admin_email = user_email LIMIT 1;
    
    -- If a company exists, inject it into app_metadata
    IF comp_id IS NOT NULL THEN
        claims := event->'claims';
        
        -- Update app_metadata to include company_id
        claims := jsonb_set(
            claims, 
            '{app_metadata, company_id}', 
            to_jsonb(comp_id)
        );
        
        -- Update the event with the modified claims
        event := jsonb_set(event, '{claims}', claims);
    END IF;
    
    RETURN event;
END;
$$;

GRANT EXECUTE ON FUNCTION public.custom_access_token_hook TO supabase_auth_admin;
GRANT EXECUTE ON FUNCTION public.custom_access_token_hook TO postgres;

-- IMPORTANT: 
-- You must manually enable this hook in the Supabase Dashboard:
-- Authentication -> Hooks -> "Custom access token (JWT) hook"
