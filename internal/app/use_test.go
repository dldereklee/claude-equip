package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoinUnderRoot(t *testing.T) {
	root := t.TempDir()

	p, err := safeJoinUnderRoot(root, ".claude/settings.json")
	if err != nil {
		t.Fatalf("unexpected error for safe path: %v", err)
	}
	if !strings.HasPrefix(p, root+string(filepath.Separator)) {
		t.Fatalf("expected joined path to stay under root, got %s", p)
	}

	if _, err := safeJoinUnderRoot(root, "../outside"); err == nil {
		t.Fatalf("expected escape path to fail")
	}
}
