package repository

import (
	"context"
	"errors"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
)

var (
	ErrProjectNotFound = errors.New("project is not found")
	ErrProjectIsNil    = errors.New("project is nil")
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *domain.Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*domain.Project, error)
}
