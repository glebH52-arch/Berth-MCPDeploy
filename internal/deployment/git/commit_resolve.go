package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type CommitResolver interface {
	ResolveCommit(ctx context.Context, repositoryURL string, branch string) (string, error)
}

type Resolver struct{}

func (r *Resolver) ResolveCommit(ctx context.Context, repositoryURL string, branch string) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"git",
		"ls-remote",
		"--exit-code",
		"--",
		repositoryURL,
		"refs/heads/"+branch,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git ls-remote: %w", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) != 2 || fields[1] != ("refs/heads/"+branch) {
		return "", fmt.Errorf("error resolve commit")
	}
	return fields[0], nil
}
