package kafkamessaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/messaging"
	"github.com/segmentio/kafka-go"
)

type KafkaEventConsumer struct {
	reader  *kafka.Reader
	handler messaging.CloningStartedHandler
}

func NewKafkaEventConsumer(reader *kafka.Reader, handler messaging.CloningStartedHandler) *KafkaEventConsumer {
	return &KafkaEventConsumer{
		reader:  reader,
		handler: handler,
	}
}

func NewReader(brokers []string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "deployment.events.v1",
		GroupID:        "deployment-service-events-v1",
		CommitInterval: 0,
	})
}

func (c *KafkaEventConsumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch deployment command: %w", err)
		}

		slog.Info("deployment command received",
			"service", "deployment-service",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
		)
		var event deploymentv1.DeploymentCloningStartedEvent

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("decode deployment command: %w", err)
		}
		if string(msg.Key) != event.DeploymentID.String() {
			return fmt.Errorf("event key does not match deployment_id")
		}
		err = c.handler.ApplyCloningStarted(ctx, event)
		if err != nil {
			return fmt.Errorf("handle cloning started event: %w", err)
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("commit deployment command: %w", err)
		}
	}
}
