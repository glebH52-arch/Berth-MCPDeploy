package deploymentv1

import (
	"time"
	"uuid"
)

const EventTypeDeploymentCloningStarted = "deployment.cloning_started.v1"

type DeploymentRequestedPayload struct {
	RepositoryURL string    `json:"repository_url"`
	Branch        string    `json:"branch"`
	CommitSHA     string    `json:"commit_sha"`
	ProjectID     uuid.UUID `json:"project_id"`
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

type DeploymentCloningStartedPayload struct{}

type DeploymentCloningStartedEvent struct {
	EventID       uuid.UUID                       `json:"event_id"`
	EventType     string                          `json:"event_type"`
	SchemaVersion int                             `json:"schema_version"`
	OccurredAt    time.Time                       `json:"occurred_at"`
	CorrelationID uuid.UUID                       `json:"correlation_id"`
	DeploymentID  uuid.UUID                       `json:"deployment_id"`
	Payload       DeploymentCloningStartedPayload `json:"payload"`
}
