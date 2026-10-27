package kafkamessaging

import (
	"context"
	"fmt"
	"time"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	"github.com/segmentio/kafka-go"
)

type KafkaEventPublisher struct {
	writer *kafka.Writer
}

func NewKafkaEventPublisher(writer *kafka.Writer) *KafkaEventPublisher {
	return &KafkaEventPublisher{
		writer: writer,
	}
}

func NewWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        "deployment.commands.v1",
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
		Async:        false,
		BatchSize:    1,
		MaxAttempts:  3,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

func (p *KafkaEventPublisher) Publish(ctx context.Context, event *events.OutboxEvent) error {
	if event == nil {
		return events.ErrEventIsNil
	}
	msg := kafka.Message{
		Key:   []byte(event.AggregateID.String()),
		Value: event.Payload,
	}
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed publish outboxevent: %w", err)
	}
	return nil
}
