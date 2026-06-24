-- Migration 006: IoT telemetry, sensor readings, GPS tracking

CREATE TYPE sensor_type AS ENUM (
  'temperature_humidity',  -- BLE RuuviTag Pro
  'ethylene_co2',          -- GasSense CO2/C2H4
  'rfid_portal',           -- Zebra FX9600
  'gps_tracker',           -- Queclink GV500
  'smart_lock'             -- Assa Abloy ENTR
);

CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');

-- Sensor device registry
CREATE TABLE iot_sensors (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
  sensor_type     sensor_type NOT NULL,
  device_id       VARCHAR(100) NOT NULL UNIQUE,  -- Hardware serial
  model           VARCHAR(100),
  vendor          VARCHAR(100),
  location_desc   VARCHAR(255),     -- "Cold Room A, Rack 3"
  slot_position   VARCHAR(20),      -- For RFID portal: door/bay identifier
  is_active       BOOLEAN DEFAULT TRUE,
  last_seen_at    TIMESTAMPTZ,
  battery_percent NUMERIC(5,2),
  installed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sensors_warehouse ON iot_sensors(warehouse_id);
CREATE INDEX idx_sensors_device ON iot_sensors(device_id);

-- Time-series sensor readings (high volume — partitioned by month)
CREATE TABLE sensor_readings (
  id                      UUID NOT NULL DEFAULT uuid_generate_v4(),
  sensor_id               UUID NOT NULL REFERENCES iot_sensors(id),
  warehouse_id            UUID NOT NULL REFERENCES warehouses(id),
  gateway_id              VARCHAR(100) NOT NULL,

  -- Payload (nullable — sensor type determines which fields are populated)
  temperature_celsius     NUMERIC(6,2),
  relative_humidity       NUMERIC(5,2),
  ethylene_ppm            NUMERIC(8,4),
  co2_ppm                 NUMERIC(8,2),
  battery_percentage      NUMERIC(5,2),

  -- GPS fields
  gps_latitude            NUMERIC(10,7),
  gps_longitude           NUMERIC(10,7),
  gps_speed_kmh           NUMERIC(6,2),

  -- RFID event
  rfid_tag_id             VARCHAR(100),
  rfid_event_type         VARCHAR(20),  -- 'entry' | 'exit'

  recorded_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  received_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (recorded_at);

-- Create partitions for current and next 3 months
CREATE TABLE sensor_readings_2026_06 PARTITION OF sensor_readings
  FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE sensor_readings_2026_07 PARTITION OF sensor_readings
  FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE sensor_readings_2026_08 PARTITION OF sensor_readings
  FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE sensor_readings_2026_09 PARTITION OF sensor_readings
  FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE INDEX idx_readings_sensor ON sensor_readings(sensor_id, recorded_at DESC);
CREATE INDEX idx_readings_warehouse ON sensor_readings(warehouse_id, recorded_at DESC);
CREATE INDEX idx_readings_gateway ON sensor_readings(gateway_id, recorded_at DESC);

-- Spoilage and threshold alerts
CREATE TABLE iot_alerts (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
  sensor_id       UUID REFERENCES iot_sensors(id),
  alert_type      VARCHAR(50) NOT NULL,     -- 'temperature_breach' | 'ethylene_spike' | 'humidity_high'
  severity        alert_severity NOT NULL DEFAULT 'warning',
  current_value   NUMERIC(10,4),
  threshold_value NUMERIC(10,4),
  message         TEXT NOT NULL,
  affected_pallet_ids UUID[],
  is_resolved     BOOLEAN NOT NULL DEFAULT FALSE,
  resolved_at     TIMESTAMPTZ,
  triggered_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_warehouse ON iot_alerts(warehouse_id, triggered_at DESC);
CREATE INDEX idx_alerts_resolved ON iot_alerts(is_resolved) WHERE is_resolved = FALSE;

-- GPS truck tracking
CREATE TABLE gps_tracks (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  booking_id      UUID REFERENCES bookings(id),
  vehicle_id      VARCHAR(100) NOT NULL,
  driver_id       UUID REFERENCES users(id),
  latitude        NUMERIC(10,7) NOT NULL,
  longitude       NUMERIC(10,7) NOT NULL,
  geo_point       GEOMETRY(Point, 4326)
    GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)) STORED,
  speed_kmh       NUMERIC(6,2),
  temperature_celsius NUMERIC(6,2),  -- Reefer temp during transit
  recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (recorded_at);

CREATE TABLE gps_tracks_2026_06 PARTITION OF gps_tracks
  FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE gps_tracks_2026_07 PARTITION OF gps_tracks
  FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE INDEX idx_gps_tracks_vehicle ON gps_tracks(vehicle_id, recorded_at DESC);
CREATE INDEX idx_gps_tracks_booking ON gps_tracks(booking_id, recorded_at DESC);
CREATE INDEX idx_gps_geo ON gps_tracks USING GIST(geo_point);
