package util

import (
	"os"
	"path/filepath"
)

func AppDir() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(cfg, "switcheroo")
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

func RegistryPath() (string, error) {
	base, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "registry.json"), nil
}

func CacheDir() (string, error) {
	base, err := AppDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(base, "cache")
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

func BackupDir() (string, error) {
	base, err := AppDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(base, "backups")
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}
