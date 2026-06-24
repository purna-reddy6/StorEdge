package parser

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SensorPayload is the canonical internal representation of any sensor reading.
// Matches the JSON schema shown in the blueprint (Part 10, IoT section):
//
//	{"gateway_id":"GW-UP-AGRA-088","facility_id":"FAC-9931-A","timestamp":"...","payload":{...}}
type SensorPayload struct {
	GatewayID          string    `json:"gateway_id"`
	FacilityID         string    `json:"facility_id"`
	SensorID           string    `json:"sensor_id"`
	TemperatureCelsius *float64  `json:"temperature_celsius"`
	RelativeHumidity   *float64  `json:"relative_humidity"`
	EthylenePPM        *float64  `json:"ethylene_ppm"`
	CO2PPM             *float64  `json:"co2_ppm"`
	BatteryPercentage  *float64  `json:"battery_percentage"`
	GPSLatitude        *float64  `json:"gps_latitude"`
	GPSLongitude       *float64  `json:"gps_longitude"`
	GPSSpeedKMH        *float64  `json:"gps_speed_kmh"`
	RFIDTagID          string    `json:"rfid_tag_id"`
	RFIDEventType      string    `json:"rfid_event_type"`
	RecordedAt         time.Time `json:"recorded_at"`
}

// Alert thresholds (configurable per facility in Phase 2)
const (
	MaxTemperatureColdRoom  = 8.0   // °C — potato cold storage max
	MinTemperatureColdRoom  = 1.0   // °C — frost risk
	MaxEthylenePPM          = 1.0   // ppm — spoilage indicator
	MaxRelativeHumidity     = 95.0  // %
	MinBatteryPercent       = 15.0  // % — low battery alert
)

type Alert struct {
	SensorID      string  `json:"sensor_id"`
	FacilityID    string  `json:"facility_id"`
	AlertType     string  `json:"alert_type"`
	Severity      string  `json:"severity"`
	CurrentValue  float64 `json:"current_value"`
	ThresholdValue float64 `json:"threshold_value"`
	Message       string  `json:"message"`
}

type SensorParser struct {
	logger *zap.Logger
}

func NewSensorParser(logger *zap.Logger) *SensorParser {
	return &SensorParser{logger: logger}
}

// Parse converts a raw MQTT/HTTP payload bytes into a structured SensorPayload.
func (p *SensorParser) Parse(raw []byte) (*SensorPayload, error) {
	var payload SensorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal sensor payload: %w", err)
	}

	if payload.RecordedAt.IsZero() {
		payload.RecordedAt = time.Now().UTC()
	}

	return &payload, nil
}

// DetectAlerts analyzes a sensor reading and returns any threshold violations.
func (p *SensorParser) DetectAlerts(payload *SensorPayload) []Alert {
	var alerts []Alert

	if payload.TemperatureCelsius != nil {
		temp := *payload.TemperatureCelsius
		if temp > MaxTemperatureColdRoom {
			alerts = append(alerts, Alert{
				SensorID: payload.SensorID, FacilityID: payload.FacilityID,
				AlertType: "temperature_breach", Severity: "critical",
				CurrentValue: temp, ThresholdValue: MaxTemperatureColdRoom,
				Message: fmt.Sprintf("Temperature %.1f°C exceeds cold room max %.1f°C — spoilage risk", temp, MaxTemperatureColdRoom),
			})
		}
		if temp < MinTemperatureColdRoom {
			alerts = append(alerts, Alert{
				SensorID: payload.SensorID, FacilityID: payload.FacilityID,
				AlertType: "temperature_low", Severity: "warning",
				CurrentValue: temp, ThresholdValue: MinTemperatureColdRoom,
				Message: fmt.Sprintf("Temperature %.1f°C below frost threshold %.1f°C", temp, MinTemperatureColdRoom),
			})
		}
	}

	if payload.EthylenePPM != nil && *payload.EthylenePPM > MaxEthylenePPM {
		ethylene := *payload.EthylenePPM
		alerts = append(alerts, Alert{
			SensorID: payload.SensorID, FacilityID: payload.FacilityID,
			AlertType: "ethylene_spike", Severity: "critical",
			CurrentValue: ethylene, ThresholdValue: MaxEthylenePPM,
			Message: fmt.Sprintf("Ethylene %.2f ppm exceeds threshold %.1f ppm — early decay detected", ethylene, MaxEthylenePPM),
		})
	}

	if payload.RelativeHumidity != nil && *payload.RelativeHumidity > MaxRelativeHumidity {
		rh := *payload.RelativeHumidity
		alerts = append(alerts, Alert{
			SensorID: payload.SensorID, FacilityID: payload.FacilityID,
			AlertType: "humidity_high", Severity: "warning",
			CurrentValue: rh, ThresholdValue: MaxRelativeHumidity,
			Message: fmt.Sprintf("Humidity %.1f%% exceeds max %.0f%%", rh, MaxRelativeHumidity),
		})
	}

	if payload.BatteryPercentage != nil && *payload.BatteryPercentage < MinBatteryPercent {
		bat := *payload.BatteryPercentage
		alerts = append(alerts, Alert{
			SensorID: payload.SensorID, FacilityID: payload.FacilityID,
			AlertType: "low_battery", Severity: "info",
			CurrentValue: bat, ThresholdValue: MinBatteryPercent,
			Message: fmt.Sprintf("Sensor battery %.0f%% — replace soon", bat),
		})
	}

	return alerts
}
