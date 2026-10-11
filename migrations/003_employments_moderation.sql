-- +migrate Up
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS moderation_status TEXT NOT NULL DEFAULT 'pending';

UPDATE companies
SET moderation_status = 'approved'
WHERE verified = TRUE AND moderation_status = 'pending';

UPDATE companies SET verified = (moderation_status = 'approved');

CREATE TABLE IF NOT EXISTS employments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    UNIQUE(driver_id, company_id)
);

CREATE INDEX IF NOT EXISTS idx_employments_driver ON employments(driver_id);
CREATE INDEX IF NOT EXISTS idx_employments_company ON employments(company_id);
CREATE INDEX IF NOT EXISTS idx_employments_status ON employments(status);

-- +migrate Down
DROP TABLE IF EXISTS employments;
ALTER TABLE companies DROP COLUMN IF EXISTS moderation_status;
