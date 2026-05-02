-- Plans table: single source of truth for all pricing
CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    desc_key TEXT NOT NULL,
    base_price INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'NGN',
    interval_key TEXT NOT NULL DEFAULT 'monthlyInterval',
    popular BOOLEAN DEFAULT FALSE,
    trial_key TEXT,
    btn_key TEXT NOT NULL,
    features JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial plans
INSERT INTO plans (id, name, name_key, desc_key, base_price, currency, interval_key, popular, trial_key, btn_key, features, sort_order) VALUES
(
    'starter', 'Starter', 'starterName', 'starterDesc',
    1200000, 'NGN', 'monthlyInterval', FALSE, 'sevenDayTrial', 'btnStartTrial',
    '["feat_50_shipments","feat_whatsapp","feat_web_portal","feat_manual_entry","feat_community"]',
    1
),
(
    'pro', 'Pro', 'proName', 'proDesc',
    3000000, 'NGN', 'monthlyInterval', TRUE, NULL, 'btnUpgradePro',
    '["feat_250_shipments","feat_whatsapp","feat_ai_parser","feat_csv_upload","feat_custom_branding","feat_priority_support"]',
    2
),
(
    'enterprise', 'Scale', 'scaleName', 'scaleDesc',
    8500000, 'NGN', 'monthlyInterval', FALSE, NULL, 'btnContactSales',
    '["feat_1000_shipments","feat_all_pro","feat_api_webhook","feat_dedicated_whatsapp","feat_247_support"]',
    3
)
ON CONFLICT (id) DO NOTHING;

-- RLS: Everyone can read active plans (public pricing page)
ALTER TABLE plans ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "public_read_plans" ON plans;
CREATE POLICY "public_read_plans" ON plans
  FOR SELECT
  USING (is_active = TRUE);

-- Realtime so frontend gets instant updates when super admin changes prices
ALTER TABLE plans REPLICA IDENTITY FULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_publication_tables
    WHERE pubname = 'supabase_realtime' AND tablename = 'plans'
  ) THEN
    ALTER PUBLICATION supabase_realtime ADD TABLE plans;
  END IF;
END
$$;
