package pgsql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateProject  = "create_project"
	stmtGetProjectByID = "get_project_by_id"
)

var statements = map[string]string{
	stmtCreateProject: `
	INSERT INTO projects(id,name,repository_url,created_at)
	VALUES ($1,$2,$3,$4)
	`,
	stmtGetProjectByID: `
	SELECT id,name,repository_url,running_deployment_id,created_at FROM projects
	WHERE id = $1
	`,
}

func prepareStatements(ctx context.Context, conn *pgx.Conn) error {
	for name, sql := range statements {
		_, err := conn.Prepare(ctx, name, sql)
		if err != nil {
			return fmt.Errorf("prepare statement %s: %w", name, err)
		}
	}
	return nil
}
