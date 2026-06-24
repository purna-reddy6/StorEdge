-- Migration 004: Bookings, inventory, and pallet tracking

CREATE TYPE booking_status AS ENUM (
  'pending',     -- Awaiting payment
  'confirmed',   -- Payment captured
  'active',      -- Goods are in warehouse
  'completed',   -- Goods fully removed
  'cancelled',   -- Cancelled before activation
  'disputed'     -- Under dispute resolution
);

CREATE TYPE commodity_type AS ENUM (
  'potato', 'wheat', 'paddy', 'onion', 'garlic',
  'fruits', 'vegetables', 'pulses', 'oilseeds', 'cereals',
  'pharma', 'dairy',
  'apparel', 'electronics', 'fmcg', 'auto_parts',
  'other'
);

CREATE TABLE bookings (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  booking_number          VARCHAR(20) NOT NULL UNIQUE,  -- SE-2026-XXXXXX
  tenant_id               UUID NOT NULL REFERENCES users(id),
  warehouse_id            UUID NOT NULL REFERENCES warehouses(id),

  -- Capacity
  pallet_count            INTEGER NOT NULL CHECK (pallet_count > 0),
  commodity_type          commodity_type NOT NULL DEFAULT 'other',
  commodity_description   VARCHAR(500),
  weight_kg               NUMERIC(12,2),
  volume_cubic_meters     NUMERIC(10,3),

  -- Pricing (locked at booking time)
  price_per_pallet_inr    NUMERIC(10,2) NOT NULL,
  total_amount_inr        NUMERIC(12,2) NOT NULL,
  commission_amount_inr   NUMERIC(12,2) NOT NULL,  -- 15% platform fee
  payout_amount_inr       NUMERIC(12,2) NOT NULL,  -- To warehouse owner

  -- Dates
  start_date              DATE NOT NULL,
  end_date                DATE NOT NULL,
  actual_inward_date      TIMESTAMPTZ,
  actual_outward_date     TIMESTAMPTZ,

  -- Status
  status                  booking_status NOT NULL DEFAULT 'pending',

  -- Payment
  payment_intent_id       VARCHAR(255),  -- Stripe payment intent
  payment_captured_at     TIMESTAMPTZ,

  -- Notes
  special_instructions    TEXT,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_tenant ON bookings(tenant_id);
CREATE INDEX idx_bookings_warehouse ON bookings(warehouse_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_dates ON bookings(start_date, end_date);

-- Pallet-level inventory tracking
CREATE TABLE pallet_items (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  booking_id              UUID NOT NULL REFERENCES bookings(id),
  warehouse_id            UUID NOT NULL REFERENCES warehouses(id),
  tenant_id               UUID NOT NULL REFERENCES users(id),

  commodity_type          commodity_type NOT NULL,
  commodity_description   VARCHAR(500),
  weight_kg               NUMERIC(10,2) NOT NULL,
  volume_cubic_meters     NUMERIC(8,3),
  bag_count               INTEGER,

  -- Physical slot assignment (from WMS slotting engine)
  slot_position           VARCHAR(20),  -- "A-03-02" (aisle-row-level)
  rfid_tag_id             VARCHAR(100),

  -- e-NWR linkage
  enwrs_pledged           BOOLEAN NOT NULL DEFAULT FALSE,
  enwrs_receipt_id        VARCHAR(100),

  -- Dates
  inward_date             TIMESTAMPTZ,
  expected_outward_date   DATE,
  actual_outward_date     TIMESTAMPTZ,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pallet_items_booking ON pallet_items(booking_id);
CREATE INDEX idx_pallet_items_warehouse ON pallet_items(warehouse_id);
CREATE INDEX idx_pallet_items_rfid ON pallet_items(rfid_tag_id);
CREATE INDEX idx_pallet_items_enwrs ON pallet_items(enwrs_receipt_id) WHERE enwrs_pledged = TRUE;

-- Stock release requests (OTP-based)
CREATE TYPE release_status AS ENUM ('pending_otp', 'otp_sent', 'authorized', 'rejected', 'completed');

CREATE TABLE stock_release_requests (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  pallet_item_id          UUID NOT NULL REFERENCES pallet_items(id),
  tenant_id               UUID NOT NULL REFERENCES users(id),
  warehouse_id            UUID NOT NULL REFERENCES warehouses(id),

  quantity_to_release_kg  NUMERIC(10,2) NOT NULL,
  release_reason          VARCHAR(255),

  status                  release_status NOT NULL DEFAULT 'pending_otp',
  otp_request_id          UUID REFERENCES otp_requests(id),

  authorized_at           TIMESTAMPTZ,
  authorized_by_operator  UUID REFERENCES users(id),
  completed_at            TIMESTAMPTZ,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_release_requests_pallet ON stock_release_requests(pallet_item_id);
CREATE INDEX idx_release_requests_tenant ON stock_release_requests(tenant_id);
CREATE INDEX idx_release_requests_status ON stock_release_requests(status);

-- Warehouse reviews
CREATE TABLE warehouse_reviews (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
  reviewer_id     UUID NOT NULL REFERENCES users(id),
  booking_id      UUID NOT NULL REFERENCES bookings(id),
  rating          INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
  comment         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(booking_id)  -- One review per booking
);

CREATE INDEX idx_reviews_warehouse ON warehouse_reviews(warehouse_id);
