package service

import (
	"context"
	"fmt"
	"uuid"

	contractsv1 "github.com/glebH52-arch/agent-native-paas/internal/contracts/deployment/v1"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	gitresolver "github.com/glebH52-arch/agent-native-paas/internal/deployment/git"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
)

type DeploymentService struct {
	deploymentRepository repository.DeploymentRepository
	projectRepository    repository.ProjectRepository
	commitResolver       gitresolver.CommitResolver
}

func NewDeploymentService(deploymentRepository repository.DeploymentRepository, projectRepository repository.ProjectRepository, commitResolver gitresolver.CommitResolver) *DeploymentService {
	return &DeploymentService{
		deploymentRepository: deploymentRepository,
		projectRepository:    projectRepository,
		commitResolver:       commitResolver,
	}
}

func (s *DeploymentService) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	deployment, err := s.deploymentRepository.GetDeployment(ctx, id)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, projectID uuid.UUID, branch string) (*domain.Deployment, error) {
	project, err := s.projectRepository.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	commitSHA, err := s.commitResolver.ResolveCommit(ctx, project.RepositoryURL, branch)
	if err != nil {
		return nil, err
	}

	deployment, err := domain.NewDeployment(projectID, project.RepositoryURL, branch, commitSHA)
	if err != nil {
		return nil, err
	}

	event, err := events.NewDeploymentRequestedEvent(deployment)
	if err != nil {
		return nil, err
	}

	err = s.deploymentRepository.CreateDeploymentWithOutbox(ctx, deployment, event)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (s *DeploymentService) ApplyCloningStarted(ctx context.Context, event contractsv1.DeploymentCloningStartedEvent) error {
	if event.EventType != "deployment.cloning_started.v1" {
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

	if event.OccurredAt.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}

	return s.deploymentRepository.ApplyCloningStarted(
		ctx,
		"deployment-service-events-v1",
		event.EventID,
		event.DeploymentID)
}
