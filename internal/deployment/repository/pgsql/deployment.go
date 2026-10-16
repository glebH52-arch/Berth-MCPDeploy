package pgsql

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.DeploymentRepository = (*PostgresDeploymentRepository)(nil)

type PostgresDeploymentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDeploymentRepository(pool *pgxpool.Pool) *PostgresDeploymentRepository {
	return &PostgresDeploymentRepository{
		pool: pool,
	}

}

func scanDeployment(row pgx.Row) (*domain.Deployment, error) {
	var d domain.Deployment

	var url *string
	var failureStage *string
	var errorClass *string
	var errorMessage *string
	var logExcerpt *string

	err := row.Scan(
		&d.ID,
		&d.ProjectID,
		&d.Status,
		&d.RepositoryURL,
		&d.Branch,
		&d.CommitSHA,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.RunningAt,
		&url,
		&d.FailedAt,
		&failureStage,
		&errorClass,
		&errorMessage,
		&logExcerpt,
	)
	if err != nil {
		return nil, err
	}

	if url != nil {
		d.URL = *url
	}
	if failureStage != nil {
		d.FailureStage = *failureStage
	}
	if errorClass != nil {
		d.ErrorClass = *errorClass
	}
	if errorMessage != nil {
		d.ErrorMessage = *errorMessage
	}
	if logExcerpt != nil {
		d.LogExcerpt = *logExcerpt
	}

	return &d, nil
}

func (r *PostgresDeploymentRepository) CreateDeploymentWithOutbox(ctx context.Context, d *domain.Deployment, event *events.OutboxEvent) error {
	if d == nil {
		return repository.ErrDeploymentIsNil
	}
	if event == nil {
		return events.ErrEventIsNil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin deployment with outbox: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, stmtCreateDeployment,
		d.ID,
		d.ProjectID,
		d.Status,
		d.RepositoryURL,
		d.Branch,
		d.CommitSHA,
		d.CreatedAt,
		d.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create deployment: %w", err)
	}
	_, err = tx.Exec(ctx, stmtCreateOutboxEvent,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit deployment with outbox: %w", err)
	}

	return nil
}

func (r *PostgresDeploymentRepository) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	d, err := scanDeployment(
		r.pool.QueryRow(ctx, stmtGetDeploymentByID, id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrDeploymentNotFound
		}
		return nil, fmt.Errorf("get deployment: %w", err)
	}
	return d, nil
}
