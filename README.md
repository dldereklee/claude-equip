# setupx

`setupx` is a lightweight Go CLI for switching Claude Code and Codex configuration profiles.

## Current features (MVP)

- Import profile packs from a git URL.
- List imported profiles.
- Apply a profile to local (repo root by default) or global scope.
- Copy + backup apply strategy.
- Conflict detection (abort unless `--force`).
- Allowlist enforcement with explicit ignored-file reporting.
- Basic interactive mode.

## Profile repository format

A profile repo must include:

```text
profiles/
  index.json
  <profile-dir>/
    manifest.json
    files/...
```

### `profiles/index.json`

```json
{
  "profiles": [
    {
      "id": "claude-default",
      "name": "Claude Default",
      "tool": "claude",
      "description": "Example profile",
      "path": "profiles/claude-default"
    }
  ]
}
```

### `manifest.json`

```json
{
  "name": "Claude Default",
  "tool": "claude",
  "files": [
    { "source": "files/CLAUDE.md", "target": "CLAUDE.md" },
    { "source": "files/settings.json", "target": ".claude/settings.json" }
  ]
}
```

## Commands

```bash
setupx import <git-url>
setupx list
setupx use [--scope local|global] [--force] <profile-id>
```

Run `setupx` with no arguments to open interactive mode.
