package domain

import (
	"errors"
	"strings"
	"time"

	"uuid"
)

type Project struct {
	ID                  uuid.UUID
	Name                string
	RepositoryURL       string
	RunningDeploymentID *uuid.UUID
	CreatedAt           time.Time
}

var (
	ErrEmptyName = errors.New("name can not be empty")
)

func NewProject(name, repositoryURL string) (*Project, error) {
	if strings.TrimSpace(repositoryURL) == "" {
		return nil, ErrEmptyRepositoryURL
	}

	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyName
	}

	now := time.Now().UTC()

	project := &Project{
		ID:            uuid.New(),
		Name:          strings.TrimSpace(name),
		RepositoryURL: strings.TrimSpace(repositoryURL),
		CreatedAt:     now,
	}
	return project, nil
}
