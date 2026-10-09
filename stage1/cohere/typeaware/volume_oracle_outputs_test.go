package typeaware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Compiler input data can live outside the repository. Its content fingerprint
// goes into Flags, just as fetched binary/archive fingerprints do. Include all
// potential source/config data, not only the roots being diagnosed: imported
// modules and config declarations can change an oracle's answer for a root.
func volumeOracleData(t *testing.T, name, config string) buildcache.Inputs {
	t.Helper()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	fingerprint := func(root string) {
		seen := map[string]bool{}
		var walk func(string) error
		walk = func(root string) error {
			real, err := filepath.EvalSymlinks(root)
			if err != nil {
				return err
			}
			if seen[real] {
				return nil
			}
			seen[real] = true
			return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() && entry.Name() == ".git" {
					return filepath.SkipDir
				}
				if entry.Type()&os.ModeSymlink != 0 {
					target, err := os.Readlink(path)
					if err != nil {
						return err
					}
					fmt.Fprintf(hash, "link %q %q\n", path, target)
					info, err := os.Stat(path)
					if os.IsNotExist(err) {
						return nil
					}
					if err != nil {
						return err
					}
					if info.IsDir() {
						return walk(path)
					}
				}
				if entry.IsDir() {
					return nil
				}
				switch strings.ToLower(filepath.Ext(path)) {
				case ".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs", ".json":
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					sum := sha256.Sum256(data)
					fmt.Fprintf(hash, "file %q %x\n", path, sum)
				}
				return nil
			})
		}
		if err := walk(root); err != nil {
			t.Fatal(err)
		}
	}
	if config != "" {
		fingerprint(repository)
		if name == "compiler" {
			corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
			if corpus == "" {
				t.Fatal("compiler output requires ADAMIC_TYPESCRIPT_SOURCE")
			}
			fingerprint(filepath.Join(corpus, "src"))
			// Package resolution can also consult the checkout's package metadata.
			if data, err := os.ReadFile(filepath.Join(corpus, "package.json")); err == nil {
				fmt.Fprintf(hash, "package %x\n", sha256.Sum256(data))
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	}
	return buildcache.Inputs{Name: "typeaware-volume-oracle-output-" + name, Flags: []string{fmt.Sprintf("source-config-data=%x", hash.Sum(nil))}, Toolchain: []string{buildcache.Tool("go", "version")}}
}

// Only deterministic finding bytes are a product. The oracle's stderr contains
// performance counters, so it is retained on errors but never cached as output.
func volumeOracleOutput(h *harness, base buildcache.Inputs, oracle, config, manifest string) result {
	inputs := base
	inputs.Flags = append([]string(nil), base.Flags...)
	oracleHash, err := volumeFileHash(oracle)
	if err != nil {
		h.t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, "oracle="+oracleHash, "config-path="+config)
	if config != "" {
		sum, err := volumeFileHash(config)
		if err != nil {
			h.t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, "config="+sum)
	}
	data, err := os.ReadFile(manifest)
	if err != nil {
		h.t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, "manifest="+string(data))
	for _, path := range strings.Split(string(data), "\n") {
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(config), path)
		}
		sum, err := volumeFileHash(path)
		if err != nil {
			h.t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, "root="+path+" sha256="+sum)
	}
	started := time.Now()
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		var stdout, stderr bytes.Buffer
		command := exec.Command(oracle, config, manifest)
		command.Dir = h.repository
		command.Stdout = &stdout
		command.Stderr = &stderr
		began := time.Now()
		err := command.Run()
		h.t.Logf("volume-build oracle-output cold=%.6fs", time.Since(began).Seconds())
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, stderr.Bytes())
		}
		return os.WriteFile(filepath.Join(directory, "findings"), stdout.Bytes(), 0444)
	})
	stdout, err := os.ReadFile(filepath.Join(directory, "findings"))
	if err != nil {
		h.t.Fatal(err)
	}
	return result{stdout: stdout, elapsed: time.Since(started)}
}
