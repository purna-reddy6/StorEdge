-- Migration 008: Seed data — Agra/Firozabad pilot region
-- Phase 1 go-to-market: Agra potato belt, Uttar Pradesh

-- Platform admin user
INSERT INTO users (id, phone, email, name, role, kyc_status, language_pref) VALUES
  ('00000000-0000-0000-0000-000000000001', '+919999999999', 'admin@storedge.in',
   'StorEdge Admin', 'admin', 'verified', 'en');

-- Sample warehouse operators (cold storage owners in Agra region)
INSERT INTO users (id, phone, email, name, role, kyc_status, language_pref) VALUES
  ('00000000-0000-0000-0000-000000000010', '+919876543210', 'vikram@singhcoldstorage.in',
   'Vikram Singh', 'operator', 'verified', 'hi'),
  ('00000000-0000-0000-0000-000000000011', '+919876543211', 'rahul@agrafrost.in',
   'Rahul Sharma', 'operator', 'verified', 'hi'),
  ('00000000-0000-0000-0000-000000000012', '+919876543212', 'priya@firozabadcoldstorage.in',
   'Priya Gupta', 'operator', 'verified', 'hi');

-- Sample farmers (Agra potato belt)
INSERT INTO users (id, phone, email, name, role, kyc_status, language_pref) VALUES
  ('00000000-0000-0000-0000-000000000020', '+917654321098', NULL,
   'Rajesh Kumar', 'farmer', 'verified', 'hi'),
  ('00000000-0000-0000-0000-000000000021', '+917654321097', NULL,
   'Suresh Yadav', 'farmer', 'verified', 'hi'),
  ('00000000-0000-0000-0000-000000000022', '+917654321096', NULL,
   'Mohan Lal', 'farmer', 'pending', 'hi');

-- Sample e-commerce trader
INSERT INTO users (id, phone, email, name, role, kyc_status, language_pref) VALUES
  ('00000000-0000-0000-0000-000000000030', '+919123456789', 'ananya@d2cbrand.in',
   'Ananya Sharma', 'trader', 'verified', 'en');

-- Seed warehouses in Agra / Firozabad region
INSERT INTO warehouses (
  id, owner_id, name, description, type, status,
  geo_location, address_line1, city, state, pincode, district,
  total_pallet_capacity, available_pallet_slots, floor_area_sqft,
  min_temperature_celsius, max_temperature_celsius,
  base_price_per_pallet_inr, current_dynamic_price_inr,
  wdra_status, wdra_registration_number,
  has_cctv, has_loading_dock, has_backup_power,
  rating, total_reviews, is_active
) VALUES
(
  '10000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000010',
  'Singh Cold Storage — Firozabad',
  'WDRA-registered cold storage specializing in potato and root vegetables. Solar-hybrid power. 10,000 MT capacity.',
  'cold_storage', 'active',
  ST_SetSRID(ST_MakePoint(78.3947, 27.1592), 4326),
  'NH-19, Firozabad Road', 'Firozabad', 'Uttar Pradesh', '283203', 'Firozabad',
  2500, 2100, 45000.00,
  2.0, 8.0,
  850.00, 920.00,
  'registered', 'WDRA/UP/2024/0142',
  TRUE, 3, TRUE,
  4.3, 47, TRUE
),
(
  '10000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000011',
  'Agra Frost — Agra Cold Hub',
  'Multi-commodity cold storage in central Agra. APMC licensed. Handles potato, onion, and garlic.',
  'cold_storage', 'active',
  ST_SetSRID(ST_MakePoint(78.0081, 27.1767), 4326),
  'Shamshabad Road, Bodla', 'Agra', 'Uttar Pradesh', '282005', 'Agra',
  1800, 1600, 32000.00,
  1.5, 10.0,
  780.00, 810.00,
  'registered', 'WDRA/UP/2023/0089',
  TRUE, 2, TRUE,
  4.1, 32, TRUE
),
(
  '10000000-0000-0000-0000-000000000003',
  '00000000-0000-0000-0000-000000000012',
  'Firozabad Cold & Ambient Hub',
  'Mixed-use facility: cold storage for agricultural and ambient warehouse for FMCG. Near Delhi-Agra highway.',
  'ambient', 'active',
  ST_SetSRID(ST_MakePoint(78.4136, 27.1512), 4326),
  'Industrial Area, Phase 2', 'Firozabad', 'Uttar Pradesh', '283203', 'Firozabad',
  3200, 2900, 60000.00,
  NULL, NULL,
  600.00, 625.00,
  'unregistered', NULL,
  TRUE, 4, TRUE,
  3.8, 15, TRUE
);

-- Seed IoT sensors for Singh Cold Storage
INSERT INTO iot_sensors (warehouse_id, sensor_type, device_id, model, vendor, location_desc, is_active) VALUES
  ('10000000-0000-0000-0000-000000000001', 'temperature_humidity', 'GW-UP-AGRA-088-S01',
   'RuuviTag Pro', 'Ruuvi', 'Cold Room A — Rack 1', TRUE),
  ('10000000-0000-0000-0000-000000000001', 'temperature_humidity', 'GW-UP-AGRA-088-S02',
   'RuuviTag Pro', 'Ruuvi', 'Cold Room A — Rack 5', TRUE),
  ('10000000-0000-0000-0000-000000000001', 'ethylene_co2', 'GW-UP-AGRA-088-E01',
   'GasSense CO2/C2H4', 'GasSense', 'Cold Room A — Central', TRUE),
  ('10000000-0000-0000-0000-000000000001', 'rfid_portal', 'GW-UP-AGRA-088-R01',
   'FX9600', 'Zebra Technologies', 'Loading Bay 1 — Entry', TRUE);
