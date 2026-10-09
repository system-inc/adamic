package tsgo_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// This is the sole cache adapter. When buildcache lands, replace this call with
// buildcache.Get(buildcache.Inputs{Name: "tsgo-test-products", Flags: []string{key}}, build).
func bridgeProductGet(key string, build func(string) error) (string, error) {
	directory, err := localBridgeProductGet(key, build)
	if err == nil {
		err = bridgeProductCheck(directory)
	}
	return directory, err
}

func localBridgeProductGet(key string, build func(string) error) (string, error) {
	cache := os.Getenv("ADAMIC_BUILD_CACHE_DIR")
	if cache == "" {
		user, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		cache = filepath.Join(user, "adamic-bridge-test-products")
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", err
	}
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		directory, err := os.MkdirTemp(cache, ".uncached-")
		if err != nil {
			return "", err
		}
		return directory, build(directory)
	}
	directory := filepath.Join(cache, key)
	lock, err := os.OpenFile(directory+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return "", err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return "", err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if err := bridgeProductCheck(directory); err == nil {
		return directory, nil
	}
	if _, err := os.Stat(directory); err == nil {
		return "", fmt.Errorf("bridge product %s is incomplete or corrupt; remove only that product and prepare again", key)
	}
	scratch, err := os.MkdirTemp(cache, ".building-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(scratch)
	if err := build(scratch); err != nil {
		return "", err
	}
	// Publish the complete set atomically. A failed build never becomes a hit.
	if err := os.Rename(scratch, directory); err != nil {
		return "", err
	}
	return directory, nil
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
