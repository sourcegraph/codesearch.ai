package sg

import (
	"strings"
	"testing"
)

func TestGetHighlightedCodeLineRange(t *testing.T) {
	code, err := GetHighlightedCodeLineRange("github.com/sourcegraph/sourcegraph", "main", "cmd/frontend/graphqlbackend/repository.go", 0, 10)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(code, "data-line=\"1\"") {
		t.Fatalf("code should contain data-line=\"1\"")
	}

	if !strings.Contains(code, "data-line=\"10\"") {
		t.Fatalf("code should contain data-line=\"10\"")
	}
}
