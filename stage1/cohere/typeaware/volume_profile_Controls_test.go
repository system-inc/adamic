package typeaware

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Generated controls are copied from profile_test.go at 7a10c877. Each mode
// owns the whole manifest; adding a source never moves an existing check.
const testVolumeProfileControlsShards = 2

// Keep the whole manifest together: its native global declaration is visible to
// native-console, and the original oracle checks the combined finding bytes.
func volumeProfileControlSources() []string {
	return append(volumeControls(),
		"function first(){const value=1;}function second(){const value=2;}",
		"const X=1;const C=class X<X>{m(){const X=2;}};",
		"const a=1;const b=2;function f(a:number,b:number){return a+b;}")
}

func volumeProfileControlsManifest(h *harness) string {
	sources := volumeProfileControlSources()
	var paths []string
	for i, text := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), text+"\nexport {};\n"))
	}
	paths = append(paths, h.write("native-globals.d.ts", "declare const console: {log():void};\n"), h.write("native-console.ts", "const detached=console.log;\nexport {};\n"))
	return h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}

func volumeProfileHarness(t *testing.T) *harness {
	t.Helper()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, repository: repository, directory: t.TempDir()}
}

// Go products retain the original build commands until GoBuild is available.
func volumeProfileStage0(h *harness) string {
	stage0 := filepath.Join(h.directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	return stage0
}

func volumeProfileNative(h *harness, stage0, archive string, sanitize bool, name string, build func(*harness) string, variant ...string) string {
	h.t.Helper()
	digest := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		return fmt.Sprintf("%x", sha256.Sum256(data))
	}
	// The stage0 executable includes the lowerer, emitter, runtime and prelude.
	// Key the Go archive by its source recipe: c-archive bytes include scratch
	// paths, so their digest would prevent reuse across otherwise identical builds.
	archiveName, archiveFlags := "typeaware checker archive", ""
	cc, err := exec.Command("go", "env", "CC").Output()
	if err != nil {
		h.t.Fatal(err)
	}
	if sanitize {
		archiveName += " asan"
		archiveFlags = "CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"
		cc = []byte("clang")
	}
	inputs := buildcache.Inputs{
		Name:      name,
		Flags:     []string{"build", "stage1/cohere/typeaware/volume_suite.ts", "--tsgo", "stage0=" + digest(stage0), "archive=" + archiveName, "go build -buildmode=c-archive ./bridge/tsgo/archive", archiveFlags, fmt.Sprintf("sanitize=%t", sanitize), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")},
		Files:     []string{"bridge/tsgo", "cohere/TypeScript/tsc/internal", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum", "cohere/TypeScript-shim", "go.mod", "go.sum", "cohere/go.mod", "cohere/go.sum"},
		Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool(strings.Fields(string(cc))[0], "--version"), buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOFLAGS", "GOEXPERIMENT")},
	}
	// A module without external dependencies need not have a go.sum. Hash every
	// manifest that exists; adding one automatically changes the input recipe.
	files := inputs.Files[:0]
	for _, path := range inputs.Files {
		if _, err := os.Stat(filepath.Join(h.repository, path)); os.IsNotExist(err) {
			continue
		} else if err != nil {
			h.t.Fatal(err)
		}
		files = append(files, path)
	}
	inputs.Files = files
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
	inputs.Flags = append(inputs.Flags, variant...)
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		local := &harness{t: h.t, repository: h.repository, directory: directory}
		product := build(local)
		return os.Rename(product, filepath.Join(directory, "volume"))
	})
	return filepath.Join(directory, "volume")
}

type volumeProfileControlCase struct {
	name     string
	sanitize bool
}

func volumeProfileControlsCases() []volumeProfileControlCase {
	return []volumeProfileControlCase{{"controls", false}, {"controls-asan", true}}
}

func TestVolumeProfileControlsUnion(t *testing.T) {
	t.Parallel()
	cases := volumeProfileControlsCases()
	if len(cases) != testVolumeProfileControlsShards {
		t.Fatalf("controls enumeration: %d != %d", len(cases), testVolumeProfileControlsShards)
	}
	seen := map[string]int{}
	var keys []string
	for i := range volumeProfileControlSources() {
		keys = append(keys, fmt.Sprintf("control-%03d.ts", i))
	}
	keys = append(keys, "native-globals.d.ts", "native-console.ts")
	for _, c := range cases {
		for _, key := range keys {
			seen[fmt.Sprintf("%s/asan=%t", key, c.sanitize)]++
		}
	}
	for _, key := range keys {
		for _, mode := range []bool{false, true} {
			if count := seen[fmt.Sprintf("%s/asan=%t", key, mode)]; count != 1 {
				t.Fatalf("control union %s asan=%t: %d owners", key, mode, count)
			}
		}
	}
	t.Logf("controls union: %d live source/mode cases in %d shards", len(seen), len(cases))
}

func runVolumeProfileControls(t *testing.T, index int) {
	started := time.Now()
	c := volumeProfileControlsCases()[index]
	h := volumeProfileHarness(t)
	var stage0, archive, oracle string
	var builds sync.WaitGroup
	for _, job := range []struct {
		product *string
		build   func(*harness) string
	}{
		{&stage0, volumeProfileStage0},
		{&archive, func(h *harness) string {
			name := "checker"
			if c.sanitize {
				name += "-asan"
			}
			return h.archive(name, "", c.sanitize)
		}},
		{&oracle, func(h *harness) string { return volumeOracle(h, "volume-oracle", "oracle_volume.go") }},
	} {
		local := volumeProfileHarness(t)
		builds.Add(1)
		go func() { defer builds.Done(); *job.product = job.build(local) }()
	}
	builds.Wait()
	if t.Failed() {
		return
	}
	t.Logf("setup: %.3fs", time.Since(started).Seconds())
	name := "typeaware volume"
	if c.sanitize {
		name += " asan"
	}
	binary := volumeProfileNative(h, stage0, archive, c.sanitize, name, func(local *harness) string {
		return local.build(stage0, "volume", filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), archive, c.sanitize)
	})
	checkStarted := time.Now()
	h.compare(c.name, oracle, binary, filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/tsconfig.json"), volumeProfileControlsManifest(h))
	t.Logf("check after product fetch: %.3fs; cooked=false", time.Since(checkStarted).Seconds())
}

func TestVolumeProfileControls_000(t *testing.T) { t.Parallel(); runVolumeProfileControls(t, 0) }
func TestVolumeProfileControls_001(t *testing.T) { t.Parallel(); runVolumeProfileControls(t, 1) }

const testShadowIndexMissingBindingShards = 1

func TestShadowIndexMissingBindingUnion(t *testing.T) {
	t.Parallel()
	cases := []string{"missing-binding"}
	if len(cases) != testShadowIndexMissingBindingShards {
		t.Fatalf("missing-binding enumeration: %d != %d", len(cases), testShadowIndexMissingBindingShards)
	}
	seen := map[string]int{}
	for _, name := range cases {
		seen[name]++
	}
	for name, count := range seen {
		if count != 1 {
			t.Fatalf("%s: %d owners", name, count)
		}
	}
	t.Logf("missing-binding union: %d cases", len(cases))
}
