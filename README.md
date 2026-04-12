# agent-equip

`agent-equip` is a lightweight Go CLI for switching Claude Code and Codex configuration profiles.

## Current features (MVP)

- Import profile packs from a git URL.
- List imported profiles.
- Apply a profile to local (repo root by default) or global scope.
- Copy + backup apply strategy.
- Conflict detection (abort unless `--force`).
- Allowlist enforcement with explicit ignored-file reporting.
- Path-safety checks that block traversal/escape paths during apply.
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
agent-equip import <git-url>
agent-equip list
agent-equip use [--scope local|global] [--force] <profile-id>
```

Run `agent-equip` with no arguments to open interactive mode.

## Install (3 options)

### Option 1 (Best): One-line installer script (recommended)

This is the easiest path for most users on macOS/Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/<OWNER>/<REPO>/main/scripts/install.sh | bash
```

Notes:
- By default it installs to `/usr/local/bin`.
- You can override defaults:

```bash
AGENT_EQUIP_REPO=<OWNER>/<REPO> AGENT_EQUIP_INSTALL_DIR="$HOME/.local/bin" bash scripts/install.sh
```

### Option 2: Download a prebuilt binary from GitHub Releases

1. Open Releases and download the artifact for your OS/arch:
   - `agent-equip_<version>_linux_amd64.tar.gz`
   - `agent-equip_<version>_darwin_arm64.tar.gz`
   - `agent-equip_<version>_windows_amd64.zip`
2. Extract and move `agent-equip` (or `agent-equip.exe`) into your PATH.

### Option 3: Build/install from source (Go toolchain required)

```bash
go install github.com/example/agent-equip@latest
```

### Why Option 1 is best

Option 1 has the best balance of:
- zero Go/toolchain setup,
- very fast install,
- consistent install experience across machines.

## CI/CD release flow

- CI (`.github/workflows/ci.yml`): runs `go test ./...` on PRs and pushes to `main`.
- Release (`.github/workflows/release.yml`): on tags like `v1.2.3`, GoReleaser builds binaries for:
  - macOS (`amd64`, `arm64`)
  - Linux (`amd64`, `arm64`)
  - Windows (`amd64`, `arm64`)
- Artifacts + checksums are published to GitHub Releases automatically.

To publish a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```
