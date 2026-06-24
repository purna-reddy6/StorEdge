package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/iot-gateway/internal/kafka"
	"github.com/storedge/storedge/services/iot-gateway/internal/parser"
	"github.com/storedge/storedge/services/iot-gateway/internal/storage"
)

func NewRouter(
	telemetryStore *storage.TelemetryStore,
	kafkaProducer *kafka.Producer,
	sensorParser *parser.SensorParser,
	logger *zap.Logger,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "iot-gateway"})
	})

	api := r.Group("/api/v1")

	// Ingest sensor readings via HTTP (used by edge simulator and direct HTTP sensors)
	api.POST("/telemetry/ingest", func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(400, gin.H{"error": "cannot read body"})
			return
		}

		payload, err := sensorParser.Parse(body)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		_ = telemetryStore.Save(payload)
		kafkaProducer.PublishTelemetry("storedge.iot.telemetry", payload)

		alerts := sensorParser.DetectAlerts(payload)
		if len(alerts) > 0 {
			alertJSON, _ := json.Marshal(alerts)
			kafkaProducer.PublishTelemetry("storedge.iot.alerts", json.RawMessage(alertJSON))
		}

		c.JSON(200, gin.H{"ingested": true, "alerts": len(alerts)})
	})

	// Get latest readings for a facility
	api.GET("/telemetry/facility/:facilityId", func(c *gin.Context) {
		facilityID := c.Param("facilityId")
		limit := 20
		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				limit = v
			}
		}
		readings, err := telemetryStore.GetLatestReadings(facilityID, limit)
		if err != nil {
			logger.Error("get readings failed", zap.Error(err))
			c.JSON(500, gin.H{"error": "failed to fetch readings"})
			return
		}
		c.JSON(200, gin.H{"readings": readings, "facility_id": facilityID})
	})

	return r
}
