package outbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
)

type EventPublisher interface {
	Publish(ctx context.Context, event *events.OutboxEvent) error
}

type Publisher struct {
	outboxRepo     repository.OutboxRepository
	eventPublisher EventPublisher
	publisherID    string
}

func NewPublisher(outboxRepo repository.OutboxRepository, eventPublisher EventPublisher, publisherID string) *Publisher {
	return &Publisher{
		outboxRepo:     outboxRepo,
		eventPublisher: eventPublisher,
		publisherID:    publisherID,
	}
}

func (p *Publisher) PublishBatch(ctx context.Context) (bool, error) {
	batch, err := p.outboxRepo.ClaimBatch(ctx, 1, p.publisherID, 30)
	if err != nil {
		return false, fmt.Errorf("error claim batch: %w", err)
	}
	if len(batch) < 1 {
		return false, nil
	}
	event := batch[0]

	publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = p.eventPublisher.Publish(publishCtx, event)
	cancel()
	if err != nil {
		markErr := p.outboxRepo.MarkFailed(ctx, event.ID, p.publisherID, err.Error())
		if markErr != nil {
			return true, errors.Join(fmt.Errorf("publish event: %w", err),
				fmt.Errorf("mark event failed: %w", markErr),
			)
		}
		return true, fmt.Errorf("publish event: %w", err)
	}

	if err := p.outboxRepo.MarkPublished(ctx, event.ID, p.publisherID); err != nil {
		return true, fmt.Errorf("mark event published: %w", err)
	}

	return true, nil

}

func (p *Publisher) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		claimed, err := p.PublishBatch(ctx)

		if ctx.Err() != nil {
			return nil
		}

		if claimed && err == nil {
			continue
		}

		if err != nil {
			slog.Error("outbox publication failed",
				"service", "outbox-publisher",
				"publisher_id", p.publisherID,
				"error", err,
			)
		}

		timer := time.NewTimer(time.Second)

		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
