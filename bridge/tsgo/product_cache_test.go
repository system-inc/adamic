package tsgo_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Retrieve shared products through the common build cache, then verify their bytes.
func bridgeProductGet(key string, build func(string) error) (string, error) {
	directory, err := buildcache.Get(buildcache.Inputs{Name: "tsgo-test-products", Flags: []string{key}}, build)
	if err == nil {
		err = bridgeProductCheck(directory)
	}
	return directory, err
}

func bridgeProductCheck(directory string) error {
	data, err := os.ReadFile(filepath.Join(directory, "products.json"))
	if err != nil {
		return err
	}
	var hashes map[string]string
	if err := json.Unmarshal(data, &hashes); err != nil {
		return err
	}
	if len(hashes) != len(bridgeProductNames) {
		return fmt.Errorf("product count: %d, want %d", len(hashes), len(bridgeProductNames))
	}
	for _, name := range bridgeProductNames {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return err
		}
		if hashes[name] != fmt.Sprintf("%x", sha256.Sum256(data)) {
			return fmt.Errorf("product digest: %s", name)
		}
	}
	return nil
}
