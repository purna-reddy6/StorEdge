package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	brokers []string
	writers map[string]*kafkago.Writer
	logger  *zap.Logger
}

func NewProducer(brokers string, logger *zap.Logger) *Producer {
	return &Producer{
		brokers: strings.Split(brokers, ","),
		writers: make(map[string]*kafkago.Writer),
		logger:  logger,
	}
}

func (p *Producer) writer(topic string) *kafkago.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		WriteTimeout: 5 * time.Second,
	}
	p.writers[topic] = w
	return w
}

func (p *Producer) PublishTelemetry(topic string, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("marshal telemetry payload", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.writer(topic).WriteMessages(ctx, kafkago.Message{
		Value: body,
		Time:  time.Now(),
	})
	if err != nil {
		p.logger.Warn("kafka publish failed", zap.String("topic", topic), zap.Error(err))
	}
}

func (p *Producer) Close() {
	for _, w := range p.writers {
		w.Close()
	}
}
