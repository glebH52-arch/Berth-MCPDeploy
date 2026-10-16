package service

import (
	"context"
	"uuid"

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
