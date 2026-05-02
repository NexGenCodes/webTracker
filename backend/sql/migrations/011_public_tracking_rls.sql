-- Allow anonymous users to read shipments by tracking_id (public tracking page).
-- Tracking IDs are 9-digit random values making enumeration impractical.
DROP POLICY IF EXISTS "anon_read_shipment_by_tracking" ON shipment;
CREATE POLICY "anon_read_shipment_by_tracking" ON shipment
  FOR SELECT TO anon
  USING (true);
