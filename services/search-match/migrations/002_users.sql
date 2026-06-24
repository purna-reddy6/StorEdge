-- Migration 002: Users and authentication

CREATE TYPE user_role AS ENUM (
  'farmer',
  'trader',       -- E-commerce / retail merchants
  'operator',     -- Warehouse owners/managers
  'admin',        -- Platform administrators
  'logistics'     -- Transport fleet operators
);

CREATE TYPE kyc_status AS ENUM (
  'pending',
  'submitted',
  'verified',
  'rejected'
);

CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  phone           VARCHAR(15) UNIQUE NOT NULL,       -- Primary identifier (India: +91XXXXXXXXXX)
  email           VARCHAR(255) UNIQUE,
  name            VARCHAR(255) NOT NULL,
  role            user_role NOT NULL DEFAULT 'trader',
  kyc_status      kyc_status NOT NULL DEFAULT 'pending',
  aadhaar_number  VARCHAR(12),                       -- Encrypted at app layer
  pan_number      VARCHAR(10),
  gst_number      VARCHAR(15),
  language_pref   VARCHAR(10) NOT NULL DEFAULT 'en', -- 'en', 'hi', 'gu', 'mr'
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  last_login_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  VARCHAR(255) NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE otp_requests (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  phone       VARCHAR(15) NOT NULL,
  otp_hash    VARCHAR(255) NOT NULL,
  purpose     VARCHAR(50) NOT NULL,  -- 'login' | 'stock_release' | 'enwrs_pledge'
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_otp_requests_phone ON otp_requests(phone);
