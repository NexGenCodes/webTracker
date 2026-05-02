-- Add performance indexes for multi-tenant scalability
CREATE INDEX IF NOT EXISTS idx_shipment_company_created ON Shipment(company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipment_company_status ON Shipment(company_id, status);
CREATE INDEX IF NOT EXISTS idx_shipment_company_user ON Shipment(company_id, user_jid);
CREATE INDEX IF NOT EXISTS idx_telemetry_company_created ON Telemetry(company_id, created_at DESC);
