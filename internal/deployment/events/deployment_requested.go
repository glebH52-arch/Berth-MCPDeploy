package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
)

var (
	ErrEventIsNil        = errors.New("event is nil")
	ErrInvalidDeployment = errors.New("invalid deployment")
)

const (
	EventTypeDeploymentRequested = "deployment.requested"
	AggregateTypeDeployment      = "deployment"
	DeploymentEventSchemaVersion = 1
)

type DeploymentRequestedPayload struct {
	RepositoryURL string `json:"repository_url"`
	Branch        string `json:"branch"`
	CommitSHA     string `json:"commit_sha"`
}

type DeploymentRequestedEvent struct {
	EventID       uuid.UUID                  `json:"event_id"`
	EventType     string                     `json:"event_type"`
	SchemaVersion int                        `json:"schema_version"`
	OccurredAt    time.Time                  `json:"occurred_at"`
	CorrelationID uuid.UUID                  `json:"correlation_id"`
	DeploymentID  uuid.UUID                  `json:"deployment_id"`
	Payload       DeploymentRequestedPayload `json:"payload"`
}

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
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

	event := DeploymentRequestedEvent{
		EventID:       eventID,
		EventType:     EventTypeDeploymentRequested,
		SchemaVersion: DeploymentEventSchemaVersion,
		OccurredAt:    now,
		CorrelationID: d.ID,
		DeploymentID:  d.ID,

		Payload: DeploymentRequestedPayload{
			RepositoryURL: d.RepositoryURL,
			Branch:        d.Branch,
			CommitSHA:     d.CommitSHA,
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
