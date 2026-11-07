package tsgo_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Main owns the cache; this adapter verifies every member of the test bundle.
func bridgeProductGet(key string, build func(string) error) (string, error) {
	directory, err := buildcache.Get(buildcache.Inputs{Name: "bridge-test-bundle", Flags: []string{key}}, build)
	if err == nil {
		err = bridgeProductCheck(directory)
	}
	return directory, err
}

func bridgeProductCheck(directory string) error {
	data, readError := os.ReadFile(filepath.Join(directory, "products.json"))
	if readError != nil {
		return readError
	}
	var hashes map[string]string
	if decodeError := json.Unmarshal(data, &hashes); decodeError != nil {
		return decodeError
	}
	if len(hashes) != len(bridgeProductNames) {
		return fmt.Errorf("product count: %d, want %d", len(hashes), len(bridgeProductNames))
	}
	for _, name := range bridgeProductNames {
		data, readError := os.ReadFile(filepath.Join(directory, name))
		if readError != nil {
			return readError
		}
		if hashes[name] != fmt.Sprintf("%x", sha256.Sum256(data)) {
			return fmt.Errorf("product digest: %s", name)
		}
	}
	return nil
}
