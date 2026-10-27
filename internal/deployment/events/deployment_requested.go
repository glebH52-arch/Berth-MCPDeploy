package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
)

var (
	ErrEventIsNil        = errors.New("event is nil")
	ErrInvalidDeployment = errors.New("invalid deployment")
)

const (
	EventTypeDeploymentRequested = "deployment.requested.v1"
	AggregateTypeDeployment      = "deployment"
	DeploymentEventSchemaVersion = 1
)

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time

	LockedBy    *string
	LockedUntil *time.Time
}

func NewDeploymentRequestedEvent(d *domain.Deployment) (*OutboxEvent, error) {
	if d == nil {
		return nil, ErrInvalidDeployment
	}

	if d.ID == uuid.Nil() || d.ProjectID == uuid.Nil() {
		return nil, ErrInvalidDeployment
	}

	if d.Status != domain.DeploymentStatusQueued {
		return nil, ErrInvalidDeployment
	}

	eventID := uuid.New()
	now := time.Now().UTC()

	event := deploymentv1.DeploymentRequestedEvent{
		EventID:       eventID,
		EventType:     EventTypeDeploymentRequested,
		SchemaVersion: DeploymentEventSchemaVersion,
		OccurredAt:    now,
		CorrelationID: d.ID,
		DeploymentID:  d.ID,

		Payload: deploymentv1.DeploymentRequestedPayload{
			RepositoryURL: d.RepositoryURL,
			Branch:        d.Branch,
			CommitSHA:     d.CommitSHA,
			ProjectID:     d.ProjectID,
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal deployment requested event: %w",
			err,
		)
	}

	return &OutboxEvent{
		ID:            eventID,
		AggregateType: AggregateTypeDeployment,
		AggregateID:   d.ID,
		EventType:     EventTypeDeploymentRequested,
		Payload:       payload,
		CreatedAt:     now,
	}, nil
}
