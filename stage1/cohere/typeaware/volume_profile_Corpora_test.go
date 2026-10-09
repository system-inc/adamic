package typeaware

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Each range has independent plain and sanitized leaves. Compiler roots are
// deliberately fine grained: checker.ts costs much more than a typical root.
const testVolumeProfileCorporaShards = 2 * (32 + 77)

func volumeProfileCorporaNative(h *harness, stage0, archive string, sanitize bool) string {
	h.t.Helper()
	name := "typeaware volume"
	if sanitize {
		name += " asan"
	}
	digest := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		return fmt.Sprintf("%x", sha256.Sum256(data))
	}
	// The stage0 executable includes the lowerer, native emitter, runtime and
	// prelude. The archive digest includes all checker sources and build flags.
	inputs := buildcache.Inputs{
		Name:      name,
		Flags:     []string{"build", "stage1/cohere/typeaware/volume_suite.ts", "--tsgo", "stage0=" + digest(stage0), "archive=" + digest(archive), fmt.Sprintf("sanitize=%t", sanitize), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")},
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	for _, directory := range []string{"stage1/cohere/typeaware", "stage1/cohere/lint", "stage1/typescript"} {
		err := filepath.WalkDir(filepath.Join(h.repository, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				relative, err := filepath.Rel(h.repository, path)
				if err != nil {
					return err
				}
				inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			h.t.Fatal(err)
		}
	}
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		builder := &harness{t: h.t, repository: h.repository, directory: directory}
		builder.build(stage0, "volume", filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), archive, sanitize)
		return nil
	})
	return filepath.Join(directory, "volume")
}

func TestVolumeProfileCorpora(t *testing.T) {
	started := time.Now()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	type shard struct {
		corpus, config string
		paths          []string
		sanitize       bool
		enabled        bool
	}
	var shards []shard
	for _, corpus := range []struct {
		name, manifest, config string
		ranges                 int
	}{
		{"repository", os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"), filepath.Join(repository, "tsconfig.json"), 32},
		{"compiler", os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), 77},
	} {
		var paths []string
		if corpus.manifest != "" {
			data, err := os.ReadFile(corpus.manifest)
			if err != nil {
				t.Fatal(err)
			}
			// Match both original consumers: only empty lines are omitted; retain
			// manifest order, relative paths, whitespace and duplicate occurrences.
			for _, path := range strings.Split(string(data), "\n") {
				if path != "" {
					paths = append(paths, path)
				}
			}
		}
		for i := 0; i < corpus.ranges; i++ {
			part := paths[len(paths)*i/corpus.ranges : len(paths)*(i+1)/corpus.ranges]
			for _, sanitize := range []bool{false, true} {
				shards = append(shards, shard{corpus.name, corpus.config, part, sanitize, corpus.manifest != ""})
			}
		}
		for _, sanitize := range []bool{false, true} {
			var union []string
			for _, s := range shards {
				if s.corpus == corpus.name && s.sanitize == sanitize {
					union = append(union, s.paths...)
				}
			}
			if !slices.Equal(union, paths) {
				t.Fatalf("%s sanitize=%t shard union differs from manifest", corpus.name, sanitize)
			}
			t.Logf("%s sanitize=%t union: %d manifest files in %d contiguous ranges", corpus.name, sanitize, len(union), corpus.ranges)
		}
	}
	if len(shards) != testVolumeProfileCorporaShards {
		t.Fatalf("enumerated %d shards; want %d", len(shards), testVolumeProfileCorporaShards)
	}
	var binary, asan, oracle string
	if os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST") != "" || os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST") != "" {
		h := &harness{t: t, repository: repository, directory: t.TempDir()}
		// These are Go builds; keep the original commands until GoBuild lands.
		stage0 := filepath.Join(h.directory, "adamic")
		h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
		archive := h.archive("checker", "", false)
		sanitized := h.archive("checker-asan", "", true)
		oracle = volumeOracle(h, "volume-oracle", "oracle_volume.go")
		binary = volumeProfileCorporaNative(h, stage0, archive, false)
		asan = volumeProfileCorporaNative(h, stage0, sanitized, true)
	}
	t.Logf("TestVolumeProfileCorpora (setup): %.3fs", time.Since(started).Seconds())
	for i, s := range shards {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			if !s.enabled {
				t.Skip("set ADAMIC_VOLUME_" + strings.ToUpper(s.corpus) + "_MANIFEST")
			}
			if len(s.paths) == 0 {
				t.Skip("empty contiguous range")
			}
			h := &harness{t: t, repository: repository, directory: t.TempDir()}
			manifest := h.write("corpus.manifest", strings.Join(s.paths, "\n")+"\n")
			executable, name := binary, s.corpus
			if s.sanitize {
				executable, name = asan, name+"-asan"
			}
			t.Logf("%s: %d files, first=%s last=%s", name, len(s.paths), s.paths[0], s.paths[len(s.paths)-1])
			started := time.Now()
			want := h.compare(name, oracle, executable, s.config, manifest)
			t.Logf("comparison sides: oracle %.3fs, native including byte check %.3fs", want.elapsed.Seconds(), time.Since(started).Seconds()-want.elapsed.Seconds())
		})
	}
}
