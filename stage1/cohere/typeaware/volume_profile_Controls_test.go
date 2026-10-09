package typeaware

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	fingerprint := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		return fmt.Sprintf("%x", sha256.Sum256(data))
	}
	var files []string
	for _, directory := range []string{"stage1/cohere/typeaware", "stage1/cohere/lint", "stage1/typescript"} {
		err := filepath.WalkDir(filepath.Join(h.repository, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && filepath.Ext(path) == ".ts" {
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
	inputs := buildcache.Inputs{Name: name, Files: files,
		Flags:     []string{"build", "stage1/cohere/typeaware/volume_suite.ts", "--tsgo", "stage0=" + fingerprint(stage0), "archive=" + fingerprint(archive), fmt.Sprintf("sanitize=%t", sanitize), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")},
		Toolchain: []string{buildcache.Tool("clang", "--version")}}
	inputs.Flags = append(inputs.Flags, variant...)
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		local := &harness{t: h.t, repository: h.repository, directory: directory}
		product := build(local)
		return os.Rename(product, filepath.Join(directory, "volume"))
	})
	return filepath.Join(directory, "volume")
}

func TestVolumeProfileControls(t *testing.T) {
	started := time.Now()
	cases := []struct {
		name     string
		sanitize bool
	}{{"controls", false}, {"controls-asan", true}}
	if len(cases) != testVolumeProfileControlsShards {
		t.Fatalf("controls enumeration: %d != %d", len(cases), testVolumeProfileControlsShards)
	}
	// Count every live source/mode occurrence in the union. No fixture total is
	// frozen: adding a control keeps both mode shards and all existing ownership.
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

	// Build shared Go dependencies before releasing parallel checks.
	h := volumeProfileHarness(t)
	stage0 := volumeProfileStage0(h)
	oracle := volumeOracle(h, "volume-oracle", "oracle_volume.go")
	normal := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	t.Logf("TestVolumeProfileControls (setup): %.3fs", time.Since(started).Seconds())
	for i, c := range cases {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			h := volumeProfileHarness(t)
			archive, name := normal, "typeaware volume"
			if c.sanitize {
				archive, name = sanitized, "typeaware volume asan"
			}
			binary := volumeProfileNative(h, stage0, archive, c.sanitize, name, func(local *harness) string {
				return local.build(stage0, "volume", filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), archive, c.sanitize)
			})
			checkStarted := time.Now()
			h.compare(c.name, oracle, binary, filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/tsconfig.json"), volumeProfileControlsManifest(h))
			t.Logf("check excluding cached build: %.3fs", time.Since(checkStarted).Seconds())
		})
	}
}
