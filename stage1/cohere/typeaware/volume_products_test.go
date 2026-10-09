package typeaware

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func volumeCommand(directory string, command *exec.Cmd) error {
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", command, err, output)
	}
	return nil
}

func volumeStage0(t *testing.T) string {
	return buildcache.GoBuild(t, "adamic", "./cmd/adamic", nil)
}

func volumeArchive(h *harness, name, overlay string, sanitize bool) string {
	if overlay != "" {
		started := time.Now()
		path := h.archive(name, overlay, sanitize)
		h.t.Logf("volume-build %s private cold=%.6fs", name, time.Since(started).Seconds())
		return path
	}
	if sanitize {
		return buildcache.GoBuild(h.t, "tsgo-asan.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	return buildcache.GoBuild(h.t, "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
}

func volumeDigest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// Native products key their compiled inputs by content, not the location of a
// fetched stage 0 or archive. Every transitive TypeScript source is included.
func volumeNative(h *harness, stage0, name, entry, archive string, sanitize bool, private bool) string {
	if private {
		started := time.Now()
		path := h.build(stage0, name, entry, archive, sanitize)
		h.t.Logf("volume-build %s private cold=%.6fs", name, time.Since(started).Seconds())
		return path
	}
	var files []string
	for _, root := range []string{"stage1/typescript", "stage1/cohere/lint", "stage1/cohere/typeaware"} {
		err := filepath.WalkDir(filepath.Join(h.repository, root), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				relative, err := filepath.Rel(h.repository, path)
				if err != nil {
					return err
				}
				files = append(files, relative)
			}
			return nil
		})
		if err != nil {
			h.t.Fatal(err)
		}
	}
	flags := []string{"stage0=" + volumeDigest(h.t, stage0), "archive=" + volumeDigest(h.t, archive), "entry=" + volumeDigest(h.t, entry), fmt.Sprintf("sanitize=%t", sanitize)}
	for _, key := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		flags = append(flags, key+"="+os.Getenv(key))
	}
	inputs := buildcache.Inputs{Name: "typeaware-" + name, Files: files, Flags: flags,
		Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool("clang", "-print-search-dirs"), buildcache.Tool("clang", "-print-resource-dir")}}
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		args := []string{"build", entry, "-o", filepath.Join(directory, "native"), "--tsgo", archive}
		if sanitize {
			args = append(args, "--sanitize")
		}
		started := time.Now()
		err := volumeCommand(h.repository, exec.Command(stage0, args...))
		h.t.Logf("volume-build %s cold=%.6fs", name, time.Since(started).Seconds())
		return err
	})
	return filepath.Join(directory, "native")
}

func volumeProductOracle(h *harness) string {
	// The virtual main is an overlay, but its entire replacement and dependency
	// closure is in the repository and explicitly keyed. Mutant overlays are private.
	inputs := buildcache.Inputs{Name: "typeaware-volume-oracle",
		Files: []string{"cohere/go.mod", "cohere/go.sum", "cohere/internal", "cohere/TypeScript-shim", "cohere/TypeScript/tsc", "stage1/cohere/typeaware/testdata/oracle_volume.go"},
		Flags: []string{buildcache.Tool("go", "env", "-json")}, Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		virtual := filepath.Join(h.repository, "cohere/adamic_volume_oracle.go")
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/oracle_volume.go")}})
		if err != nil {
			return err
		}
		overlay := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(overlay, data, 0600); err != nil {
			return err
		}
		started := time.Now()
		err = volumeCommand(filepath.Join(h.repository, "cohere"), exec.Command("go", "build", "-overlay", overlay, "-o", filepath.Join(directory, "oracle"), virtual))
		h.t.Logf("volume-build oracle cold=%.6fs", time.Since(started).Seconds())
		// Paths in an overlay are local build instructions, not shared build products.
		if removeErr := os.Remove(overlay); err == nil {
			err = removeErr
		}
		return err
	})
	return filepath.Join(directory, "oracle")
}
