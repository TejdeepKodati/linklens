-- ============================================================
--  LinkLens — Initial Schema
--  Run: psql $DATABASE_URL -f migrations/001_initial.sql
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── Users ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    name          VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ── URLs ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS urls (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT         NOT NULL,
    short_code   VARCHAR(50)  UNIQUE NOT NULL,
    title        VARCHAR(500) NOT NULL DEFAULT '',
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_urls_user_id    ON urls(user_id);
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);

-- ── Clicks ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS clicks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    url_id      UUID        NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    ip_address  INET,
    user_agent  TEXT        NOT NULL DEFAULT '',
    referer     TEXT        NOT NULL DEFAULT '',
    country     VARCHAR(10) NOT NULL DEFAULT '',
    device_type VARCHAR(50) NOT NULL DEFAULT 'desktop',
    browser     VARCHAR(100) NOT NULL DEFAULT '',
    os          VARCHAR(100) NOT NULL DEFAULT '',
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clicks_url_id     ON clicks(url_id);
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at DESC);
CREATE INDEX IF NOT EXISTS idx_clicks_url_date   ON clicks(url_id, clicked_at DESC);
