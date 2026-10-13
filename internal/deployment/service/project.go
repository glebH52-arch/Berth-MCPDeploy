package service

import (
	"context"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
)

type ProjectService struct {
	projectRepository repository.ProjectRepository
}

func NewProjectService(projectRepository repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepository: projectRepository,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, name, repositoryURL string) (*domain.Project, error) {
	project, err := domain.NewProject(name, repositoryURL)
	if err != nil {
		return nil, err
	}
	err = s.projectRepository.CreateProject(ctx, project)
	if err != nil {
		return nil, err
	}
	return project, nil
}
