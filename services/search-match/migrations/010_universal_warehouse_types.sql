-- Migration 010: Expand warehouse_type enum and seed multi-vertical warehouses
-- StorEdge is now a universal storage marketplace, not agricultural-only.

-- Add new warehouse type variants for non-agricultural verticals
ALTER TYPE warehouse_type ADD VALUE IF NOT EXISTS 'industrial';
ALTER TYPE warehouse_type ADD VALUE IF NOT EXISTS 'pharmaceutical';
ALTER TYPE warehouse_type ADD VALUE IF NOT EXISTS 'retail_backroom';

-- Rename 'ambient' to a clearer alias (keep 'ambient' for backward compatibility)
-- The front-end will display these with user-friendly labels via warehouseTypeLabel map.

-- ─── Seed: Diverse multi-vertical warehouses ─────────────────────────────────
-- Owner: Vikram Singh (operator, id = 00000000-0000-0000-0000-000000000010)

INSERT INTO warehouses (
  id, owner_id, name, description, type,
  geo_location, address_line1, city, state, pincode,
  total_pallet_capacity, available_pallet_slots,
  base_price_per_pallet_inr, current_price_per_pallet_inr,
  min_temperature_celsius, max_temperature_celsius,
  rating, total_reviews,
  wdra_status, gst_number,
  status
) VALUES

-- 1. Industrial / General Merchandise — Bangalore
(
  'f1000001-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000010',
  'Bangalore Central Logistics Hub',
  'Grade-A industrial warehouse in Whitefield. High-ceiling racking, ESFR sprinklers, 24/7 access. Suitable for e-commerce, FMCG, electronics, and auto parts.',
  'industrial',
  ST_SetSRID(ST_MakePoint(77.7480, 12.9698), 4326),
  'Plot 42, KIADB Industrial Area, Whitefield', 'Bengaluru', 'Karnataka', '560066',
  2000, 1450,
  720, 720,
  NULL, NULL,
  4.6, 89,
  'unregistered', '29AABCT1332L000',
  'active'
),

-- 2. Pharmaceutical / Cold Chain — Mumbai
(
  'f1000002-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000010',
  'Mumbai Pharma Cold Vault',
  'GDP-compliant pharmaceutical cold storage near BKC. Temperature-mapped rooms, validated at 2-8°C and 15-25°C zones. Suitable for vaccines, APIs, and FMCG pharma.',
  'pharmaceutical',
  ST_SetSRID(ST_MakePoint(72.8777, 19.0760), 4326),
  'Unit 7, Bandra Kurla Complex', 'Mumbai', 'Maharashtra', '400051',
  500, 320,
  1800, 1800,
  2, 8,
  4.9, 44,
  'unregistered', '27AABCT1332L000',
  'active'
),

-- 3. Self-Storage / Urban Residential — Delhi NCR
(
  'f1000003-0000-0000-0000-000000000003',
  '00000000-0000-0000-0000-000000000010',
  'Gurugram SecureStore',
  'Clean, climate-controlled self-storage units in Sector 32. Ideal for household goods, furniture, seasonal items, and personal electronics. Drive-up access.',
  'self_storage',
  ST_SetSRID(ST_MakePoint(77.0266, 28.4595), 4326),
  'Plot 18, Sector 32, DLF Phase 2', 'Gurugram', 'Haryana', '122022',
  300, 240,
  480, 480,
  18, 28,
  4.4, 127,
  'unregistered', '06AABCT1332L000',
  'active'
),

-- 4. Retail Backroom / Commercial — Chennai
(
  'f1000004-0000-0000-0000-000000000004',
  '00000000-0000-0000-0000-000000000010',
  'Chennai Retail Stock Depot',
  'Surplus backroom space from a shuttered hypermarket in Anna Nagar. Ideal for retail inventory overflow, last-mile staging, and e-commerce returns processing.',
  'retail_backroom',
  ST_SetSRID(ST_MakePoint(80.2148, 13.0827), 4326),
  '3rd Avenue, Anna Nagar East', 'Chennai', 'Tamil Nadu', '600102',
  800, 650,
  380, 380,
  NULL, NULL,
  4.2, 56,
  'unregistered', '33AABCT1332L000',
  'active'
),

-- 5. Bonded Warehouse — Hyderabad (near airport)
(
  'f1000005-0000-0000-0000-000000000005',
  '00000000-0000-0000-0000-000000000010',
  'Hyderabad Air Cargo Bonded Store',
  'Customs-bonded facility adjacent to RGIA airport. Licensed for duty-deferred storage of imported electronics, auto parts, and machinery before clearance.',
  'bonded',
  ST_SetSRID(ST_MakePoint(78.4294, 17.2403), 4326),
  'Air Cargo Complex, Shamshabad', 'Hyderabad', 'Telangana', '500108',
  1200, 900,
  950, 950,
  NULL, NULL,
  4.7, 33,
  'unregistered', '36AABCT1332L000',
  'active'
)

ON CONFLICT (id) DO NOTHING;
