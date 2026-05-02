-- Drop dead columns from early iterations
ALTER TABLE companies
DROP COLUMN IF EXISTS otp_code,
DROP COLUMN IF EXISTS otp_expires_at;
