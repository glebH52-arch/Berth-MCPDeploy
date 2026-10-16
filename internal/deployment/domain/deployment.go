package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
)

type Deployment struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	Status        DeploymentStatus
	RepositoryURL string
	Branch        string
	CommitSHA     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	URL           string
	RunningAt     *time.Time
	FailedAt      *time.Time
	FailureStage  string
	ErrorClass    string
	ErrorMessage  string
	LogExcerpt    string
}

var (
	ErrInvalidDeploymentTransition   = errors.New("invalid deployment transition")
	ErrInvalidProjectID              = errors.New("invalid project ID")
	ErrEmptyRepositoryURL            = errors.New("repository URL cannot be empty")
	ErrEmptyBranch                   = errors.New("branch cannot be empty")
	ErrEmptyCommitSHA                = errors.New("commit SHA cannot be empty")
	ErrInvalidFailureStatus          = errors.New("invalid failure status")
	ErrEmptyErrorClass               = errors.New("error class cannot be empty")
	ErrEmptyErrorMessage             = errors.New("error message cannot be empty")
	ErrFailureStatusRequiresFail     = errors.New("failure status transition requires Fail method")
	ErrInvalidApplicationURL         = errors.New("invalid application URL")
	ErrRunningStatusRequiresComplete = errors.New("running status requires Complete method")
)

func (d *Deployment) ChangeStatus(newStatus DeploymentStatus) error {
	switch newStatus {
	case DeploymentStatusFailed:
		return ErrFailureStatusRequiresFail
	case DeploymentStatusRunning:
		return ErrRunningStatusRequiresComplete
	}

	if !CanTransition(d.Status, newStatus) {
		return ErrInvalidDeploymentTransition
	}

	now := time.Now().UTC()

	d.Status = newStatus
	d.UpdatedAt = now

	return nil
}

func (d *Deployment) Fail(errorClass, errorMessage, logExcerpt string) error {
	if !CanTransition(d.Status, DeploymentStatusFailed) {
		return ErrInvalidDeploymentTransition
	}

	errorClass = strings.TrimSpace(errorClass)
	errorMessage = strings.TrimSpace(errorMessage)

	if errorClass == "" {
		return ErrEmptyErrorClass
	}
	if errorMessage == "" {
		return ErrEmptyErrorMessage
	}

	now := time.Now().UTC()
	stage := string(d.Status)

	d.Status = DeploymentStatusFailed
	d.FailureStage = stage
	d.ErrorClass = errorClass
	d.ErrorMessage = errorMessage
	d.LogExcerpt = logExcerpt
	d.FailedAt = &now
	d.UpdatedAt = now

	return nil
}

func (d *Deployment) Complete(applicationURL string) error {
	applicationURL = strings.TrimSpace(applicationURL)

	if applicationURL == "" {
		return ErrInvalidApplicationURL
	}

	parsedURL, err := url.Parse(applicationURL)
	if err != nil ||
		(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
		parsedURL.Hostname() == "" {
		return ErrInvalidApplicationURL
	}

	if !CanTransition(d.Status, DeploymentStatusRunning) {
		return ErrInvalidDeploymentTransition
	}

	now := time.Now().UTC()

	d.Status = DeploymentStatusRunning
	d.URL = applicationURL
	d.RunningAt = &now
	d.UpdatedAt = now

	return nil
}

func NewDeployment(projectID uuid.UUID, repositoryURL, branch, commitSHA string) (*Deployment, error) {
	if projectID == uuid.Nil() {
		return nil, ErrInvalidProjectID
	}

	if strings.TrimSpace(repositoryURL) == "" {
		return nil, ErrEmptyRepositoryURL
	}

	if strings.TrimSpace(branch) == "" {
		return nil, ErrEmptyBranch
	}

	if strings.TrimSpace(commitSHA) == "" {
		return nil, ErrEmptyCommitSHA
	}

	now := time.Now().UTC()

	deployment := &Deployment{
		ID:            uuid.New(),
		ProjectID:     projectID,
		Status:        DeploymentStatusQueued,
		RepositoryURL: strings.TrimSpace(repositoryURL),
		Branch:        strings.TrimSpace(branch),
		CommitSHA:     strings.TrimSpace(commitSHA),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return deployment, nil
}
