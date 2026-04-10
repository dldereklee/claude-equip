package store

import (
	"encoding/json"
	"os"

	"github.com/example/switcheroo/internal/profile"
	"github.com/example/switcheroo/internal/util"
)

func LoadRegistry() (profile.Registry, error) {
	path, err := util.RegistryPath()
	if err != nil {
		return profile.Registry{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return profile.Registry{}, nil
		}
		return profile.Registry{}, err
	}
	var r profile.Registry
	if err := json.Unmarshal(b, &r); err != nil {
		return profile.Registry{}, err
	}
	return r, nil
}

func SaveRegistry(r profile.Registry) error {
	path, err := util.RegistryPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
