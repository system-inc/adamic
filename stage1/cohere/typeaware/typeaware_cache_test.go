package typeaware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Common archives, native binaries and emitted C are immutable build products.
// Keep the process-local publication guard, backed by content-addressed products
// across gate processes. Oracle commands use their more precise sixBuildRecipe.
func cachedTypeAwareProduct(key string, build func(string) (string, error)) (string, error) {
	return sharedProduct("cached "+key, func(_ string) (string, error) {
		inputs := buildcache.Inputs{
			Name:      "typeaware shared " + key,
			Files:     []string{"go.mod", "go.work", "internal", "cmd/adamic", "bridge/tsgo", "stage1/cohere/typeaware", "stage1/typescript", "cohere/go.mod", "cohere/go.sum", "cohere/internal", "cohere/TypeScript/tsc"},
			Flags:     []string{os.Getenv("CC"), os.Getenv("CGO_CFLAGS"), os.Getenv("CGO_LDFLAGS"), os.Getenv("GOFLAGS"), os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("CGO_ENABLED")},
			Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version"), buildcache.Tool("cc", "--version")},
		}
		directory, err := buildcache.Get(inputs, func(directory string) error {
			value, err := build(directory)
			if err != nil {
				return err
			}
			record := struct {
				Value string
				File  bool
			}{Value: value}
			if filepath.IsAbs(value) {
				relative, err := filepath.Rel(directory, value)
				if err != nil {
					return err
				}
				record.Value, record.File = relative, true
			}
			data, err := json.Marshal(record)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "product.json"), data, 0644)
		})
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(filepath.Join(directory, "product.json"))
		if err != nil {
			return "", err
		}
		var record struct {
			Value string
			File  bool
		}
		if err := json.Unmarshal(data, &record); err != nil {
			return "", err
		}
		if record.File {
			return filepath.Join(directory, record.Value), nil
		}
		return record.Value, nil
	})
}
