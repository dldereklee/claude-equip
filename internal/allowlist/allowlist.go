package allowlist

import "strings"

func Allowed(tool, scope, target string) bool {
	t := clean(target)
	tool = strings.ToLower(tool)
	scope = strings.ToLower(scope)

	switch tool {
	case "claude":
		if scope == "local" {
			return t == "claude.md" || t == "agents.md" || t == ".mcp.json" || t == ".claude/settings.json" || t == ".claude/settings.local.json" || strings.HasPrefix(t, ".claude/agents/")
		}
		return t == ".claude/settings.json" || strings.HasPrefix(t, ".claude/agents/")
	case "codex":
		if scope == "local" {
			return t == "agents.md"
		}
		return t == ".codex/config.toml" || t == ".codex/agents.md"
	default:
		return false
	}
}

func clean(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	s = strings.TrimPrefix(s, "./")
	s = strings.TrimPrefix(s, "/")
	return strings.ToLower(s)
}
