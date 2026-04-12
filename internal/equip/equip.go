package equip

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FileTarget struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Source struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	CacheDir string `json:"cache_dir"`
}

type Profile struct {
	ID, Name, Tool, Description string
	SourceID                    string       `json:"source_id"`
	ManifestRel                 string       `json:"manifest_rel"`
	Files                       []FileTarget `json:"files"`
}

type Registry struct {
	Sources  []Source  `json:"sources"`
	Profiles []Profile `json:"profiles"`
}

func (r *Registry) upsertSource(s Source) {
	for i := range r.Sources {
		if r.Sources[i].ID == s.ID {
			r.Sources[i] = s
			return
		}
	}
	r.Sources = append(r.Sources, s)
}

func (r *Registry) upsertProfile(p Profile) {
	for i := range r.Profiles {
		if r.Profiles[i].ID == p.ID {
			r.Profiles[i] = p
			return
		}
	}
	r.Profiles = append(r.Profiles, p)
}

func (r Registry) findProfile(id string) (Profile, Source, bool) {
	var p Profile
	for _, item := range r.Profiles {
		if item.ID == id {
			p = item
			break
		}
	}
	if p.ID == "" {
		return Profile{}, Source{}, false
	}
	for _, s := range r.Sources {
		if s.ID == p.SourceID {
			return p, s, true
		}
	}
	return Profile{}, Source{}, false
}

// --- Persistence ---

func dataDir(sub string) (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(cfg, "agent-equip", sub)
	return p, os.MkdirAll(p, 0o755)
}

func LoadRegistry() (Registry, error) {
	dir, err := dataDir("")
	if err != nil {
		return Registry{}, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "registry.json"))
	if os.IsNotExist(err) {
		return Registry{}, nil
	} else if err != nil {
		return Registry{}, err
	}
	var r Registry
	return r, json.Unmarshal(b, &r)
}

func SaveRegistry(r Registry) error {
	dir, err := dataDir("")
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "registry.json"), b, 0o644)
}

// --- CLI ---

func Run(args []string) error {
	if len(args) == 0 {
		return runInteractive()
	}
	switch args[0] {
	case "import":
		if len(args) < 2 {
			return errors.New("usage: agent-equip import <git-url>")
		}
		return importRepo(args[1])
	case "list":
		return listProfiles()
	case "use":
		fs := flag.NewFlagSet("use", flag.ContinueOnError)
		scope := fs.String("scope", "", "local or global")
		force := fs.Bool("force", false, "overwrite conflicts")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if rest := fs.Args(); len(rest) < 1 {
			return errors.New("usage: agent-equip use [--scope local|global] [--force] <profile-id>")
		} else {
			return useProfile(rest[0], *scope, *force)
		}
	case "version", "--version":
		fmt.Println("agent-equip v0.1.0")
		return nil
	case "help", "-h", "--help":
		fmt.Println(`agent-equip - switch Claude Code/Codex setups

Commands:
  import <git-url>                           Import profiles from a git repository
  list                                       List imported profiles
  use [--scope local|global] [--force] <id>  Apply a profile
  version                                    Print version

Run without arguments for interactive mode.`)
		return nil
	default:
		return fmt.Errorf("unknown command: %s (try 'agent-equip help')", args[0])
	}
}

func runInteractive() error {
	for {
		fmt.Println("\nagent-equip interactive\n  1) import  2) list  3) use  4) quit")
		choice, err := prompt("Select [1-4]: ")
		if err != nil {
			return err
		}
		switch strings.TrimSpace(choice) {
		case "1":
			if url, err := prompt("Git URL: "); err != nil {
				return err
			} else if url == "" {
				fmt.Println("URL cannot be empty")
			} else if err := importRepo(url); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		case "2":
			if err := listProfiles(); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		case "3":
			if id, err := prompt("Profile ID: "); err != nil {
				return err
			} else if id == "" {
				fmt.Println("Profile ID cannot be empty")
			} else if err := useProfile(id, "", false); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		case "4", "q", "quit", "exit":
			return nil
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// --- Commands ---

func importRepo(url string) error {
	h := sha1.Sum([]byte(url))
	id := hex.EncodeToString(h[:8])
	cacheDir, err := dataDir("cache")
	if err != nil {
		return err
	}
	dir := filepath.Join(cacheDir, id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if out, err := exec.Command("git", "clone", "--depth", "1", url, dir).CombinedOutput(); err != nil {
			return fmt.Errorf("git clone: %v: %s", err, out)
		}
	} else if out, err := exec.Command("git", "-C", dir, "pull", "--ff-only").CombinedOutput(); err != nil {
		return fmt.Errorf("git pull: %v: %s", err, out)
	}

	var idx struct{ Profiles []struct{ ID, Name, Tool, Description, Path string } }
	if err := readJSON(filepath.Join(dir, "profiles", "index.json"), &idx); err != nil {
		return fmt.Errorf("index.json: %w", err)
	}
	reg, err := LoadRegistry()
	if err != nil {
		return err
	}
	reg.upsertSource(Source{ID: id, URL: url, CacheDir: dir})
	for _, p := range idx.Profiles {
		var m struct{ Files []FileTarget }
		if err := readJSON(filepath.Join(dir, p.Path, "manifest.json"), &m); err != nil {
			return fmt.Errorf("manifest %s: %w", p.ID, err)
		}
		reg.upsertProfile(Profile{
			ID: p.ID, Name: p.Name, Tool: p.Tool, Description: p.Description,
			SourceID: id, ManifestRel: filepath.Join(p.Path, "manifest.json"), Files: m.Files,
		})
	}
	if err := SaveRegistry(reg); err != nil {
		return err
	}
	fmt.Printf("Imported %d profiles from %s\n", len(idx.Profiles), url)
	return nil
}

func listProfiles() error {
	reg, err := LoadRegistry()
	if err != nil {
		return err
	}
	if len(reg.Profiles) == 0 {
		fmt.Println("No profiles imported yet.")
		return nil
	}
	for _, p := range reg.Profiles {
		desc := ""
		if p.Description != "" {
			desc = " - " + p.Description
		}
		fmt.Printf("  %s [%s] %s%s\n", p.ID, p.Tool, p.Name, desc)
	}
	return nil
}

func useProfile(profileID, scope string, force bool) error {
	reg, err := LoadRegistry()
	if err != nil {
		return err
	}
	p, source, ok := reg.findProfile(profileID)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileID)
	}
	if scope == "" {
		fmt.Println("Apply config to:\n  1) local\n  2) global")
		v, err := prompt("Select [1-2]: ")
		if err != nil {
			return err
		}
		scope = "local"
		if v == "2" || strings.EqualFold(v, "global") {
			scope = "global"
		}
	}
	root, err := targetRoot(scope)
	if err != nil {
		return err
	}
	manifestDir := filepath.Dir(filepath.Join(source.CacheDir, p.ManifestRel))

	type item struct{ src, dst, target string }
	var apply []item
	for _, f := range p.Files {
		if !allowed(p.Tool, scope, f.Target) {
			fmt.Printf("  skipped (not allowed): %s\n", f.Target)
			continue
		}
		src, err := safeJoin(manifestDir, f.Source)
		if err != nil {
			fmt.Printf("  skipped (unsafe source): %s\n", f.Target)
			continue
		}
		dst, err := safeJoin(root, f.Target)
		if err != nil {
			fmt.Printf("  skipped (unsafe target): %s\n", f.Target)
			continue
		}
		apply = append(apply, item{src, dst, f.Target})
	}

	if !force {
		for _, it := range apply {
			if filesConflict(it.src, it.dst) {
				return fmt.Errorf("conflict at %s (use --force to overwrite)", it.target)
			}
		}
	}

	bdir, _ := dataDir("backups")
	bdir = filepath.Join(bdir, profileID+"-"+time.Now().UTC().Format("20060102-150405"))
	os.MkdirAll(bdir, 0o755)
	for _, it := range apply {
		if b, err := os.ReadFile(it.dst); err == nil {
			rel, _ := filepath.Rel(root, it.dst)
			bp := filepath.Join(bdir, rel)
			os.MkdirAll(filepath.Dir(bp), 0o755)
			os.WriteFile(bp, b, 0o644)
		}
		if err := copyFile(it.src, it.dst); err != nil {
			return err
		}
	}
	fmt.Printf("Applied %d files (%s → %s). Backup: %s\n", len(apply), p.Name, scope, bdir)
	return nil
}

// --- Helpers ---

func prompt(msg string) (string, error) {
	fmt.Print(msg)
	v, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(v), err
}

func readJSON(p string, v any) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func targetRoot(scope string) (string, error) {
	if strings.EqualFold(scope, "global") {
		return os.UserHomeDir()
	}
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		return filepath.Clean(strings.TrimSpace(string(out))), nil
	}
	return os.Getwd()
}

func allowed(tool, scope, target string) bool {
	t := strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(target, "\\", "/")), "/"), "./"))
	switch strings.ToLower(tool) {
	case "claude":
		if strings.EqualFold(scope, "local") {
			return t == "claude.md" || t == "agents.md" || t == ".mcp.json" ||
				t == ".claude/settings.json" || t == ".claude/settings.local.json" ||
				strings.HasPrefix(t, ".claude/agents/")
		}
		return t == ".claude/settings.json" || strings.HasPrefix(t, ".claude/agents/")
	case "codex":
		if strings.EqualFold(scope, "local") {
			return t == "agents.md"
		}
		return t == ".codex/config.toml" || t == ".codex/agents.md"
	}
	return false
}

func filesConflict(src, dst string) bool {
	sb, e1 := os.ReadFile(src)
	db, e2 := os.ReadFile(dst)
	return e2 == nil && (e1 != nil || string(sb) != string(db))
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(dst), 0o755)
	return os.WriteFile(dst, b, 0o644)
}

func safeJoin(root, rel string) (string, error) {
	base, _ := filepath.Abs(root)
	joined, _ := filepath.Abs(filepath.Join(base, filepath.FromSlash(rel)))
	r, _ := filepath.Rel(base, joined)
	if r == ".." || strings.HasPrefix(r, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return joined, nil
}
