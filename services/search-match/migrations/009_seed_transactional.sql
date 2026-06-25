-- 009_seed_transactional.sql
-- Seed bookings, pallet items, IoT alerts, sensor readings, e-NWR receipts.
-- User IDs (from 008_seed_data.sql):
--   Vikram Singh  operator: 00000000-0000-0000-0000-000000000010
--   Rajesh Kumar  farmer  : 00000000-0000-0000-0000-000000000020
--   Suresh Yadav  farmer  : 00000000-0000-0000-0000-000000000021
--   Ananya Sharma trader  : 00000000-0000-0000-0000-000000000030
-- Warehouses (all owned by Vikram Singh ...010):
--   Singh Cold Storage : 10000000-0000-0000-0000-000000000001
--   Agra Frost         : 10000000-0000-0000-0000-000000000002

-- ── Bookings ─────────────────────────────────────────────────────────────────
INSERT INTO bookings (
  id, booking_number, tenant_id, warehouse_id,
  pallet_count, commodity_type,
  price_per_pallet_inr, total_amount_inr, commission_amount_inr, payout_amount_inr,
  start_date, end_date, status
) VALUES
(
  'b0000001-0000-0000-0000-000000000001',
  'BK-2025-001001',
  '00000000-0000-0000-0000-000000000030',  -- Ananya Sharma (trader)
  '10000000-0000-0000-0000-000000000001',  -- Singh Cold Storage
  40, 'potato',
  520.00, 62400.00, 6240.00, 56160.00,
  '2025-11-01', '2026-02-28', 'active'
),
(
  'b0000002-0000-0000-0000-000000000002',
  'BK-2025-001002',
  '00000000-0000-0000-0000-000000000020',  -- Rajesh Kumar (farmer)
  '10000000-0000-0000-0000-000000000001',  -- Singh Cold Storage
  25, 'onion',
  600.00, 45000.00, 4500.00, 40500.00,
  '2025-12-01', '2026-03-31', 'active'
),
(
  'b0000003-0000-0000-0000-000000000003',
  'BK-2025-001003',
  '00000000-0000-0000-0000-000000000030',  -- Ananya Sharma (trader)
  '10000000-0000-0000-0000-000000000002',  -- Agra Frost
  15, 'garlic',
  400.00, 18000.00, 1800.00, 16200.00,
  '2025-10-15', '2026-01-15', 'completed'
)
ON CONFLICT (id) DO NOTHING;

-- ── Pallet Items ─────────────────────────────────────────────────────────────
-- pallet_items has: id, booking_id, warehouse_id, tenant_id, commodity_type,
--                   weight_kg, slot_position, inward_date, expected_outward_date
INSERT INTO pallet_items (
  id, booking_id, warehouse_id, tenant_id,
  commodity_type, weight_kg, slot_position,
  inward_date, expected_outward_date
) VALUES
('c0000001-0000-0000-0000-000000000001',
 'b0000001-0000-0000-0000-000000000001',
 '10000000-0000-0000-0000-000000000001',
 '00000000-0000-0000-0000-000000000030',
 'potato', 1000.00, 'A-01-001', '2025-11-01', '2026-02-28'),

('c0000002-0000-0000-0000-000000000002',
 'b0000001-0000-0000-0000-000000000001',
 '10000000-0000-0000-0000-000000000001',
 '00000000-0000-0000-0000-000000000030',
 'potato', 1000.00, 'A-01-002', '2025-11-01', '2026-02-28'),

('c0000003-0000-0000-0000-000000000003',
 'b0000001-0000-0000-0000-000000000001',
 '10000000-0000-0000-0000-000000000001',
 '00000000-0000-0000-0000-000000000030',
 'potato', 950.00, 'A-01-003', '2025-11-02', '2026-02-28'),

('c0000004-0000-0000-0000-000000000004',
 'b0000002-0000-0000-0000-000000000002',
 '10000000-0000-0000-0000-000000000001',
 '00000000-0000-0000-0000-000000000020',
 'onion', 800.00, 'B-02-001', '2025-12-01', '2026-03-31'),

('c0000005-0000-0000-0000-000000000005',
 'b0000002-0000-0000-0000-000000000002',
 '10000000-0000-0000-0000-000000000001',
 '00000000-0000-0000-0000-000000000020',
 'onion', 820.00, 'B-02-002', '2025-12-01', '2026-03-31')
ON CONFLICT (id) DO NOTHING;

-- ── IoT Sensor Readings (last 24h, in current month's partition) ──────────────
INSERT INTO sensor_readings (
  sensor_id, warehouse_id, gateway_id,
  temperature_celsius, relative_humidity, recorded_at
)
SELECT
  s.id,
  s.warehouse_id,
  'GW-UP-AGRA-088',
  (3.5 + random() * 1.5)::numeric(6,2),
  (88 + random() * 6)::numeric(5,2),
  NOW() - (gs * INTERVAL '1 hour')
FROM iot_sensors s
CROSS JOIN generate_series(0, 23) AS gs
WHERE s.warehouse_id = '10000000-0000-0000-0000-000000000001'
  AND s.sensor_type = 'temperature_humidity';

-- ── IoT Alerts ───────────────────────────────────────────────────────────────
INSERT INTO iot_alerts (
  id, warehouse_id, sensor_id, alert_type, severity, message,
  is_resolved, triggered_at
) VALUES
(
  'a0000001-0000-0000-0000-000000000001',
  '10000000-0000-0000-0000-000000000001',
  (SELECT id FROM iot_sensors
   WHERE warehouse_id='10000000-0000-0000-0000-000000000001'
     AND sensor_type='temperature_humidity' LIMIT 1),
  'temperature_breach', 'warning',
  'Temperature in Cold Room A rose above 8°C. Current: 8.7°C. Check compressor unit.',
  FALSE,
  NOW() - INTERVAL '3 hours'
),
(
  'a0000002-0000-0000-0000-000000000002',
  '10000000-0000-0000-0000-000000000001',
  (SELECT id FROM iot_sensors
   WHERE warehouse_id='10000000-0000-0000-0000-000000000001'
     AND sensor_type='temperature_humidity' LIMIT 1),
  'humidity_breach', 'critical',
  'Humidity spike: 96% in Rack 5. Condensation risk — potato spoilage risk HIGH.',
  FALSE,
  NOW() - INTERVAL '45 minutes'
),
(
  'a0000003-0000-0000-0000-000000000003',
  '10000000-0000-0000-0000-000000000002',
  (SELECT id FROM iot_sensors
   WHERE warehouse_id='10000000-0000-0000-0000-000000000001'
     AND sensor_type='temperature_humidity' LIMIT 1),
  'power_outage', 'critical',
  'Power interrupted. Generator backup activated. Monitoring temperature.',
  TRUE,
  NOW() - INTERVAL '2 days'
)
ON CONFLICT (id) DO NOTHING;

UPDATE iot_alerts
SET resolved_at = NOW() - INTERVAL '23 hours', is_resolved = TRUE
WHERE id = 'a0000003-0000-0000-0000-000000000003';

-- ── e-NWR Receipts ────────────────────────────────────────────────────────────
-- max_loan_amount_inr is generated (market_value_inr × ltv_ratio=0.70) — do not insert
INSERT INTO enwrs_receipts (
  id, receipt_number, pallet_item_id, warehouse_id, depositor_id,
  commodity_type, quantity_kg, market_value_inr,
  original_quantity_kg, status, expiry_date
) VALUES
(
  'e0000001-0000-0000-0000-000000000001',
  'ENWR/UP/2025/001234',
  'c0000001-0000-0000-0000-000000000001',
  '10000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000020',  -- Rajesh Kumar
  'potato', 2950.00, 767000.00, 2950.00, 'issued', '2026-02-28'
),
(
  'e0000002-0000-0000-0000-000000000002',
  'ENWR/UP/2025/001890',
  'c0000004-0000-0000-0000-000000000004',
  '10000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000030',  -- Ananya Sharma (trader)
  'onion', 1620.00, 648000.00, 1620.00, 'issued', '2026-03-31'
)
ON CONFLICT (id) DO NOTHING;

-- ── Stock Release Request ─────────────────────────────────────────────────────
-- stock_release_requests: id, pallet_item_id, tenant_id, warehouse_id, quantity_to_release_kg, status
INSERT INTO stock_release_requests (
  id, pallet_item_id, tenant_id, warehouse_id, quantity_to_release_kg, status
) VALUES
(
  'd0000001-0000-0000-0000-000000000001',
  'c0000005-0000-0000-0000-000000000005',
  '00000000-0000-0000-0000-000000000020',  -- Rajesh Kumar
  '10000000-0000-0000-0000-000000000001',  -- Singh Cold Storage
  820.00, 'pending_otp'
)
ON CONFLICT (id) DO NOTHING;
