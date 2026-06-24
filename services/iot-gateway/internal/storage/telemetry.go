package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/storedge/storedge/services/iot-gateway/internal/parser"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

type TelemetryStore struct {
	db *sql.DB
}

func NewTelemetryStore(db *sql.DB) *TelemetryStore {
	return &TelemetryStore{db: db}
}

// Save persists a sensor reading to the partitioned sensor_readings table.
func (s *TelemetryStore) Save(payload *parser.SensorPayload) error {
	_, err := s.db.Exec(`
		INSERT INTO sensor_readings (
			sensor_id, warehouse_id, gateway_id,
			temperature_celsius, relative_humidity,
			ethylene_ppm, co2_ppm, battery_percentage,
			gps_latitude, gps_longitude, gps_speed_kmh,
			rfid_tag_id, rfid_event_type,
			recorded_at, received_at
		) VALUES (
			(SELECT id FROM iot_sensors WHERE device_id = $1 LIMIT 1),
			(SELECT warehouse_id FROM iot_sensors WHERE device_id = $1 LIMIT 1),
			$2,
			$3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12,
			$13, NOW()
		)`,
		payload.SensorID, payload.GatewayID,
		payload.TemperatureCelsius, payload.RelativeHumidity,
		payload.EthylenePPM, payload.CO2PPM, payload.BatteryPercentage,
		payload.GPSLatitude, payload.GPSLongitude, payload.GPSSpeedKMH,
		emptyIfBlank(payload.RFIDTagID), emptyIfBlank(payload.RFIDEventType),
		payload.RecordedAt,
	)
	return err
}

// SaveAlert persists a threshold alert to iot_alerts.
func (s *TelemetryStore) SaveAlert(facilityID, sensorDeviceID, alertType, severity string,
	currentVal, threshold float64, message string) error {
	_, err := s.db.Exec(`
		INSERT INTO iot_alerts (warehouse_id, sensor_id, alert_type, severity, current_value, threshold_value, message)
		SELECT
			(SELECT warehouse_id FROM iot_sensors WHERE device_id = $1 LIMIT 1),
			(SELECT id FROM iot_sensors WHERE device_id = $1 LIMIT 1),
			$2, $3, $4, $5, $6
		`,
		sensorDeviceID, alertType, severity, currentVal, threshold, message,
	)
	return err
}

// GetLatestReadings returns the most recent readings for a facility.
func (s *TelemetryStore) GetLatestReadings(facilityID string, limit int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT sr.gateway_id, is2.device_id as sensor_id,
			sr.temperature_celsius, sr.relative_humidity,
			sr.ethylene_ppm, sr.battery_percentage,
			sr.recorded_at
		FROM sensor_readings sr
		JOIN iot_sensors is2 ON sr.sensor_id = is2.id
		WHERE sr.warehouse_id = $1::uuid
		ORDER BY sr.recorded_at DESC LIMIT $2`,
		facilityID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []map[string]interface{}
	for rows.Next() {
		var gatewayID, sensorID string
		var temp, humidity, ethylene, battery *float64
		var recordedAt time.Time
		if err := rows.Scan(&gatewayID, &sensorID, &temp, &humidity, &ethylene, &battery, &recordedAt); err != nil {
			continue
		}
		readings = append(readings, map[string]interface{}{
			"gateway_id":          gatewayID,
			"sensor_id":           sensorID,
			"temperature_celsius": temp,
			"relative_humidity":   humidity,
			"ethylene_ppm":        ethylene,
			"battery_percentage":  battery,
			"recorded_at":         recordedAt,
		})
	}
	return readings, rows.Err()
}

func emptyIfBlank(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
