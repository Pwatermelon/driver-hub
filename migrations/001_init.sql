-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    inn TEXT NOT NULL UNIQUE,
    ogrn TEXT,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,
    password_hash TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_id TEXT NOT NULL UNIQUE,
    esia_oid TEXT UNIQUE,
    email TEXT,
    phone TEXT,
    last_name TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    middle_name TEXT,
    birth_date DATE,
    snils TEXT,
    inn TEXT,
    experience_years INT NOT NULL DEFAULT 0,
    categories TEXT[] NOT NULL DEFAULT '{}',
    verification_status TEXT NOT NULL DEFAULT 'none',
    verified_at TIMESTAMPTZ,
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_drivers_hub_id ON drivers(hub_id);
CREATE INDEX IF NOT EXISTS idx_drivers_snils ON drivers(snils);
CREATE INDEX IF NOT EXISTS idx_drivers_esia_oid ON drivers(esia_oid);

CREATE TABLE IF NOT EXISTS driver_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL UNIQUE REFERENCES drivers(id) ON DELETE CASCADE,
    series TEXT NOT NULL DEFAULT '',
    number TEXT NOT NULL DEFAULT '',
    issue_date DATE,
    expiry_date DATE,
    categories TEXT[] NOT NULL DEFAULT '{}',
    issuer TEXT,
    source TEXT NOT NULL DEFAULT 'manual',
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    raw_esia_payload TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medical_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    number TEXT NOT NULL DEFAULT '',
    issued_at DATE,
    valid_until DATE,
    clinic TEXT,
    result TEXT NOT NULL DEFAULT 'fit',
    categories TEXT[] NOT NULL DEFAULT '{}',
    file_url TEXT,
    source TEXT NOT NULL DEFAULT 'upload',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS license_suspensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    start_date DATE,
    end_date DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    authority TEXT,
    case_number TEXT,
    source TEXT NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS criminal_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    has_record BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    checked_at DATE,
    valid_until DATE,
    source TEXT NOT NULL DEFAULT 'manual',
    file_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS access_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    purpose TEXT NOT NULL DEFAULT 'employment',
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(driver_id, company_id)
);

CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS complaints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    category TEXT NOT NULL DEFAULT 'other',
    text TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'medium',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blacklist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lifted_at TIMESTAMPTZ,
    UNIQUE(driver_id, company_id)
);

CREATE TABLE IF NOT EXISTS accidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    description TEXT NOT NULL,
    fault TEXT NOT NULL DEFAULT 'unknown',
    damage_level TEXT NOT NULL DEFAULT 'minor',
    location TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    article TEXT,
    amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    issued_at DATE,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    source TEXT NOT NULL DEFAULT 'company',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS esia_sessions (
    state TEXT PRIMARY KEY,
    nonce TEXT NOT NULL,
    driver_id UUID REFERENCES drivers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    actor_type TEXT NOT NULL,
    actor_id UUID,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id UUID,
    meta JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS esia_sessions;
DROP TABLE IF EXISTS fines;
DROP TABLE IF EXISTS accidents;
DROP TABLE IF EXISTS blacklist_entries;
DROP TABLE IF EXISTS complaints;
DROP TABLE IF EXISTS recommendations;
DROP TABLE IF EXISTS access_grants;
DROP TABLE IF EXISTS criminal_records;
DROP TABLE IF EXISTS license_suspensions;
DROP TABLE IF EXISTS medical_certificates;
DROP TABLE IF EXISTS driver_licenses;
DROP TABLE IF EXISTS drivers;
DROP TABLE IF EXISTS companies;
