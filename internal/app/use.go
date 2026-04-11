package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/switcheroo/internal/allowlist"
	"github.com/example/switcheroo/internal/profile"
	"github.com/example/switcheroo/internal/store"
	"github.com/example/switcheroo/internal/ui"
	"github.com/example/switcheroo/internal/util"
)

type applyItem struct {
	Src    string
	Dst    string
	Target string
}

func UseProfile(profileID, scope string, force bool) error {
	reg, err := store.LoadRegistry()
	if err != nil {
		return err
	}
	p, source, ok := findProfile(reg, profileID)
	if !ok {
		return fmt.Errorf("unknown profile: %s", profileID)
	}
	if scope == "" {
		scope, err = ui.ChooseScope()
		if err != nil {
			return err
		}
	}

	root, err := targetRoot(scope)
	if err != nil {
		return err
	}
	manifestDir := filepath.Dir(filepath.Join(source.CacheDir, p.ManifestRel))

	var applyList []applyItem
	ignored := make([]string, 0)
	for _, f := range p.Files {
		if !allowlist.Allowed(p.Tool, scope, f.Target) {
			ignored = append(ignored, f.Target)
			continue
		}
		src, err := safeJoinUnderRoot(manifestDir, f.Source)
		if err != nil {
			ignored = append(ignored, f.Target+" (unsafe source path)")
			continue
		}
		dst, err := safeJoinUnderRoot(root, f.Target)
		if err != nil {
			ignored = append(ignored, f.Target+" (unsafe target path)")
			continue
		}
		applyList = append(applyList, applyItem{Src: src, Dst: dst, Target: f.Target})
	}

	fmt.Printf("Applying %s (%s) to %s\n", p.Name, p.Tool, scope)
	if len(ignored) > 0 {
		fmt.Println("Ignored files (not in allowlist):")
		for _, ig := range ignored {
			fmt.Printf("  - %s\n", ig)
		}
	}

	if !force {
		for _, it := range applyList {
			if existsAndDifferent(it.Src, it.Dst) {
				return fmt.Errorf("conflict at %s (use --force to overwrite)", it.Target)
			}
		}
	}

	backupRoot, err := util.BackupDir()
	if err != nil {
		return err
	}
	tag := time.Now().UTC().Format("20060102-150405")
	backupDir := filepath.Join(backupRoot, profileID+"-"+tag)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}

	for _, it := range applyList {
		if err := backupExisting(root, backupDir, it.Dst); err != nil {
			return err
		}
		if err := copyFile(it.Src, it.Dst); err != nil {
			return err
		}
	}
	fmt.Printf("Applied %d files. Backup: %s\n", len(applyList), backupDir)
	return nil
}

func findProfile(reg profile.Registry, id string) (profile.ProfileRegistry, profile.SourceRef, bool) {
	var p profile.ProfileRegistry
	for _, item := range reg.Profiles {
		if item.ID == id {
			p = item
			break
		}
	}
	if p.ID == "" {
		return profile.ProfileRegistry{}, profile.SourceRef{}, false
	}
	for _, s := range reg.Sources {
		if s.ID == p.SourceID {
			return p, s, true
		}
	}
	return profile.ProfileRegistry{}, profile.SourceRef{}, false
}

func targetRoot(scope string) (string, error) {
	if strings.EqualFold(scope, "global") {
		return os.UserHomeDir()
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err == nil {
		return filepath.Clean(strings.TrimSpace(string(out))), nil
	}
	return os.Getwd()
}

func safeJoinUnderRoot(root, rel string) (string, error) {
	baseAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	joined := filepath.Join(baseAbs, filepath.FromSlash(rel))
	joinedAbs, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(baseAbs, joinedAbs)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes target root")
	}
	return joinedAbs, nil
}

func existsAndDifferent(src, dst string) bool {
	sb, err1 := os.ReadFile(src)
	db, err2 := os.ReadFile(dst)
	if err2 != nil {
		return false
	}
	if err1 != nil {
		return true
	}
	return string(sb) != string(db)
}

func backupExisting(root, backupRoot, dst string) error {
	b, err := os.ReadFile(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	rel, err := filepath.Rel(root, dst)
	if err != nil {
		return err
	}
	bp := filepath.Join(backupRoot, rel)
	if err := os.MkdirAll(filepath.Dir(bp), 0o755); err != nil {
		return err
	}
	return os.WriteFile(bp, b, 0o644)
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
