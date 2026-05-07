-- name: GetAllCompanies :many
SELECT id FROM companies;

-- name: GetAllActiveCompanies :many
SELECT * FROM companies 
WHERE auth_status = 'active' 
  AND subscription_status IN ('active', 'trialing')
  AND (subscription_expiry IS NULL OR subscription_expiry > CURRENT_TIMESTAMP);

-- name: GetCompanyByID :one
SELECT * FROM companies WHERE id = $1;



-- name: UpdateCompanySettings :exec
UPDATE companies SET name = $2, admin_email = $3, logo_url = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanyAuthStatus :exec
UPDATE companies SET auth_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanySubscriptionStatus :exec
UPDATE companies 
SET subscription_status = $2, 
    subscription_expiry = GREATEST(subscription_expiry, CURRENT_TIMESTAMP) + INTERVAL '30 days',
    updated_at = CURRENT_TIMESTAMP 
WHERE id = $1;

-- name: CreateCompany :one
INSERT INTO companies (name, admin_email, setup_token, subscription_expiry, plan_type) 
VALUES ($1, $2, $3, CURRENT_TIMESTAMP + INTERVAL '7 days', 'trial') 
RETURNING *;



-- name: GetSystemConfig :one
SELECT value FROM SystemConfig WHERE company_id = $1 AND key = $2;

-- name: SetSystemConfig :exec
INSERT INTO SystemConfig (company_id, key, value, updated_at) 
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT(company_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP;

-- name: GetUserLanguage :one
SELECT language FROM UserPreference WHERE company_id = $1 AND jid = $2;

-- name: SetUserLanguage :exec
INSERT INTO UserPreference (company_id, jid, language, updated_at) 
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT(company_id, jid) DO UPDATE SET language = EXCLUDED.language, updated_at = CURRENT_TIMESTAMP;

-- name: GetGroupAuthority :one
SELECT is_authorized, updated_at FROM GroupAuthority WHERE company_id = $1 AND jid = $2;

-- name: SetGroupAuthority :exec
INSERT INTO GroupAuthority (company_id, jid, is_authorized, updated_at) 
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT(company_id, jid) DO UPDATE SET is_authorized = EXCLUDED.is_authorized, updated_at = CURRENT_TIMESTAMP;

-- name: HasAuthorizedGroups :one
SELECT COUNT(*) FROM GroupAuthority WHERE company_id = $1 AND is_authorized = true;

-- name: GetAuthorizedGroups :many
SELECT jid FROM GroupAuthority WHERE company_id = $1 AND is_authorized = true;

-- name: CountAuthorizedGroups :one
SELECT COUNT(*) FROM GroupAuthority WHERE company_id = $1 AND is_authorized = true;

-- name: RunAgedCleanup :execresult
DELETE FROM Shipment 
WHERE company_id = $1 AND ((status = 'delivered' AND updated_at < $2) OR (created_at < $3));

-- name: CreateShipment :exec
INSERT INTO Shipment (
    company_id, tracking_id, user_jid, status, created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time, sender_timezone, recipient_timezone, sender_name, sender_phone, origin, recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination, cargo_type, weight, cost, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
);

-- name: GetShipment :one
SELECT * FROM Shipment WHERE company_id = $1 AND tracking_id = $2;

-- name: ListShipments :many
SELECT * FROM Shipment WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateShipmentStatus :exec
UPDATE Shipment SET status = $3, destination = $4, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: DeleteShipment :exec
DELETE FROM Shipment WHERE company_id = $1 AND tracking_id = $2;

-- name: BulkDeleteShipments :execresult
DELETE FROM Shipment WHERE company_id = $1 AND tracking_id = ANY($2::text[]);

-- name: DeleteDeliveredShipments :exec
DELETE FROM Shipment WHERE company_id = $1 AND status = 'delivered';

-- name: TransitionStatusToIntransit :many
UPDATE Shipment
SET status = 'intransit', updated_at = CURRENT_TIMESTAMP
WHERE company_id = $1 AND status = 'pending' AND scheduled_transit_time <= $2
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: TransitionStatusToOutForDelivery :many
UPDATE Shipment
SET status = 'outfordelivery', updated_at = CURRENT_TIMESTAMP
WHERE company_id = $1 AND status = 'intransit' AND outfordelivery_time <= $2
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: TransitionStatusToDelivered :many
UPDATE Shipment
SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
WHERE company_id = $1 AND status = 'outfordelivery' AND expected_delivery_time <= $2
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: GetLastShipmentIDForUser :one
SELECT tracking_id FROM Shipment WHERE company_id = $1 AND user_jid = $2 ORDER BY created_at DESC LIMIT 1;

-- name: FindSimilarShipment :one
SELECT tracking_id FROM Shipment 
WHERE company_id = $1 AND user_jid = $2 AND recipient_phone = $3 AND $3 != ''
ORDER BY created_at DESC LIMIT 1;

-- name: CountCreatedSince :one
SELECT COUNT(*) FROM Shipment WHERE company_id = $1 AND created_at >= $2;

-- name: CountDeliveredSince :one
SELECT COUNT(*) FROM Shipment WHERE company_id = $1 AND status = 'delivered' AND updated_at >= $2;

-- name: ListAllShipments :many
SELECT * FROM Shipment WHERE company_id = $1 ORDER BY created_at DESC;

-- name: CountShipments :one
SELECT COUNT(*) FROM Shipment WHERE company_id = $1;

-- name: CountShipmentsByStatus :one
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'pending') AS pending,
    COUNT(*) FILTER (WHERE status = 'intransit') AS intransit,
    COUNT(*) FILTER (WHERE status = 'outfordelivery') AS outfordelivery,
    COUNT(*) FILTER (WHERE status = 'delivered') AS delivered,
    COUNT(*) FILTER (WHERE status = 'canceled') AS canceled
FROM Shipment WHERE company_id = $1;



-- name: UpdateShipmentDynamic :exec
UPDATE Shipment
SET 
  sender_name = COALESCE(NULLIF($3::text, ''), sender_name),
  sender_phone = COALESCE(NULLIF($4::text, ''), sender_phone),
  origin = COALESCE(NULLIF($5::text, ''), origin),
  recipient_name = COALESCE(NULLIF($6::text, ''), recipient_name),
  recipient_phone = COALESCE(NULLIF($7::text, ''), recipient_phone),
  recipient_email = COALESCE(NULLIF($8::text, ''), recipient_email),
  recipient_id = COALESCE(NULLIF($9::text, ''), recipient_id),
  recipient_address = COALESCE(NULLIF($10::text, ''), recipient_address),
  destination = COALESCE(NULLIF($11::text, ''), destination),
  cargo_type = COALESCE(NULLIF($12::text, ''), cargo_type),
  scheduled_transit_time = COALESCE(NULLIF($13::timestamp, '0001-01-01 00:00:00'::timestamp), scheduled_transit_time),
  expected_delivery_time = COALESCE(NULLIF($14::timestamp, '0001-01-01 00:00:00'::timestamp), expected_delivery_time),
  outfordelivery_time = COALESCE(NULLIF($15::timestamp, '0001-01-01 00:00:00'::timestamp), outfordelivery_time),
  status = COALESCE(NULLIF($16::text, ''), status),
  updated_at = CURRENT_TIMESTAMP
WHERE company_id = $1 AND tracking_id = $2;

-- name: RecordEvent :exec
INSERT INTO Telemetry (company_id, event_type, metadata, created_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP);

-- name: GetTelemetryStats :many
SELECT event_type, COUNT(*) as count
FROM Telemetry
WHERE company_id = $1 AND created_at >= $2
GROUP BY event_type;

-- name: GetRecentEvents :many
SELECT * FROM Telemetry
WHERE company_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: BulkUpdateStatus :exec
UPDATE Shipment
SET status = $3, updated_at = CURRENT_TIMESTAMP
WHERE company_id = $1 AND tracking_id = ANY($2::text[]);

-- name: GetCompanyByEmail :one
SELECT * FROM companies WHERE admin_email = $1;

-- name: SetCompanyPassword :exec
UPDATE companies SET admin_password_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;


-- name: UpdateCompanyOnboarding :exec
UPDATE companies SET whatsapp_phone = $2, tracking_prefix = $3, auth_status = 'active', updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: RecordPayment :one
INSERT INTO payments (company_id, reference, amount, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (reference) DO NOTHING
RETURNING id;

-- name: GetActivePlans :many
SELECT id, name, name_key, desc_key, base_price, currency, interval_key, popular, trial_key, btn_key, features, sort_order
FROM plans
WHERE is_active = TRUE
ORDER BY sort_order ASC;

-- name: GetPlanByID :one
SELECT id, name, name_key, desc_key, base_price, currency, interval_key, popular, trial_key, btn_key, features
FROM plans
WHERE id = $1 AND is_active = TRUE;

-- name: UpdatePlanPrice :exec
UPDATE plans SET base_price = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanyPlan :exec
UPDATE companies SET plan_type = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanySubscription :exec
UPDATE companies SET subscription_status = $2, subscription_expiry = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: LogAudit :exec
INSERT INTO audit_log (actor_email, action, target_company_id, details)
VALUES ($1, $2, $3, $4);

-- name: GetAuditLogs :many
SELECT * FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPlatformAnalytics :one
SELECT 
    (SELECT COUNT(*) FROM companies) as total_tenants,
    (SELECT COUNT(*) FROM companies WHERE created_at >= date_trunc('month', CURRENT_TIMESTAMP)) as new_tenants_this_month,
    (SELECT COUNT(*) FROM Shipment) as total_shipments,
    (SELECT COUNT(*) FROM Shipment WHERE created_at >= CURRENT_DATE) as shipments_today,
    (SELECT jsonb_object_agg(plan_type, count) FROM (SELECT plan_type, COUNT(*) as count FROM companies GROUP BY plan_type) t) as plan_distribution,
    (SELECT jsonb_object_agg(subscription_status, count) FROM (SELECT subscription_status, COUNT(*) as count FROM companies GROUP BY subscription_status) t) as subscription_distribution;
-- name: GetCompanyPayments :many
SELECT * FROM payments WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
