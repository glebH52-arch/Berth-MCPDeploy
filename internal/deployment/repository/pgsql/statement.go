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
	stmtMarkDeploymentCloning   = "mark_deployment_cloning"
	stmtFailDeployment          = "fail_deployment"
	stmtCompleteDeployment      = "complete_deployment"
	stmtGetDeploymentStatus     = "get_deployment_status"

	stmtCreateOutboxEvent = "create_outbox_event"
	stmtClaimOutboxBatch  = "claim_outbox_batch"
	stmtMarkPublished     = "mark_published"
	stmtMarkFailed        = "mark_failed"

	stmtRegisterInboxEvent  = "register_inbox_event"
	stmtMarkInboxProcessed  = "mark_inbox_processed"
	stmtGetInboxProcessedAt = "get_processed_at"
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

	stmtMarkDeploymentCloning: `
	UPDATE deployments
	SET status = 'cloning', updated_at = NOW()
	WHERE id = $1 and status = 'queued'
	`,

	stmtGetDeploymentStatus: `
	SELECT status FROM deployments
	WHERE id = $1
	`,

	stmtCreateOutboxEvent: `
	INSERT INTO outbox_events(id,aggregate_type,aggregate_id,event_type,payload,created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`,

	stmtClaimOutboxBatch: `
	WITH candidates AS (
		SELECT id FROM outbox_events
		WHERE published_at IS NULL AND (locked_until IS NULL OR locked_until <=NOW())
		ORDER BY created_at , id
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	)
	UPDATE outbox_events AS out
	SET locked_by = $2, locked_until = NOW() + ($3 * INTERVAL '1 second')
	FROM candidates AS cand
	WHERE out.id = cand.id
	RETURNING out.id,out.aggregate_type,out.aggregate_id,out.event_type,out.payload,
    out.created_at,out.locked_by,out.locked_until;
	`,

	stmtMarkPublished: `
	UPDATE outbox_events
	SET published_at = NOW(), locked_by = NULL, locked_until = NULL
	WHERE id = $1 AND published_at IS NULL AND locked_by = $2 AND locked_until > NOW();
	`,

	stmtMarkFailed: `
	UPDATE outbox_events
	SET attempts = attempts + 1, last_error = $3, locked_by = NULL, locked_until = NULL
	WHERE id = $1 AND published_at IS NULL AND locked_by = $2 AND locked_until > NOW();
	`,

	stmtRegisterInboxEvent: `
	INSERT INTO inbox_events(consumer_name,event_id,event_type)
	VALUES($1,$2,$3)
	ON CONFLICT (consumer_name, event_id) DO NOTHING
	`,

	stmtMarkInboxProcessed: `
	UPDATE inbox_events
	SET processed_at = NOW()
	WHERE consumer_name = $1 AND event_id = $2
	`,

	stmtGetInboxProcessedAt: `
	SELECT processed_at FROM inbox_events
	WHERE consumer_name = $1 AND event_id = $2
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
