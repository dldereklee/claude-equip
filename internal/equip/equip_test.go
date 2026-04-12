package equip

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllowed(t *testing.T) {
	for _, tc := range []struct {
		tool, scope, target string
		want                bool
	}{
		{"claude", "local", "CLAUDE.md", true},
		{"claude", "local", "agents.md", true},
		{"claude", "local", ".claude/settings.json", true},
		{"claude", "local", ".claude/agents/a.md", true},
		{"claude", "local", ".bashrc", false},
		{"claude", "local", ".claude/agents/../../bashrc", false},
		{"claude", "global", ".claude/settings.json", true},
		{"claude", "global", "CLAUDE.md", false},
		{"CLAUDE", "LOCAL", "CLAUDE.md", true},
		{"codex", "local", "agents.md", true},
		{"codex", "global", ".codex/config.toml", true},
		{"codex", "global", ".codex/../config.toml", false},
		{"codex", "global", ".codex/other.toml", false},
		{"unknown", "local", "CLAUDE.md", false},
	} {
		if got := allowed(tc.tool, tc.scope, tc.target); got != tc.want {
			t.Errorf("allowed(%q,%q,%q)=%v want %v", tc.tool, tc.scope, tc.target, got, tc.want)
		}
	}
}

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()
	if _, err := safeJoin(root, ".claude/settings.json"); err != nil {
		t.Fatalf("valid path rejected: %v", err)
	}
	if _, err := safeJoin(root, "../outside"); err == nil {
		t.Fatal("traversal should fail")
	}
	if _, err := safeJoin(root, "../../etc/passwd"); err == nil {
		t.Fatal("deep traversal should fail")
	}
}

func TestRegistryOps(t *testing.T) {
	var reg Registry
	reg.upsertSource(Source{ID: "s1", URL: "a"})
	reg.upsertSource(Source{ID: "s2", URL: "b"})
	reg.upsertSource(Source{ID: "s1", URL: "updated"})
	if len(reg.Sources) != 2 || reg.Sources[0].URL != "updated" {
		t.Fatalf("sources: %+v", reg.Sources)
	}
	reg.upsertProfile(Profile{ID: "p1", Name: "A", SourceID: "s1"})
	reg.upsertProfile(Profile{ID: "p1", Name: "Updated", SourceID: "s1"})
	if len(reg.Profiles) != 1 || reg.Profiles[0].Name != "Updated" {
		t.Fatalf("profiles: %+v", reg.Profiles)
	}
	if _, _, ok := reg.findProfile("p1"); !ok {
		t.Fatal("should find p1")
	}
	if _, _, ok := reg.findProfile("nope"); ok {
		t.Fatal("should not find nope")
	}
}

func TestFilesConflict(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	os.WriteFile(a, []byte("x"), 0o644)
	if filesConflict(a, b) {
		t.Error("no conflict when dest missing")
	}
	os.WriteFile(b, []byte("x"), 0o644)
	if filesConflict(a, b) {
		t.Error("no conflict when same")
	}
	os.WriteFile(b, []byte("y"), 0o644)
	if !filesConflict(a, b) {
		t.Error("conflict expected")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "s"), filepath.Join(dir, "d", "f")
	os.WriteFile(src, []byte("data"), 0o644)
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(dst); string(got) != "data" {
		t.Errorf("got %q", got)
	}
}

func TestRegistryPersistence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	reg := Registry{
		Sources:  []Source{{ID: "s1", URL: "https://example.com"}},
		Profiles: []Profile{{ID: "p1", Name: "Test", SourceID: "s1"}},
	}
	if err := SaveRegistry(reg); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Sources) != 1 || len(loaded.Profiles) != 1 || loaded.Profiles[0].Name != "Test" {
		t.Errorf("loaded: %+v", loaded)
	}
}
