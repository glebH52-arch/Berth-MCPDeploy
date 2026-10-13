
ALTER TABLE projects
DROP CONSTRAINT IF EXISTS fk_projects_running_deployment;
DROP INDEX IF EXISTS idx_deployments_one_active_per_project;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS projects;
