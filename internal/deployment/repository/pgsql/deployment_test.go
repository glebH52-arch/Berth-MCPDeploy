package pgsql

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/glebH52-arch/agent-native-paas/internal/deployment/domain"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/events"
	"github.com/glebH52-arch/agent-native-paas/internal/deployment/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreateDeploymentWithOutbox_Rollback(t *testing.T) {
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, connString)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	project, err := domain.NewProject(
		"outbox rollback test",
		"https://example.com/test.git",
	)
	if err != nil {
		t.Fatal(err)
	}

	projectRepo := NewPostgresProjectRepository(pool)
	if err := projectRepo.CreateProject(ctx, project); err != nil {
		t.Fatalf("create test project: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := pool.Exec(
			cleanupCtx,
			"DELETE FROM projects WHERE id = $1",
			project.ID,
		); err != nil {
			t.Errorf("cleanup test project: %v", err)
		}
	})
	deployment, err := domain.NewDeployment(
		project.ID,
		project.RepositoryURL,
		"main",
		"7a7f6bfde2ee7119bda014a67e6979934d0ea006",
	)
	if err != nil {
		t.Fatal(err)
	}

	event, err := events.NewDeploymentRequestedEvent(deployment)
	if err != nil {
		t.Fatal(err)
	}
	event.Payload = nil
	repo := NewPostgresDeploymentRepository(pool)
	err = repo.CreateDeploymentWithOutbox(ctx, deployment, event)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected PostgreSQL error, got: %v", err)
	}
	if pgErr.Code != "23502" ||
		pgErr.TableName != "outbox_events" ||
		pgErr.ColumnName != "payload" {
		t.Fatalf("expected outbox payload NOT NULL violation, got: %v", err)
	}
	_, err = repo.GetDeployment(ctx, deployment.ID)
	if !errors.Is(err, repository.ErrDeploymentNotFound) {
		t.Fatalf("expected deployment to be rolled back, got: %v", err)
	}
}
