package repository

import (
	"context"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
)

type OutboxRepository interface {
	ClaimBatch(ctx context.Context, limit int, publisherID string, leaseSeconds int) ([]*events.OutboxEvent, error)
	MarkPublished(ctx context.Context, eventID uuid.UUID, publisherID string) error
	MarkFailed(ctx context.Context, eventID uuid.UUID, publisherID string, errorMessage string) error
}
