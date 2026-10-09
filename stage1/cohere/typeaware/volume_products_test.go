package typeaware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func volumeCommand(directory string, command *exec.Cmd) error {
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", command, err, output)
	}
	return nil
}

func volumeStage0(t *testing.T) string { return buildcache.GoBuild(t, "adamic", "./cmd/adamic", nil) }

func volumeArchive(h *harness, name, overlay string, sanitize bool) string {
	if overlay != "" {
		return volumePreparedInput(h.t, name)
	}
	if sanitize {
		return buildcache.GoBuild(h.t, "tsgo-asan.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	return buildcache.GoBuild(h.t, "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
}

func volumeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// The callback writes only its product directory. Flags key fetched build inputs
// by bytes, not by cache location; Files cover transitive TypeScript imports.
func volumeNativeSpec(repository, stage0, name, entry, archive string, sanitize bool) (buildcache.Inputs, func(string) error, error) {
	var files []string
	for _, root := range []string{"stage1/typescript", "stage1/cohere/lint", "stage1/cohere/typeaware"} {
		if err := filepath.WalkDir(filepath.Join(repository, root), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				files = append(files, relative)
			}
			return nil
		}); err != nil {
			return buildcache.Inputs{}, nil, err
		}
	}
	var flags []string
	for _, input := range []struct{ name, path string }{{"stage0", stage0}, {"archive", archive}, {"entry", entry}} {
		sum, err := volumeFileHash(input.path)
		if err != nil {
			return buildcache.Inputs{}, nil, err
		}
		flags = append(flags, input.name+"="+sum)
	}
	flags = append(flags, fmt.Sprintf("sanitize=%t", sanitize))
	for _, key := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		flags = append(flags, key+"="+os.Getenv(key))
	}
	inputs := buildcache.Inputs{Name: "typeaware-" + name, Files: files, Flags: flags, Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool("clang", "-print-search-dirs"), buildcache.Tool("clang", "-print-resource-dir")}}
	build := func(directory string) error {
		args := []string{"build", entry, "-o", filepath.Join(directory, "native"), "--tsgo", archive}
		if sanitize {
			args = append(args, "--sanitize")
		}
		started := time.Now()
		err := volumeCommand(repository, exec.Command(stage0, args...))
		fmt.Printf("volume-build %s cold=%.6fs\n", name, time.Since(started).Seconds())
		return err
	}
	return inputs, build, nil
}

func volumeNative(h *harness, stage0, name, entry, archive string, sanitize, private bool) string {
	if private {
		return volumePreparedInput(h.t, name)
	}
	inputs, build, err := volumeNativeSpec(h.repository, stage0, name, entry, archive, sanitize)
	if err != nil {
		h.t.Fatal(err)
	}
	return filepath.Join(buildcache.Product(h.t, inputs, build), "native")
}

// A package-local child module provides legitimate access to cohere's internal
// production rules without an overlay. Require its oracle body to remain exactly
// the existing oracle body, so the independent reference cannot silently drift.
func volumeOracleWorkspace(repository string) (string, error) {
	paths := []string{"stage1/cohere/typeaware/testdata/oracle_volume.go", "stage1/cohere/typeaware/testdata/volume-oracle-go/main.go"}
	var bodies [][]byte
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(repository, path))
		if err != nil {
			return "", err
		}
		at := bytes.Index(data, []byte("package main"))
		if at < 0 {
			return "", fmt.Errorf("oracle %s lacks package main", path)
		}
		bodies = append(bodies, data[at:])
	}
	if !bytes.Equal(bodies[0], bodies[1]) {
		return "", fmt.Errorf("package-local oracle body differs from original oracle_volume.go")
	}
	return filepath.Join(repository, "stage1/cohere/typeaware/testdata/volume-oracle-go/go.work"), nil
}

func volumeProductOracle(h *harness) string {
	workspace, err := volumeOracleWorkspace(h.repository)
	if err != nil {
		h.t.Fatal(err)
	}
	return buildcache.GoBuild(h.t, "volume-oracle", "./stage1/cohere/typeaware/testdata/volume-oracle-go", nil, "GOWORK="+workspace)
}
