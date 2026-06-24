-- Migration 005: e-NWR (Electronic Negotiable Warehouse Receipt) and FinTech

CREATE TYPE enwrs_status AS ENUM (
  'draft',
  'issued',           -- Officially issued by WDRA-registered warehouse
  'pledged',          -- Used as loan collateral with a bank
  'partially_released',
  'fully_released',
  'cancelled',
  'expired'
);

CREATE TYPE loan_status AS ENUM (
  'applied',
  'under_review',
  'sanctioned',
  'disbursed',
  'partially_repaid',
  'fully_repaid',
  'defaulted',
  'npa'  -- Non-Performing Asset
);

-- Electronic Negotiable Warehouse Receipts
CREATE TABLE enwrs_receipts (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  receipt_number          VARCHAR(50) NOT NULL UNIQUE,   -- NERL/CCRL assigned number
  pallet_item_id          UUID NOT NULL REFERENCES pallet_items(id),
  warehouse_id            UUID NOT NULL REFERENCES warehouses(id),
  depositor_id            UUID NOT NULL REFERENCES users(id),  -- Farmer/trader

  -- Commodity details (at time of deposit)
  commodity_type          commodity_type NOT NULL,
  commodity_variety       VARCHAR(100),    -- e.g., "Kufri Jyoti" potato variety
  quantity_kg             NUMERIC(12,2) NOT NULL,
  quality_grade           VARCHAR(20),     -- 'A', 'B', 'FAQ' etc.

  -- Valuation
  market_value_inr        NUMERIC(14,2) NOT NULL,   -- At time of issuance
  current_market_value_inr NUMERIC(14,2),            -- Updated by pricing engine
  ltv_ratio               NUMERIC(5,4) DEFAULT 0.70, -- 70% LTV (RBI guideline)
  max_loan_amount_inr     NUMERIC(14,2)
    GENERATED ALWAYS AS (market_value_inr * ltv_ratio) STORED,

  -- Repository
  repository              VARCHAR(10) NOT NULL DEFAULT 'NERL',  -- 'NERL' | 'CCRL'
  repository_receipt_id   VARCHAR(100),   -- Repository's own reference
  wdra_registration_number VARCHAR(100),

  -- Status and dates
  status                  enwrs_status NOT NULL DEFAULT 'draft',
  issued_at               TIMESTAMPTZ,
  expiry_date             DATE NOT NULL,

  -- Split tracking (partial releases)
  original_quantity_kg    NUMERIC(12,2) NOT NULL,
  released_quantity_kg    NUMERIC(12,2) NOT NULL DEFAULT 0,
  remaining_quantity_kg   NUMERIC(12,2)
    GENERATED ALWAYS AS (original_quantity_kg - released_quantity_kg) STORED,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_enwrs_depositor ON enwrs_receipts(depositor_id);
CREATE INDEX idx_enwrs_warehouse ON enwrs_receipts(warehouse_id);
CREATE INDEX idx_enwrs_status ON enwrs_receipts(status);
CREATE INDEX idx_enwrs_receipt_number ON enwrs_receipts(receipt_number);

-- Loan applications against e-NWRs
CREATE TABLE loan_applications (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  application_number      VARCHAR(30) NOT NULL UNIQUE,
  applicant_id            UUID NOT NULL REFERENCES users(id),
  enwrs_receipt_id        UUID NOT NULL REFERENCES enwrs_receipts(id),

  -- Loan details
  requested_amount_inr    NUMERIC(14,2) NOT NULL,
  sanctioned_amount_inr   NUMERIC(14,2),
  interest_rate_percent   NUMERIC(5,2),     -- Annual %
  tenure_days             INTEGER,

  -- Platform fee
  origination_fee_inr     NUMERIC(10,2),   -- 1.5% of sanctioned amount
  origination_fee_paid    BOOLEAN DEFAULT FALSE,

  -- Bank details
  partner_bank_name       VARCHAR(100),    -- e.g. "Canara Bank", "Union Bank"
  bank_loan_reference     VARCHAR(100),
  bank_account_number     VARCHAR(30),
  bank_ifsc               VARCHAR(15),

  status                  loan_status NOT NULL DEFAULT 'applied',
  applied_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  sanctioned_at           TIMESTAMPTZ,
  disbursed_at            TIMESTAMPTZ,

  -- PSL classification (RBI Priority Sector Lending)
  is_psl_eligible         BOOLEAN DEFAULT TRUE,
  psl_limit_inr           NUMERIC(14,2) DEFAULT 7500000,  -- ₹75 lakh

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loans_applicant ON loan_applications(applicant_id);
CREATE INDEX idx_loans_enwrs ON loan_applications(enwrs_receipt_id);
CREATE INDEX idx_loans_status ON loan_applications(status);

-- Loan repayments
CREATE TABLE loan_repayments (
  id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  loan_id                 UUID NOT NULL REFERENCES loan_applications(id),
  amount_inr              NUMERIC(14,2) NOT NULL,
  payment_reference       VARCHAR(100),
  paid_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
