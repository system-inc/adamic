package typeaware

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// typeAwareBuildSetting is name as toolchain.Env sets it, the last setting winning as it does for a program run with
// it: what a product's go build runs with, which the caller's own environment may not say. CC is unset on every Loom
// box, so a key reading the caller's CC said nothing, while Env names clang and cgo ran it (Tests, Oct 12).
func typeAwareBuildSetting(name string) string {
	environment := toolchain.Env()
	for index := len(environment) - 1; index >= 0; index-- {
		if value, ok := strings.CutPrefix(environment[index], name+"="); ok {
			return value
		}
	}
	return ""
}

// typeAwareBuildSettings are names as toolchain.Env sets them, each NAME=value, for a key's Flags.
func typeAwareBuildSettings(names ...string) []string {
	settings := make([]string, 0, len(names))
	for _, name := range names {
		settings = append(settings, name+"="+typeAwareBuildSetting(name))
	}
	return settings
}

// typeAwareCompilerReport is compiler's own report, by Tool: compiler is a CC value, a program and its arguments.
func typeAwareCompilerReport(compiler string) string {
	fields := strings.Fields(compiler)
	if len(fields) == 0 {
		return "CC: unset"
	}
	return buildcache.Tool(fields[0], append(fields[1:], "--version")...)
}

// typeAwareSharedInputs keys a shared product: stage 0 and the checker archives, built by go under toolchain.Env, whose
// cgo runs Env's CC, and the emitted C and native binaries, which native compiles with clang.
func typeAwareSharedInputs(key string) buildcache.Inputs {
	return buildcache.Inputs{
		Name: "typeaware shared " + key,
		Files: []string{
			"go.mod",
			"go.work",
			"internal",
			"cmd/adamic",
			"bridge/tsgo",
			"stage1/cohere/typeaware",
			"stage1/typescript",
			"cohere/go.mod",
			"cohere/go.sum",
			"cohere/internal",
			"cohere/TypeScript/tsc",
		},
		Flags: append(
			typeAwareBuildSettings("CC", "CGO_CFLAGS", "CGO_LDFLAGS", "GOFLAGS", "GOOS", "GOARCH", "CGO_ENABLED"),
			"go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "),
		),
		Toolchain: []string{
			runtime.Version(),
			buildcache.Tool("go", "version"),
			buildcache.Tool("clang", "--version"),
			typeAwareCompilerReport(typeAwareBuildSetting("CC")),
		},
	}
}

// Every key whose go build runs cgo names the compiler that build runs: Env's clang where the caller sets no CC, and
// the caller's where it does, so a changed compiler misses. The shared products, the volume guard's and the volume
// profile's (controls and corpus) are each asked. Not parallel: it sets CC. Mutant: the caller's CC in the key, which
// is empty on a Loom box while cgo runs clang.
func TestTypeAwareKeysNameTheBuildsCompiler(test *testing.T) {
	keys := []struct {
		name   string
		inputs func() buildcache.Inputs
	}{
		{"shared", func() buildcache.Inputs { return typeAwareSharedInputs("compiler probe") }},
		{
			"volume guard",
			func() buildcache.Inputs { return volumeGuardInputs(context.Background(), "compiler probe", false) },
		},
		{"volume profile", func() buildcache.Inputs {
			return volumeProfileNativeInputs(volumeProfileHarness(test), "", "", false, "compiler probe")
		}},
	}
	test.Setenv("CC", "")
	if unsetenvError := os.Unsetenv("CC"); unsetenvError != nil {
		test.Fatal(unsetenvError)
	}
	unset := make([]buildcache.Inputs, len(keys))
	for index, key := range keys {
		unset[index] = key.inputs()
		if !slices.Contains(unset[index].Flags, "CC=clang") {
			test.Errorf(
				"%s: with CC unset the key's flags %q don't name clang, which cgo runs",
				key.name,
				unset[index].Flags,
			)
		}
	}
	test.Setenv("CC", "cc")
	for index, key := range keys {
		changed := key.inputs()
		if !slices.Contains(changed.Flags, "CC=cc") ||
			slices.Equal(changed.Flags, unset[index].Flags) ||
			slices.Equal(changed.Toolchain, unset[index].Toolchain) {
			test.Errorf(
				"%s: a changed compiler keys the same: flags %q, toolchain %q",
				key.name,
				changed.Flags,
				changed.Toolchain,
			)
		}
	}
}

// Stage 0, native binaries and emitted C are immutable build products.
// Keep the process-local publication guard, backed by content-addressed products
// across gate processes. The checker archives and Go oracles are buildcache.GoBuild* products.
func cachedTypeAwareProduct(key string, build func(string) (string, error)) (string, error) {
	return sharedProduct("cached "+key, func(_ string) (string, error) {
		inputs := typeAwareSharedInputs(key)
		directory, getError := buildcache.Get(inputs, func(directory string) error {
			value, buildError := build(directory)
			if buildError != nil {
				return buildError
			}
			record := struct {
				Value string
				File  bool
			}{Value: value}
			if filepath.IsAbs(value) {
				relative, relativePathError := filepath.Rel(directory, value)
				if relativePathError != nil {
					return relativePathError
				}
				record.Value, record.File = relative, true
			}
			data, encodeError := json.Marshal(record)
			if encodeError != nil {
				return encodeError
			}
			return os.WriteFile(filepath.Join(directory, "product.json"), data, 0644)
		})
		if getError != nil {
			return "", getError
		}
		data, readError := os.ReadFile(filepath.Join(directory, "product.json"))
		if readError != nil {
			return "", readError
		}
		var record struct {
			Value string
			File  bool
		}
		if decodeError := json.Unmarshal(data, &record); decodeError != nil {
			return "", decodeError
		}
		if record.File {
			return filepath.Join(directory, record.Value), nil
		}
		return record.Value, nil
	})
}
