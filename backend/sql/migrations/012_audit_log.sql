-- Create audit log table for tracking super admin actions
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    actor_email TEXT NOT NULL,
    action TEXT NOT NULL,        -- 'delete_company', 'change_plan', etc.
    target_company_id UUID,
    details JSONB,               -- flexible payload
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
