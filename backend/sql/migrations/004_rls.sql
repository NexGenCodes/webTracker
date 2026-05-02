-- ==========================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- ==========================================
ALTER TABLE shipment ENABLE ROW LEVEL SECURITY;

-- Allow public read of tracking data via anonymous key
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies 
        WHERE tablename = 'shipment' AND policyname = 'Public can read shipments by tracking_id'
    ) THEN
        CREATE POLICY "Public can read shipments by tracking_id" 
        ON shipment FOR SELECT 
        USING (true);
    END IF;
END $$;
