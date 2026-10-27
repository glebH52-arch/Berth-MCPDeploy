package repository

import (
	"context"
	"errors"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
)

var (
	ErrDeploymentNotFound       = errors.New("deployment is not found")
	ErrDeploymentIsNil          = errors.New("deployment is nil")
	ErrDeploymentStatusConflict = errors.New("deployment is not queued")
)

type DeploymentRepository interface {
	CreateDeploymentWithOutbox(ctx context.Context, deployment *domain.Deployment, event *events.OutboxEvent) error
	GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error)
	ApplyCloningStarted(ctx context.Context, consumerName string, eventID uuid.UUID, deploymentID uuid.UUID) error
}
