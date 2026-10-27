package pgsql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.OutboxRepository = (*PostgresOutboxRepository)(nil)

type PostgresOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOutboxRepository(pool *pgxpool.Pool) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{
		pool: pool,
	}
}

var (
	ErrLeaseLost = errors.New("lease is lost")
)

func (r *PostgresOutboxRepository) ClaimBatch(ctx context.Context, limit int, publisherID string, leaseSeconds int) ([]*events.OutboxEvent, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	publisherID = strings.TrimSpace(publisherID)
	if publisherID == "" {
		return nil, fmt.Errorf("publisher ID is required")
	}

	if leaseSeconds <= 0 {
		return nil, fmt.Errorf("lease seconds must be positive")
	}

	rows, err := r.pool.Query(ctx, stmtClaimOutboxBatch, limit, publisherID, leaseSeconds)
	if err != nil {
		return nil, fmt.Errorf("claim outbox batch: %w", err)
	}
	defer rows.Close()

	batch := make([]*events.OutboxEvent, 0, limit)

	for rows.Next() {
		event := &events.OutboxEvent{}

		err := rows.Scan(
			&event.ID,
			&event.AggregateType,
			&event.AggregateID,
			&event.EventType,
			&event.Payload,
			&event.CreatedAt,
			&event.LockedBy,
			&event.LockedUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}

		batch = append(batch, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox events: %w", err)
	}

	return batch, nil
}

func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, eventID uuid.UUID, publisherID string) error {

	publisherID = strings.TrimSpace(publisherID)
	if publisherID == "" {
		return fmt.Errorf("publisher ID is required")
	}

	tag, err := r.pool.Exec(ctx, stmtMarkPublished, eventID, publisherID)
	if err != nil {
		return fmt.Errorf("mark event published: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}

	return nil

}

func (r *PostgresOutboxRepository) MarkFailed(ctx context.Context, eventID uuid.UUID, publisherID string, errorMessage string) error {

	publisherID = strings.TrimSpace(publisherID)
	if publisherID == "" {
		return fmt.Errorf("publisher ID is required")
	}
	tag, err := r.pool.Exec(ctx, stmtMarkFailed, eventID, publisherID, errorMessage)
	if err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}

	return nil
}
