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

func volumeProfileControlsManifest(harness *harness) string {
	sources := volumeProfileControlSources()
	var paths []string
	for index, text := range sources {
		paths = append(paths, harness.write(fmt.Sprintf("control-%03d.ts", index), text+"\nexport {};\n"))
	}
	paths = append(
		paths,
		harness.write("native-globals.d.ts", "declare const console: {log():void};\n"),
		harness.write("native-console.ts", "const detached=console.log;\nexport {};\n"),
	)
	return harness.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}

func volumeProfileHarness(test *testing.T) *harness {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return &harness{t: test, repository: repository, directory: test.TempDir()}
}

// Go products retain the original build commands until GoBuild is available.
func volumeProfileStage0(harness *harness) string {
	return harness.stage0()
}

// typeAwareCompilerFiles are the sources a volume product's compilers are built from: stage 0 (./cmd/adamic) and the
// checker archive (./bridge/tsgo/archive), each module file that pins what they import from outside the tree, and the
// rule packages cohere's oracle reads.
var typeAwareCompilerFiles = []string{
	"go.mod",
	"go.work",
	"cmd/adamic",
	"internal",
	"bridge/tsgo",
	"cohere/go.mod",
	"cohere/go.sum",
	"cohere/internal",
	"cohere/policy",
	"cohere/rule_runner",
	"cohere/TypeScript/tsc/internal",
	"cohere/TypeScript/tsc/go.mod",
	"cohere/TypeScript/tsc/go.sum",
	"cohere/TypeScript-shim",
}

// typeAwareTrackedTypeScript lists the .ts files the tree carries under each root, relative to repository, in roots'
// order. The listing is key-time work (buildcache.Keying): every file it lists goes into a key's Files, so a .ts added
// or changed under a root moves that key by its bytes. Walked under Workshop's trace, it put every directory under
// stage1/cohere/lint into the read sets of the volume products and the type-symbol native, so deleting a doc file under
// a rule's evidence/ moved all five keys (#xfvshc6).
func typeAwareTrackedTypeScript(repository string, roots ...string) ([]string, error) {
	var files []string
	var err error
	buildcache.Keying(func() {
		for _, root := range roots {
			err = filepath.WalkDir(filepath.Join(
				repository,
				root,
			), func(path string, entry fs.DirEntry, pathError error) error {
				if pathError != nil {
					return pathError
				}
				// Only what the tree carries: lint's tests write stage1/cohere/lint/.generated/registry.ts into a
				// checkout, which a runner's source never has, and keyed it on Workshop alone (Planner's diff of
				// landable-7's tree).
				if !buildcache.Tracked(repository, path, entry.IsDir()) {
					if entry.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
					relative, relativePathError := filepath.Rel(repository, path)
					if relativePathError != nil {
						return relativePathError
					}
					files = append(files, filepath.ToSlash(relative))
				}
				return nil
			})
			if err != nil {
				return
			}
		}
	})
	return files, err
}

// typeAwareCompilerKey keys a volume product by its compilers' sources, the TypeScript it compiles, the Go release and
// the C compiler the archive's cgo runs. It asks go nothing: go env and go list answer for the machine and checkout
// they run in, and a runner's answers keyed Workshop's products apart (#nm31pcn: typeaware_volume_lowered, and the
// corpus's stage 0, archives and oracle, on landable-6's tree a221706a).
func typeAwareCompilerKey(harness *harness, inputs *buildcache.Inputs, sanitize bool) {
	harness.t.Helper()
	inputs.Files = append(inputs.Files, typeAwareCompilerFiles...)
	files, typeAwareTrackedTypeScriptError := typeAwareTrackedTypeScript(
		harness.repository,
		"stage1/cohere/typeaware",
		"stage1/cohere/lint",
		"stage1/typescript",
	)
	if typeAwareTrackedTypeScriptError != nil {
		harness.t.Fatal(typeAwareTrackedTypeScriptError)
	}
	inputs.Files = append(inputs.Files, files...)
	// cgo's C compiler: the CC toolchain.Env runs the archive's go build with (clang where the caller sets none, as on
	// every Loom box), and clang for the sanitized archive, which sets CC itself. It reads no go: Env is this
	// process's.
	compiler := typeAwareBuildSetting("CC")
	if sanitize {
		compiler = "clang"
	}
	inputs.Flags = append(inputs.Flags, "CC="+compiler)
	inputs.Toolchain = []string{
		runtime.Version(),
		buildcache.Tool("clang", "--version"),
		typeAwareCompilerReport(compiler),
	}
}

func volumeProfileNativeInputs(
	harness *harness,
	stage0, archive string,
	sanitize bool,
	name string,
	variant ...string,
) buildcache.Inputs {
	harness.t.Helper()
	archiveName, archiveFlags := "typeaware checker archive", ""
	if sanitize {
		archiveName += " asan"
		archiveFlags = "CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"
	}
	reproducible := strings.Join(buildcache.ReproducibleGoFlags(), " ")
	inputs := buildcache.Inputs{
		Name: name,
		Flags: []string{
			"build",
			"stage1/cohere/typeaware/volume_suite.ts",
			"--tsgo",
			"go build " + reproducible + " ./cmd/adamic",
			"archive=" + archiveName,
			"go build " + reproducible + " -buildmode=c-archive ./bridge/tsgo/archive",
			archiveFlags,
			fmt.Sprintf("sanitize=%t", sanitize),
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		},
	}
	typeAwareCompilerKey(harness, &inputs, sanitize)
	inputs.Flags = append(inputs.Flags, variant...)
	return inputs
}

func volumeProfileNative(
	h *harness,
	stage0, archive string,
	sanitize bool,
	name string,
	build func(*harness) string,
	variant ...string,
) string {
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
func volumeProfileControlsLowered(harness *harness, stage0 string) string {
	inputs := volumeProfileNativeInputs(harness, stage0, "", false, "typeaware volume lowered")
	directory := buildcache.Product(harness.t, inputs, func(directory string) error {
		program, loadError := load.Load(
			[]string{filepath.Join(harness.repository, "stage1/cohere/typeaware/volume_suite.ts")},
		)
		if loadError != nil {
			return loadError
		}
		program.EnableTsGo()
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		source, tsGoCError := native.TsGoC(ir)
		if tsGoCError != nil {
			return tsGoCError
		}
		return os.WriteFile(filepath.Join(directory, "volume.c"), []byte(source), 0600)
	})
	return filepath.Join(directory, "volume.c")
}

func TestVolumeProfileControlsLower(test *testing.T) {
	test.Parallel()
	harness := volumeProfileHarness(test)
	volumeProfileControlsLowered(harness, volumeProfileStage0(harness))
}

type volumeProfileControlCase struct {
	name     string
	sanitize bool
}

func volumeProfileControlsCases() []volumeProfileControlCase {
	return []volumeProfileControlCase{{"controls", false}, {"controls-asan", true}}
}

func TestVolumeProfileControlsUnion(test *testing.T) {
	test.Parallel()
	cases := volumeProfileControlsCases()
	if len(cases) != testVolumeProfileControlsShards {
		test.Fatalf("controls enumeration: %d != %d", len(cases), testVolumeProfileControlsShards)
	}
	seen := map[string]int{}
	var keys []string
	for index := range volumeProfileControlSources() {
		keys = append(keys, fmt.Sprintf("control-%03d.ts", index))
	}
	keys = append(keys, "native-globals.d.ts", "native-console.ts")
	for _, controlCase := range cases {
		for _, key := range keys {
			seen[fmt.Sprintf("%s/asan=%t", key, controlCase.sanitize)]++
		}
	}
	for _, key := range keys {
		for _, mode := range []bool{false, true} {
			if count := seen[fmt.Sprintf("%s/asan=%t", key, mode)]; count != 1 {
				test.Fatalf("control union %s asan=%t: %d owners", key, mode, count)
			}
		}
	}
	test.Logf("controls union: %d live source/mode cases in %d shards", len(seen), len(cases))
}

func volumeProfileControlsBinary(h *harness, stage0, archive string, sanitize bool) string {
	test := h.t
	name := "typeaware volume"
	if sanitize {
		name += " asan"
	}
	lowered := volumeProfileControlsLowered(h, stage0)
	binary := volumeProfileNative(h, stage0, archive, sanitize, name, func(local *harness) string {
		source, readError := os.ReadFile(lowered)
		if readError != nil {
			test.Fatal(readError)
		}
		path := filepath.Join(local.directory, "volume")
		if typeAwareNativeBuildError := typeAwareNativeBuild(
			string(source),
			path,
			archive,
			sanitize,
		); typeAwareNativeBuildError != nil {
			test.Fatal(typeAwareNativeBuildError)
		}
		return path
	})
	return binary
}

func runVolumeProfileControls(test *testing.T, index int) {
	started := time.Now()
	controlCase := volumeProfileControlsCases()[index]
	h := volumeProfileHarness(test)
	var stage0, archive, oracle string
	var builds sync.WaitGroup
	for _, job := range []struct {
		product *string
		build   func(*harness) string
	}{
		{&stage0, volumeProfileStage0},
		{&archive, func(harness *harness) string { return harness.archive(controlCase.sanitize) }},
		{&oracle, func(harness *harness) string { return volumeOracle(harness, "volume-oracle", "oracle_volume.go") }},
	} {
		local := volumeProfileHarness(test)
		builds.Add(1)
		go func() { defer builds.Done(); *job.product = job.build(local) }()
	}
	builds.Wait()
	if test.Failed() {
		return
	}
	test.Logf("setup: %.3fs", time.Since(started).Seconds())
	binary := volumeProfileControlsBinary(h, stage0, archive, controlCase.sanitize)
	checkStarted := time.Now()
	h.compare(
		controlCase.name,
		oracle,
		binary,
		filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/tsconfig.json"),
		volumeProfileControlsManifest(h),
	)
	test.Logf("check after product fetch: %.3fs; cooked=false", time.Since(checkStarted).Seconds())
}

func TestVolumeProfileControls_000(test *testing.T) {
	test.Parallel()
	runVolumeProfileControls(test, 0)
}
func TestVolumeProfileControls_001(test *testing.T) {
	test.Parallel()
	runVolumeProfileControls(test, 1)
}

const testShadowIndexMissingBindingShards = 1

func TestShadowIndexMissingBindingUnion(test *testing.T) {
	test.Parallel()
	cases := []string{"missing-binding"}
	if len(cases) != testShadowIndexMissingBindingShards {
		test.Fatalf("missing-binding enumeration: %d != %d", len(cases), testShadowIndexMissingBindingShards)
	}
	seen := map[string]int{}
	for _, name := range cases {
		seen[name]++
	}
	for name, count := range seen {
		if count != 1 {
			test.Fatalf("%s: %d owners", name, count)
		}
	}
	test.Logf("missing-binding union: %d cases", len(cases))
}
