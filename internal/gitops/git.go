package gitops

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/setupx/internal/util"
)

func Sync(url string) (string, string, error) {
	h := sha1.Sum([]byte(url))
	id := hex.EncodeToString(h[:8])
	cacheRoot, err := util.CacheDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(cacheRoot, id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "--depth", "1", url, dir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", "", fmt.Errorf("git clone failed: %v: %s", err, string(out))
		}
		return id, dir, nil
	}
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("git pull failed: %v: %s", err, string(out))
	}
	return id, dir, nil
}
