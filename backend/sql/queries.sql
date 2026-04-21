-- name: GetAllCompanies :many
SELECT id FROM companies;

-- name: GetAllActiveCompanies :many
SELECT * FROM companies WHERE auth_status = 'active';

-- name: GetCompanyByID :one
SELECT * FROM companies WHERE id = $1;

-- name: GetCompanyBySetupToken :one
SELECT * FROM companies WHERE setup_token = $1;

-- name: UpdateCompanySettings :exec
UPDATE companies SET name = $2, admin_email = $3, logo_url = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanyAuthStatus :exec
UPDATE companies SET auth_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateCompanySubscriptionStatus :exec
UPDATE companies 
SET subscription_status = $2, 
    subscription_expiry = CURRENT_TIMESTAMP + INTERVAL '30 days',
    updated_at = CURRENT_TIMESTAMP 
WHERE id = $1;

-- name: CreateCompany :one
INSERT INTO companies (name, admin_email, setup_token) VALUES ($1, $2, $3) RETURNING *;

-- name: RegenerateSetupToken :exec
UPDATE companies SET setup_token = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

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

-- name: RunAgedCleanup :exec
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

-- name: UpdateShipmentFieldSenderName :exec
UPDATE Shipment SET sender_name = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldSenderPhone :exec
UPDATE Shipment SET sender_phone = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldOrigin :exec
UPDATE Shipment SET origin = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldRecipientName :exec
UPDATE Shipment SET recipient_name = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldRecipientPhone :exec
UPDATE Shipment SET recipient_phone = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldRecipientEmail :exec
UPDATE Shipment SET recipient_email = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldRecipientID :exec
UPDATE Shipment SET recipient_id = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldRecipientAddress :exec
UPDATE Shipment SET recipient_address = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldDestination :exec
UPDATE Shipment SET destination = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldCargoType :exec
UPDATE Shipment SET cargo_type = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldScheduledTransitTime :exec
UPDATE Shipment SET scheduled_transit_time = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldExpectedDeliveryTime :exec
UPDATE Shipment SET expected_delivery_time = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

-- name: UpdateShipmentFieldOutfordeliveryTime :exec
UPDATE Shipment SET outfordelivery_time = $3, updated_at = CURRENT_TIMESTAMP WHERE company_id = $1 AND tracking_id = $2;

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
