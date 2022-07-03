package githelpers

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

func CloneRepoWithTimeout(repoURL string, clonePath string, timeoutSeconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	err := exec.CommandContext(ctx, "git", "clone", "--depth=1", repoURL, clonePath).Run()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ctx.Err()
	}

	return err
}

func GetRepoCommitID(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func checkoutRepoCommitID(repoPath, commitID string) error {
	cmd := exec.Command("git", "checkout", commitID)
	cmd.Dir = repoPath
	return cmd.Run()
}
