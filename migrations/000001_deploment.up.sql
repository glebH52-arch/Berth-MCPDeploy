
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    repository_url TEXT NOT NULL,
    running_deployment_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT projects_name_not_empty
        CHECK (length(trim(name)) > 0),

    CONSTRAINT projects_repository_url_not_empty
        CHECK (length(trim(repository_url)) > 0)
);

CREATE TABLE deployments (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    repository_url TEXT NOT NULL,
    branch TEXT NOT NULL,
    commit_sha TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    running_at TIMESTAMPTZ,
    url TEXT,
    failed_at TIMESTAMPTZ,
    failure_stage TEXT,
    error_class TEXT,
    error_message TEXT,
    log_excerpt TEXT,
    CONSTRAINT fk_deployments_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    CONSTRAINT deployments_status_check
        CHECK (status IN (
            'queued',
            'cloning',
            'clone_failed',
            'building',
            'build_failed',
            'starting',
            'start_failed',
            'healthchecking',
            'healthcheck_failed',
            'running'
        )),
    CONSTRAINT deployments_required_fields_check
        CHECK (
            length(trim(repository_url)) > 0
            AND length(trim(branch)) > 0
            AND length(trim(commit_sha)) > 0
        ),
    CONSTRAINT deployments_running_check
        CHECK (
            status <> 'running'
            OR (
                url IS NOT NULL
                AND length(trim(url)) > 0
                AND running_at IS NOT NULL
            )
        ),
    CONSTRAINT deployments_failure_check
        CHECK (
            status NOT IN (
                'clone_failed',
                'build_failed',
                'start_failed',
                'healthcheck_failed'
            )
            OR (
                failed_at IS NOT NULL
                AND failure_stage IS NOT NULL
                AND length(trim(failure_stage)) > 0
                AND error_class IS NOT NULL
                AND length(trim(error_class)) > 0
                AND error_message IS NOT NULL
                AND length(trim(error_message)) > 0
            )
        ),
    CONSTRAINT deployments_project_pair_unique
        UNIQUE (id, project_id)
);

ALTER TABLE projects
ADD CONSTRAINT fk_projects_running_deployment
FOREIGN KEY (running_deployment_id, id)
REFERENCES deployments(id, project_id)
DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX idx_deployments_project_created
    ON deployments(project_id, created_at DESC);

CREATE INDEX idx_deployments_status
    ON deployments(status);

CREATE INDEX idx_deployments_project_status
    ON deployments(project_id, status);

CREATE UNIQUE INDEX idx_deployments_one_active_per_project
    ON deployments(project_id)
    WHERE status IN (
    'queued',
    'cloning',
    'building',
    'starting',
    'healthchecking'
);