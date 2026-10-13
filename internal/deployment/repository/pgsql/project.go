package pgsql

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ProjectRepository = (*PostgresProjectRepository)(nil)

type PostgresProjectRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProjectRepository(pool *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{
		pool: pool,
	}

}

func (r *PostgresProjectRepository) CreateProject(ctx context.Context, project *domain.Project) error {
	if project == nil {
		return repository.ErrProjectIsNil
	}
	_, err := r.pool.Exec(ctx, stmtCreateProject, project.ID, project.Name, project.RepositoryURL, project.CreatedAt)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	return nil
}
func (r *PostgresProjectRepository) GetProject(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	project := domain.Project{}
	row := r.pool.QueryRow(ctx, stmtGetProjectByID, id)
	err := row.Scan(&project.ID, &project.Name, &project.RepositoryURL, &project.RunningDeploymentID, &project.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProjectNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &project, nil
}
