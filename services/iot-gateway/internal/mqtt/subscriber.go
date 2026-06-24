package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/iot-gateway/internal/kafka"
	"github.com/storedge/storedge/services/iot-gateway/internal/parser"
	"github.com/storedge/storedge/services/iot-gateway/internal/storage"
)

type Subscriber struct {
	client         pahomqtt.Client
	sensorParser   *parser.SensorParser
	telemetryStore *storage.TelemetryStore
	kafkaProducer  *kafka.Producer
	logger         *zap.Logger
}

func NewSubscriber(
	brokerURL string,
	sensorParser *parser.SensorParser,
	telemetryStore *storage.TelemetryStore,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) *Subscriber {
	s := &Subscriber{
		sensorParser:   sensorParser,
		telemetryStore: telemetryStore,
		kafkaProducer:  kafkaProducer,
		logger:         logger,
	}

	opts := pahomqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(fmt.Sprintf("storedge-gateway-%d", time.Now().UnixMilli())).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectionLostHandler(func(c pahomqtt.Client, err error) {
			logger.Warn("MQTT connection lost", zap.Error(err))
		}).
		SetOnConnectHandler(func(c pahomqtt.Client) {
			logger.Info("MQTT connected")
		})

	s.client = pahomqtt.NewClient(opts)
	return s
}

func (s *Subscriber) Connect() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (s *Subscriber) Subscribe(topic string) {
	s.client.Subscribe(topic, 1, s.handleMessage)
	s.logger.Info("subscribed to MQTT topic", zap.String("topic", topic))
}

func (s *Subscriber) Disconnect() {
	s.client.Disconnect(500)
}

func (s *Subscriber) handleMessage(_ pahomqtt.Client, msg pahomqtt.Message) {
	payload, err := s.sensorParser.Parse(msg.Payload())
	if err != nil {
		s.logger.Error("parse sensor payload", zap.Error(err), zap.String("topic", msg.Topic()))
		return
	}

	// Persist to PostgreSQL (with local SQLite buffer on connectivity loss in edge runtime)
	if err := s.telemetryStore.Save(payload); err != nil {
		s.logger.Error("save telemetry", zap.Error(err))
	}

	// Forward to Kafka for downstream consumers (AI engine, WMS alerts)
	s.kafkaProducer.PublishTelemetry("storedge.iot.telemetry", payload)

	// Check thresholds and emit alerts
	alerts := s.sensorParser.DetectAlerts(payload)
	for _, alert := range alerts {
		s.logger.Warn("sensor alert detected",
			zap.String("type", alert.AlertType),
			zap.String("severity", alert.Severity),
			zap.String("facility", alert.FacilityID),
			zap.String("message", alert.Message),
		)
		alertPayload, _ := json.Marshal(alert)
		s.kafkaProducer.PublishTelemetry("storedge.iot.alerts", json.RawMessage(alertPayload))
	}
}
