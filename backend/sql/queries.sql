-- name: GetSystemConfig :one
SELECT value FROM SystemConfig WHERE key = $1;

-- name: SetSystemConfig :exec
INSERT INTO SystemConfig (key, value, updated_at) 
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP;

-- name: GetUserLanguage :one
SELECT language FROM UserPreference WHERE jid = $1;

-- name: SetUserLanguage :exec
INSERT INTO UserPreference (jid, language, updated_at) 
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT(jid) DO UPDATE SET language = EXCLUDED.language, updated_at = CURRENT_TIMESTAMP;

-- name: GetGroupAuthority :one
SELECT is_authorized, updated_at FROM GroupAuthority WHERE jid = $1;

-- name: SetGroupAuthority :exec
INSERT INTO GroupAuthority (jid, is_authorized, updated_at) 
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT(jid) DO UPDATE SET is_authorized = EXCLUDED.is_authorized, updated_at = CURRENT_TIMESTAMP;

-- name: HasAuthorizedGroups :one
SELECT COUNT(*) FROM GroupAuthority WHERE is_authorized = true;

-- name: GetAuthorizedGroups :many
SELECT jid FROM GroupAuthority WHERE is_authorized = true;

-- name: CountAuthorizedGroups :one
SELECT COUNT(*) FROM GroupAuthority WHERE is_authorized = true;

-- name: RunAgedCleanup :exec
DELETE FROM Shipment 
WHERE (status = 'delivered' AND updated_at < $1) OR (created_at < $2);

-- name: CreateShipment :exec
INSERT INTO Shipment (
    tracking_id, user_jid, status, created_at, scheduled_transit_time, outfordelivery_time, expected_delivery_time, sender_timezone, recipient_timezone, sender_name, sender_phone, origin, recipient_name, recipient_phone, recipient_email, recipient_id, recipient_address, destination, cargo_type, weight, cost, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
);

-- name: GetShipment :one
SELECT * FROM Shipment WHERE tracking_id = $1;

-- name: ListShipments :many
SELECT * FROM Shipment ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateShipmentStatus :exec
UPDATE Shipment SET status = $2, destination = $3, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: DeleteShipment :exec
DELETE FROM Shipment WHERE tracking_id = $1;

-- name: DeleteDeliveredShipments :exec
DELETE FROM Shipment WHERE status = 'delivered';

-- name: TransitionStatusToIntransit :many
UPDATE Shipment
SET status = 'intransit', updated_at = CURRENT_TIMESTAMP
WHERE status = 'pending' AND scheduled_transit_time <= $1
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: TransitionStatusToOutForDelivery :many
UPDATE Shipment
SET status = 'outfordelivery', updated_at = CURRENT_TIMESTAMP
WHERE status = 'intransit' AND outfordelivery_time <= $1
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: TransitionStatusToDelivered :many
UPDATE Shipment
SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
WHERE status = 'outfordelivery' AND expected_delivery_time <= $1
RETURNING tracking_id, status AS new_status, user_jid, recipient_email;

-- name: GetLastShipmentIDForUser :one
SELECT tracking_id FROM Shipment WHERE user_jid = $1 ORDER BY created_at DESC LIMIT 1;

-- name: FindSimilarShipment :one
SELECT tracking_id FROM Shipment 
WHERE user_jid = $1 
AND (
    (recipient_phone = $2 AND $2 != '') OR 
    (recipient_email = $3 AND $3 != '') OR 
    (recipient_id = $4 AND $4 != '')
)
ORDER BY created_at DESC LIMIT 1;

-- name: CountCreatedSince :one
SELECT COUNT(*) FROM Shipment WHERE created_at >= $1;

-- name: CountDeliveredSince :one
SELECT COUNT(*) FROM Shipment WHERE status = 'delivered' AND updated_at >= $1;

-- name: ListAllShipments :many
SELECT * FROM Shipment ORDER BY created_at DESC;

-- name: CountShipments :one
SELECT COUNT(*) FROM Shipment;

-- name: CountShipmentsByStatus :one
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'pending') AS pending,
    COUNT(*) FILTER (WHERE status = 'intransit') AS intransit,
    COUNT(*) FILTER (WHERE status = 'outfordelivery') AS outfordelivery,
    COUNT(*) FILTER (WHERE status = 'delivered') AS delivered,
    COUNT(*) FILTER (WHERE status = 'canceled') AS canceled
FROM Shipment;

-- name: UpdateShipmentFieldSenderName :exec
UPDATE Shipment SET sender_name = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldSenderPhone :exec
UPDATE Shipment SET sender_phone = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldOrigin :exec
UPDATE Shipment SET origin = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldRecipientName :exec
UPDATE Shipment SET recipient_name = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldRecipientPhone :exec
UPDATE Shipment SET recipient_phone = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldRecipientEmail :exec
UPDATE Shipment SET recipient_email = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldRecipientID :exec
UPDATE Shipment SET recipient_id = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldRecipientAddress :exec
UPDATE Shipment SET recipient_address = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldDestination :exec
UPDATE Shipment SET destination = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldCargoType :exec
UPDATE Shipment SET cargo_type = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldScheduledTransitTime :exec
UPDATE Shipment SET scheduled_transit_time = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldExpectedDeliveryTime :exec
UPDATE Shipment SET expected_delivery_time = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;

-- name: UpdateShipmentFieldOutfordeliveryTime :exec
UPDATE Shipment SET outfordelivery_time = $2, updated_at = CURRENT_TIMESTAMP WHERE tracking_id = $1;
