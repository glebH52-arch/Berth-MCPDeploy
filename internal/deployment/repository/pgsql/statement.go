package pgsql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateProject        = "create_project"
	stmtGetProjectByID       = "get_project_by_id"
	stmtUpdateProject        = "update_project"
	stmtDeleteProject        = "delete_project"
	stmtSetRunningDeployment = "set_running_deployment"

	stmtCreateDeployment        = "create_deployment"
	stmtGetDeploymentByID       = "get_deployment_by_id"
	stmtListDeploymentByProject = "list_deployment_by_project"
	stmtUpdateDeploymentStatus  = "update_deployment_status"
	stmtFailDeployment          = "fail_deployment"
	stmtCompleteDeployment      = "complete_deployment"

	stmtCreateOutboxEvent = "create_outbox_event"
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
	stmtUpdateProject: `

	`,
	stmtDeleteProject: `

	`,
	stmtSetRunningDeployment: `

	`,
	stmtCreateDeployment: `
	INSERT INTO deployments (id,project_id,status,repository_url,branch,commit_sha,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
	stmtGetDeploymentByID: `
	SELECT id,project_id,status,repository_url,branch,commit_sha,created_at,
			updated_at,running_at,url,failed_at,failure_stage,error_class,error_message,log_excerpt
	FROM deployments
	WHERE id = $1
	`,

	stmtCreateOutboxEvent: `
	INSERT INTO outbox_events(id,aggregate_type,aggregate_id,event_type,payload,created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
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
