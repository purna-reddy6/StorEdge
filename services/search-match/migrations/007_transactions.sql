-- Migration 007: Financial transactions, payouts, and dispute escrow

CREATE TYPE transaction_type AS ENUM (
  'booking_payment',      -- Tenant pays for storage
  'platform_commission',  -- 15% StorEdge commission
  'operator_payout',      -- Payment to warehouse owner
  'logistics_fee',        -- 8% on freight bookings
  'enwrs_origination_fee', -- 1.5% on loan origination
  'insurance_commission', -- 20% on insurance purchases
  'refund',
  'dispute_escrow_hold',
  'dispute_escrow_release'
);

CREATE TYPE transaction_status AS ENUM (
  'pending',
  'processing',
  'completed',
  'failed',
  'refunded'
);

CREATE TABLE transactions (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  reference_number        VARCHAR(30) NOT NULL UNIQUE,  -- SE-TXN-XXXXXXXXX
  transaction_type        transaction_type NOT NULL,

  -- Parties
  from_user_id            UUID REFERENCES users(id),
  to_user_id              UUID REFERENCES users(id),

  -- Linked entities
  booking_id              UUID REFERENCES bookings(id),
  loan_id                 UUID REFERENCES loan_applications(id),

  -- Amounts
  amount_inr              NUMERIC(14,2) NOT NULL,
  currency                VARCHAR(3) NOT NULL DEFAULT 'INR',

  -- Payment gateway
  payment_gateway         VARCHAR(20) DEFAULT 'stripe',  -- 'stripe' | 'razorpay'
  gateway_reference       VARCHAR(255),
  gateway_fee_inr         NUMERIC(10,2),

  status                  transaction_status NOT NULL DEFAULT 'pending',
  processed_at            TIMESTAMPTZ,
  failed_reason           TEXT,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_from ON transactions(from_user_id);
CREATE INDEX idx_transactions_to ON transactions(to_user_id);
CREATE INDEX idx_transactions_booking ON transactions(booking_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

-- Logistics / freight bookings
CREATE TYPE freight_status AS ENUM (
  'requested', 'matched', 'assigned', 'in_transit', 'delivered', 'cancelled'
);

CREATE TABLE freight_bookings (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  booking_number          VARCHAR(20) NOT NULL UNIQUE,
  requester_id            UUID NOT NULL REFERENCES users(id),
  driver_id               UUID REFERENCES users(id),
  booking_id              UUID REFERENCES bookings(id),   -- Linked warehouse booking

  -- Route
  pickup_address          TEXT NOT NULL,
  pickup_geo              GEOMETRY(Point, 4326),
  delivery_address        TEXT NOT NULL,
  delivery_geo            GEOMETRY(Point, 4326),
  distance_km             NUMERIC(8,2),

  -- Cargo
  commodity_type          commodity_type NOT NULL,
  weight_kg               NUMERIC(12,2) NOT NULL,
  requires_reefer         BOOLEAN DEFAULT FALSE,
  reefer_temp_min         NUMERIC(5,2),
  reefer_temp_max         NUMERIC(5,2),

  -- Pricing
  agreed_freight_inr      NUMERIC(12,2),
  dispatch_fee_inr        NUMERIC(10,2),   -- 8% StorEdge fee
  driver_payout_inr       NUMERIC(12,2),

  status                  freight_status NOT NULL DEFAULT 'requested',
  picked_up_at            TIMESTAMPTZ,
  delivered_at            TIMESTAMPTZ,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_freight_requester ON freight_bookings(requester_id);
CREATE INDEX idx_freight_driver ON freight_bookings(driver_id);
CREATE INDEX idx_freight_status ON freight_bookings(status);
CREATE INDEX idx_freight_pickup_geo ON freight_bookings USING GIST(pickup_geo);
CREATE INDEX idx_freight_delivery_geo ON freight_bookings USING GIST(delivery_geo);

-- Dispute management
CREATE TYPE dispute_status AS ENUM (
  'opened', 'under_review', 'resolved_tenant', 'resolved_operator', 'escalated', 'closed'
);

CREATE TABLE disputes (
  id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  booking_id          UUID NOT NULL REFERENCES bookings(id),
  raised_by           UUID NOT NULL REFERENCES users(id),
  against_user        UUID NOT NULL REFERENCES users(id),
  dispute_type        VARCHAR(50) NOT NULL,  -- 'spoilage' | 'access_denied' | 'billing' | 'quality'
  description         TEXT NOT NULL,
  evidence_urls       TEXT[],
  escrow_amount_inr   NUMERIC(12,2),
  status              dispute_status NOT NULL DEFAULT 'opened',
  resolution_notes    TEXT,
  resolved_at         TIMESTAMPTZ,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disputes_booking ON disputes(booking_id);
CREATE INDEX idx_disputes_raised_by ON disputes(raised_by);
CREATE INDEX idx_disputes_status ON disputes(status);
