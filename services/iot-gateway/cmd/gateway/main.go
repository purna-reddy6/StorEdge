package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/iot-gateway/internal/handler"
	"github.com/storedge/storedge/services/iot-gateway/internal/kafka"
	"github.com/storedge/storedge/services/iot-gateway/internal/mqtt"
	"github.com/storedge/storedge/services/iot-gateway/internal/parser"
	"github.com/storedge/storedge/services/iot-gateway/internal/storage"
)

func main() {
	_ = godotenv.Load()

	logger, _ := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	dbURL := getenv("DATABASE_URL", "postgres://storedge:storedge_dev@localhost:5432/storedge?sslmode=disable")
	kafkaBrokers := getenv("KAFKA_BROKERS", "localhost:9092")
	mqttBroker := getenv("MQTT_BROKER", "tcp://localhost:1883")
	port := getenv("PORT", "8083")

	db, err := storage.NewPostgres(dbURL)
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}
	defer db.Close()

	kafkaProducer := kafka.NewProducer(kafkaBrokers, logger)
	defer kafkaProducer.Close()

	sensorParser := parser.NewSensorParser(logger)
	telemetryStore := storage.NewTelemetryStore(db)

	// MQTT subscriber — listens on storedge/sensors/# topic
	mqttSub := mqtt.NewSubscriber(mqttBroker, sensorParser, telemetryStore, kafkaProducer, logger)
	if err := mqttSub.Connect(); err != nil {
		logger.Warn("MQTT connect failed — running in HTTP-only mode", zap.Error(err))
	} else {
		mqttSub.Subscribe("storedge/sensors/#")
		logger.Info("MQTT subscriber active", zap.String("broker", mqttBroker))
		defer mqttSub.Disconnect()
	}

	// HTTP API for simulator and health
	router := handler.NewRouter(telemetryStore, kafkaProducer, sensorParser, logger)
	srv := &http.Server{
		Addr:        ":" + port,
		Handler:     router,
		ReadTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("IoT gateway started", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Start sensor simulator in dev mode
	if os.Getenv("APP_ENV") == "development" {
		go runSimulator(kafkaProducer, logger)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("IoT gateway stopped")
}

// runSimulator emits fake sensor readings every 30s in dev mode.
func runSimulator(producer *kafka.Producer, logger *zap.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		reading := map[string]interface{}{
			"gateway_id":          "GW-UP-AGRA-088",
			"facility_id":         "10000000-0000-0000-0000-000000000001",
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"sensor_id":           "GW-UP-AGRA-088-S01",
			"temperature_celsius": 4.2 + (float64(time.Now().Second()%5) * 0.1),
			"relative_humidity":   88.5,
			"ethylene_ppm":        0.12,
			"battery_percentage":  94.0,
		}
		producer.PublishTelemetry("storedge.iot.telemetry", reading)
		logger.Debug("simulator: emitted sensor reading")
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
