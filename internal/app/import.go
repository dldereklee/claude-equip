package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/switcheroo/internal/gitops"
	"github.com/example/switcheroo/internal/profile"
	"github.com/example/switcheroo/internal/store"
)

func ImportRepo(url string) error {
	sourceID, cacheDir, err := gitops.Sync(url)
	if err != nil {
		return err
	}

	idxPath := filepath.Join(cacheDir, "profiles", "index.json")
	b, err := os.ReadFile(idxPath)
	if err != nil {
		return fmt.Errorf("read index.json: %w", err)
	}
	var idx profile.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return fmt.Errorf("parse index.json: %w", err)
	}

	reg, err := store.LoadRegistry()
	if err != nil {
		return err
	}

	upsertSource(&reg, profile.SourceRef{ID: sourceID, URL: url, CacheDir: cacheDir})
	for _, p := range idx.Profiles {
		manifestPath := filepath.Join(cacheDir, p.Path, "manifest.json")
		mb, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("read manifest for %s: %w", p.ID, err)
		}
		var m profile.Manifest
		if err := json.Unmarshal(mb, &m); err != nil {
			return fmt.Errorf("parse manifest for %s: %w", p.ID, err)
		}
		upsertProfile(&reg, profile.ProfileRegistry{
			ID:          p.ID,
			Name:        p.Name,
			Tool:        p.Tool,
			Description: p.Description,
			SourceID:    sourceID,
			ManifestRel: filepath.Join(p.Path, "manifest.json"),
			Files:       m.Files,
		})
	}

	if err := store.SaveRegistry(reg); err != nil {
		return err
	}
	fmt.Printf("Imported %d profiles from %s\n", len(idx.Profiles), url)
	return nil
}

func upsertSource(r *profile.Registry, s profile.SourceRef) {
	for i := range r.Sources {
		if r.Sources[i].ID == s.ID {
			r.Sources[i] = s
			return
		}
	}
	r.Sources = append(r.Sources, s)
}

func upsertProfile(r *profile.Registry, p profile.ProfileRegistry) {
	for i := range r.Profiles {
		if r.Profiles[i].ID == p.ID {
			r.Profiles[i] = p
			return
		}
	}
	r.Profiles = append(r.Profiles, p)
}
