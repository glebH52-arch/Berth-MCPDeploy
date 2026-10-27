package kafkamessaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
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
		Topic:        "deployment.events.v1",
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
		Async:        false,
		BatchSize:    1,
		MaxAttempts:  3,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

func (p *KafkaEventPublisher) Publish(ctx context.Context, event deploymentv1.DeploymentCloningStartedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal cloning started event: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(event.DeploymentID.String()),
		Value: payload,
	}
	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("publish cloning started event: %w", err)
	}
	return nil
}
