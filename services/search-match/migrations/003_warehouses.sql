-- Migration 003: Warehouses and physical facility metadata

CREATE TYPE warehouse_type AS ENUM (
  'cold_storage',        -- Temperature-controlled agricultural
  'ambient',             -- Standard commercial dry warehouse
  'bonded',              -- Customs bonded
  'hazmat',              -- Hazardous materials
  'self_storage',        -- P2P residential/small commercial
  'micro_fulfillment'    -- E-commerce last-mile hub
);

CREATE TYPE wdra_status AS ENUM (
  'registered',    -- Active WDRA registration — can issue e-NWRs
  'pending',       -- Registration application submitted
  'unregistered'   -- Not WDRA registered
);

CREATE TYPE warehouse_status AS ENUM (
  'active',
  'inactive',
  'under_maintenance',
  'suspended'
);

CREATE TABLE warehouses (
  id                              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner_id                        UUID NOT NULL REFERENCES users(id),
  name                            VARCHAR(255) NOT NULL,
  description                     TEXT,
  type                            warehouse_type NOT NULL,
  status                          warehouse_status NOT NULL DEFAULT 'active',

  -- Location (PostGIS)
  geo_location                    GEOMETRY(Point, 4326) NOT NULL,
  address_line1                   VARCHAR(255) NOT NULL,
  address_line2                   VARCHAR(255),
  city                            VARCHAR(100) NOT NULL,
  state                           VARCHAR(100) NOT NULL DEFAULT 'Uttar Pradesh',
  pincode                         VARCHAR(10) NOT NULL,
  district                        VARCHAR(100),

  -- Capacity
  total_pallet_capacity           INTEGER NOT NULL CHECK (total_pallet_capacity > 0),
  available_pallet_slots          INTEGER NOT NULL,
  floor_area_sqft                 NUMERIC(10,2),
  ceiling_height_ft               NUMERIC(6,2),

  -- Temperature specs (cold storage)
  min_temperature_celsius         NUMERIC(5,2),
  max_temperature_celsius         NUMERIC(5,2),

  -- Pricing
  base_price_per_pallet_inr       NUMERIC(10,2) NOT NULL,  -- INR/pallet/month
  current_dynamic_price_inr       NUMERIC(10,2),            -- Set by pricing engine
  min_booking_days                INTEGER NOT NULL DEFAULT 1,

  -- Compliance
  wdra_status                     wdra_status NOT NULL DEFAULT 'unregistered',
  wdra_registration_number        VARCHAR(100),
  wdra_expiry_date                DATE,
  apmc_license_number             VARCHAR(100),
  fssai_license_number            VARCHAR(100),
  gst_number                      VARCHAR(15),
  fire_noc_valid_until            DATE,
  pest_control_valid_until        DATE,

  -- Features (bitmask via columns)
  has_cctv                        BOOLEAN DEFAULT FALSE,
  has_weighbridge                 BOOLEAN DEFAULT FALSE,
  has_loading_dock                INTEGER DEFAULT 0,  -- Number of docks
  has_reefer_connectivity         BOOLEAN DEFAULT FALSE,
  has_solar_power                 BOOLEAN DEFAULT FALSE,
  has_backup_power                BOOLEAN DEFAULT FALSE,

  -- Ratings
  rating                          NUMERIC(3,2) DEFAULT 0.00,
  total_reviews                   INTEGER DEFAULT 0,

  -- Media
  cover_image_url                 VARCHAR(500),
  image_urls                      TEXT[],

  -- Metadata
  is_active                       BOOLEAN NOT NULL DEFAULT TRUE,
  onboarded_at                    TIMESTAMPTZ,
  created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PostGIS spatial index — critical for geo-matching queries
CREATE INDEX idx_warehouses_geo ON warehouses USING GIST(geo_location);
CREATE INDEX idx_warehouses_owner ON warehouses(owner_id);
CREATE INDEX idx_warehouses_type ON warehouses(type);
CREATE INDEX idx_warehouses_wdra ON warehouses(wdra_status);
CREATE INDEX idx_warehouses_status ON warehouses(status);
CREATE INDEX idx_warehouses_city ON warehouses(city);
CREATE INDEX idx_warehouses_pincode ON warehouses(pincode);

-- Full-text search index on name + description
CREATE INDEX idx_warehouses_fts ON warehouses
  USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Warehouse floor plan (drag-and-drop grid)
CREATE TABLE warehouse_floor_plans (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  warehouse_id    UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
  grid_rows       INTEGER NOT NULL,
  grid_columns    INTEGER NOT NULL,
  grid_data       JSONB NOT NULL DEFAULT '{}',  -- Slot occupancy map
  svg_layout_url  VARCHAR(500),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_floor_plans_warehouse ON warehouse_floor_plans(warehouse_id);

-- Warehouse availability calendar
CREATE TABLE warehouse_availability (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  warehouse_id    UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
  date            DATE NOT NULL,
  available_slots INTEGER NOT NULL,
  booked_slots    INTEGER NOT NULL DEFAULT 0,
  occupancy_rate  NUMERIC(5,4) GENERATED ALWAYS AS
                  (CASE WHEN available_slots + booked_slots > 0
                        THEN booked_slots::NUMERIC / (available_slots + booked_slots)
                        ELSE 0 END) STORED,
  dynamic_price   NUMERIC(10,2),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_availability_warehouse_date ON warehouse_availability(warehouse_id, date);
CREATE INDEX idx_availability_date ON warehouse_availability(date);
