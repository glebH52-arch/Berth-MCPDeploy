package messaging

import (
	"context"

	deploymentv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
)

type CloningStartedHandler interface {
	ApplyCloningStarted(
		ctx context.Context,
		event deploymentv1.DeploymentCloningStartedEvent,
	) error
}
