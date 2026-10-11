package typeaware

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
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
	return h.stage0()
}

// typeAwareCompilerFiles are the sources a volume product's compilers are built from: stage 0 (./cmd/adamic) and the
// checker archive (./bridge/tsgo/archive), each module file that pins what they import from outside the tree, and the
// rule packages cohere's oracle reads.
var typeAwareCompilerFiles = []string{"go.mod", "go.work", "cmd/adamic", "internal", "bridge/tsgo", "cohere/go.mod", "cohere/go.sum", "cohere/internal", "cohere/policy", "cohere/rule_runner", "cohere/TypeScript/tsc/internal", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum", "cohere/TypeScript-shim"}

// typeAwareCompilerKey keys a volume product by its compilers' sources, the TypeScript it compiles, the Go release and
// the C compiler the archive's cgo runs. It asks go nothing: go env and go list answer for the machine and checkout they
// run in, and a runner's answers keyed Workshop's products apart (#nm31pcn: typeaware_volume_lowered, and the corpus's
// stage 0, archives and oracle, on landable-6's tree a221706a).
func typeAwareCompilerKey(h *harness, inputs *buildcache.Inputs, sanitize bool) {
	h.t.Helper()
	inputs.Files = append(inputs.Files, typeAwareCompilerFiles...)
	for _, directory := range []string{"stage1/cohere/typeaware", "stage1/cohere/lint", "stage1/typescript"} {
		err := filepath.WalkDir(filepath.Join(h.repository, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Only what the tree carries: lint's tests write stage1/cohere/lint/.generated/registry.ts into a checkout,
			// which a runner's source never has, and keyed it on Workshop alone (Planner's diff of landable-7's tree).
			if !buildcache.Tracked(h.repository, path, entry.IsDir()) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
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
	// cgo's C compiler: cc where CC is unset, as on every Loom box, and clang for the sanitized archive.
	compiler := "cc"
	if sanitize {
		compiler = "clang"
	}
	inputs.Toolchain = []string{runtime.Version(), buildcache.Tool("clang", "--version"), buildcache.Tool(compiler, "--version")}
}

func volumeProfileNativeInputs(h *harness, stage0, archive string, sanitize bool, name string, variant ...string) buildcache.Inputs {
	h.t.Helper()
	archiveName, archiveFlags := "typeaware checker archive", ""
	if sanitize {
		archiveName += " asan"
		archiveFlags = "CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"
	}
	inputs := buildcache.Inputs{
		Name:  name,
		Flags: []string{"build", "stage1/cohere/typeaware/volume_suite.ts", "--tsgo", "go build ./cmd/adamic", "archive=" + archiveName, "go build -buildmode=c-archive ./bridge/tsgo/archive", archiveFlags, fmt.Sprintf("sanitize=%t", sanitize), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")},
	}
	typeAwareCompilerKey(h, &inputs, sanitize)
	inputs.Flags = append(inputs.Flags, variant...)
	return inputs
}

func volumeProfileNative(h *harness, stage0, archive string, sanitize bool, name string, build func(*harness) string, variant ...string) string {
	inputs := volumeProfileNativeInputs(h, stage0, archive, sanitize, name, variant...)
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		local := &harness{t: h.t, repository: h.repository, directory: directory}
		product := build(local)
		return os.Rename(product, filepath.Join(directory, "volume"))
	})
	return filepath.Join(directory, "volume")
}

// This is compileLibrary + TSGoC from cmd/adamic, split before clang so each
// cold build unit can finish under the deadline. Both sanitizer modes use the
// exact same lowered C and the original BuildTSGo flags and runtime.
func volumeProfileControlsLowered(h *harness, stage0 string) string {
	inputs := volumeProfileNativeInputs(h, stage0, "", false, "typeaware volume lowered")
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		program, err := load.Load([]string{filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts")})
		if err != nil {
			return err
		}
		program.EnableTSGo()
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		source, err := native.TSGoC(ir)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "volume.c"), []byte(source), 0600)
	})
	return filepath.Join(directory, "volume.c")
}

func TestVolumeProfileControlsLower(t *testing.T) {
	t.Parallel()
	h := volumeProfileHarness(t)
	volumeProfileControlsLowered(h, volumeProfileStage0(h))
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

func volumeProfileControlsBinary(h *harness, stage0, archive string, sanitize bool) string {
	t := h.t
	name := "typeaware volume"
	if sanitize {
		name += " asan"
	}
	lowered := volumeProfileControlsLowered(h, stage0)
	binary := volumeProfileNative(h, stage0, archive, sanitize, name, func(local *harness) string {
		source, err := os.ReadFile(lowered)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(local.directory, "volume")
		if err := typeAwareNativeBuild(string(source), path, archive, sanitize); err != nil {
			t.Fatal(err)
		}
		return path
	})
	return binary
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
	binary := volumeProfileControlsBinary(h, stage0, archive, c.sanitize)
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
