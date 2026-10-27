package worker

import (
	"context"
	"fmt"
	"strings"
	"time"
	"uuid"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
)

type CommandHandler interface {
	Handle(ctx context.Context, event deploymentv1.DeploymentRequestedEvent) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event deploymentv1.DeploymentCloningStartedEvent) error
}

type Worker struct {
	publisher EventPublisher
}

func NewWorker(publisher EventPublisher) *Worker {
	return &Worker{
		publisher: publisher,
	}
}

func (w *Worker) Handle(ctx context.Context, event deploymentv1.DeploymentRequestedEvent) error {
	if event.EventType != "deployment.requested.v1" {
		return fmt.Errorf("unsupported event type: %q", event.EventType)
	}

	if event.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema version: %d", event.SchemaVersion)
	}
	if event.EventID == uuid.Nil() {
		return fmt.Errorf("event_id is required")
	}

	if event.DeploymentID == uuid.Nil() {
		return fmt.Errorf("deployment_id is required")
	}

	if event.CorrelationID == uuid.Nil() {
		return fmt.Errorf("correlation_id is required")
	}

	if event.Payload.ProjectID == uuid.Nil() {
		return fmt.Errorf("project_id is required")
	}

	if event.OccurredAt.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}

	if strings.TrimSpace(event.Payload.RepositoryURL) == "" {
		return fmt.Errorf("repository_url is required")
	}

	if strings.TrimSpace(event.Payload.Branch) == "" {
		return fmt.Errorf("branch is required")
	}

	if strings.TrimSpace(event.Payload.CommitSHA) == "" {
		return fmt.Errorf("commit_sha is required")
	}

	newEvent := deploymentv1.DeploymentCloningStartedEvent{
		EventID:       uuid.New(),
		EventType:     deploymentv1.EventTypeDeploymentCloningStarted,
		SchemaVersion: 1,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: event.CorrelationID,
		DeploymentID:  event.DeploymentID,
		Payload:       deploymentv1.DeploymentCloningStartedPayload{},
	}

	return w.publisher.Publish(ctx, newEvent)
}
