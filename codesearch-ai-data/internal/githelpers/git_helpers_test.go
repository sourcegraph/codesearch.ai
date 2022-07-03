package githelpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRepoCloning(t *testing.T) {
	repoPath := fmt.Sprintf("/tmp/test-%d", time.Now().Unix())
	err := CloneRepoWithTimeout("https://github.com/kelseyhightower/nocode", repoPath, 10)
	if err != nil {
		t.Fatal(err)
	}

	wantCommitID := "6c073b08f7987018cbb2cb9a5747c84913b3608e"
	err = checkoutRepoCommitID(repoPath, wantCommitID)
	if err != nil {
		t.Fatal(err)
	}

	gotCommitID, err := GetRepoCommitID(repoPath)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(repoPath)
	if err != nil {
		t.Fatal(err)
	}

	if gotCommitID != wantCommitID {
		t.Fatalf("Want commit id: %s, got %s", wantCommitID, gotCommitID)
	}
}

func TestRepoTimeout(t *testing.T) {
	repoPath := fmt.Sprintf("/tmp/test-timeout-%d", time.Now().Unix())
	err := CloneRepoWithTimeout("https://github.com/sourcegraph/sourcegraph", repoPath, 1)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected timeout, got error %+v", err)
	}

	err = os.RemoveAll(repoPath)
	if err != nil {
		t.Fatal(err)
	}
}
