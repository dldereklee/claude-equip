package allowlist

import "testing"

func TestClaudeLocalAllowsKnownFiles(t *testing.T) {
	cases := []string{"CLAUDE.md", ".claude/settings.json", ".claude/agents/reviewer.md", ".mcp.json", "AGENTS.md"}
	for _, c := range cases {
		if !Allowed("claude", "local", c) {
			t.Fatalf("expected allowed: %s", c)
		}
	}
}

func TestTraversalIsDenied(t *testing.T) {
	if Allowed("claude", "local", ".claude/agents/../../bashrc") {
		t.Fatalf("path traversal should not be allowed")
	}
	if Allowed("codex", "global", ".codex/../config.toml") {
		t.Fatalf("normalized escape should not be allowed")
	}
}

func TestCodexGlobalAllowlist(t *testing.T) {
	if !Allowed("codex", "global", ".codex/config.toml") {
		t.Fatalf("expected codex global config to be allowed")
	}
	if Allowed("codex", "global", ".codex/other.toml") {
		t.Fatalf("unexpected codex global file allowed")
	}
}
