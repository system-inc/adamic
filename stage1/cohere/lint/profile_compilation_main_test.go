package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testProfileCompilationShards = 1

func compilationCase(t *testing.T) string {
	t.Helper()
	paths, err := registry.Witnesses("rules/no-var")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("profile witness corpus is empty")
	}
	return filepath.ToSlash(filepath.Join("stage1/cohere/lint", paths[0])) + "\tno-var"
}

func compilationShard(key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % testProfileCompilationShards)
}

func TestProfileCompilationUnion(t *testing.T) {
	t.Parallel()
	cases := []string{compilationCase(t)}
	assignments := make([][]string, testProfileCompilationShards)
	for _, key := range cases {
		assignments[compilationShard(key)] = append(assignments[compilationShard(key)], key)
	}
	// This table is the top-level functions go test -list exposes.
	enumerated := []func(*testing.T){TestProfileCompilation_000}
	if len(enumerated) != testProfileCompilationShards {
		t.Fatal("top-level shard count differs from const")
	}
	want := map[string]bool{}
	seen := map[string]bool{}
	total := 0
	for _, key := range cases {
		if want[key] {
			t.Fatal("duplicate corpus case", key)
		}
		want[key] = true
	}
	for _, part := range assignments {
		for _, key := range part {
			if !want[key] || seen[key] {
				t.Fatal("unknown or repeated case", key)
			}
			seen[key] = true
			total++
		}
	}
	if total != len(cases) || len(seen) != len(want) {
		t.Fatal("missing corpus case")
	}
	t.Logf("union: %d live cases, each exactly once", total)
}

func compilationSelected(t *testing.T) {
	t.Helper()
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" && value != "0/1" {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q for one fixed shard", value)
	}
}

func compilationBudget(t *testing.T) func() {
	started := time.Now()
	return func() {
		elapsed := time.Since(started)
		cooked := elapsed > 60*time.Second
		t.Logf("wall=%s cooked=%t", elapsed, cooked)
		if cooked {
			t.Errorf("cooked: exceeded 60s budget")
		}
	}
}

// These are inputs to the non-Go lowering pipeline, not a hand-keyed Go build.
// go list enumerates the compiler implementation, embeds and module pins.
func compilationInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(context.Background(), "go", "list", "-deps", "-json", "./internal/load", "./internal/lower", "./internal/native", "./internal/javascript", "./stage1/cohere/lint/registry")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	files := map[string]bool{}
	add := func(path string) {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(rel, "../") {
			files[filepath.ToSlash(rel)] = true
		}
	}
	for {
		var p struct {
			Dir                                                             string
			GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, EmbedFiles []string
			Module                                                          *struct{ GoMod string }
		}
		err := decoder.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, names := range [][]string{p.GoFiles, p.CgoFiles, p.CFiles, p.CXXFiles, p.HFiles, p.SFiles, p.EmbedFiles} {
			for _, name := range names {
				add(filepath.Join(p.Dir, name))
			}
		}
		if p.Module != nil && p.Module.GoMod != "" {
			add(p.Module.GoMod)
			sum := filepath.Join(filepath.Dir(p.Module.GoMod), "go.sum")
			if _, err := os.Stat(sum); err == nil {
				add(sum)
			}
		}
	}
	for _, path := range portFiles(t) {
		files[filepath.ToSlash(filepath.Join("stage1/cohere/lint", path))] = true
	}
	for _, path := range []string{"go.work", "stage1/cohere/lint/profile_compilation_main_test.go", "stage1/cohere/lint/profile_compilation_ir_test.go", "stage1/cohere/lint/profile_test.go", "stage1/cohere/lint/lint_setup_clock_regression_test.go"} {
		files[path] = true
	}
	var paths []string
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return buildcache.Inputs{Name: "lint-profile-lowered", Files: paths, Flags: []string{"source-root=" + packageDirectory}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
}

func compilationProducts(t *testing.T) string {
	return compilationProductKinds(t, []string{"scanner", "counted", "profiled"})
}

func compilationProductKinds(t *testing.T, kinds []string) string {
	t.Helper()
	inputs := compilationInputs(t)
	cdir := compilationEmission(t, inputs, "c")
	jsdir := compilationEmission(t, inputs, "javascript")
	source, err := os.ReadFile(filepath.Join(cdir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("lowered-sha256=%x", sha256.Sum256(source)))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := t.TempDir()
	var group sync.WaitGroup
	for _, kind := range kinds {
		kind := kind
		variant := inputs
		variant.Name = "lint-profile-" + kind
		options := native.Options{Count: kind == "counted", Jobs: 1}
		variant.Flags = append(append([]string{}, inputs.Flags...), native.Flags(options)...)
		if kind == "profiled" {
			variant.Flags = append(variant.Flags, "-g", "-lm")
		}
		group.Add(1)
		go func() {
			defer group.Done()
			product := buildcache.Product(t, variant, func(dir string) error {
				if kind != "profiled" {
					return native.Build(string(source), filepath.Join(dir, kind), options)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.c"), source, 0644); err != nil {
					return err
				}
				entries, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
				if err != nil {
					return err
				}
				units := []string{filepath.Join(dir, "main.c")}
				for _, entry := range entries {
					if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
						continue
					}
					data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", entry.Name()))
					if err != nil {
						return err
					}
					path := filepath.Join(dir, entry.Name())
					if err := os.WriteFile(path, data, 0644); err != nil {
						return err
					}
					if strings.HasSuffix(path, ".c") {
						units = append(units, path)
					}
				}
				flags := append(native.Flags(options), "-g", "-o", filepath.Join(dir, kind))
				flags = append(flags, units...)
				flags = append(flags, "-lm")
				command := exec.CommandContext(context.Background(), "clang", flags...)
				if output, err := command.CombinedOutput(); err != nil {
					return fmt.Errorf("profile clang: %w\n%s", err, output)
				}
				return nil
			})
			data, err := os.ReadFile(filepath.Join(product, kind))
			if err == nil {
				err = os.WriteFile(filepath.Join(directory, kind), data, 0755)
			}
			if err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if t.Failed() {
		t.FailNow()
	}
	module, err := os.ReadFile(filepath.Join(jsdir, "lint.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "lint.mjs"), module, 0644); err != nil {
		t.Fatal(err)
	}
	return directory
}

func compilationOracleOutputs(t *testing.T, oracle, manifestPath string) ([]byte, []byte) {
	t.Helper()
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	sourcePath := strings.SplitN(string(text), "\t", 2)[0]
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	witness := strings.SplitN(compilationCase(t), "\t", 2)[0]
	inputs := buildcache.Inputs{Name: "lint-profile-oracle-outputs", Files: []string{witness, "stage1/cohere/lint/profile_compilation_main_test.go"}, Flags: []string{"mode=no-var", fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)), fmt.Sprintf("case-sha256=%x", sha256.Sum256(source))}, Toolchain: []string{runtime.GOOS, runtime.GOARCH}}
	product := buildcache.Product(t, inputs, func(dir string) error {
		for _, mode := range []string{"wire", "count"} {
			args := []string{"--manifest", manifestPath}
			if mode == "count" {
				args = append(args, "--count")
			}
			data := execute(t, "", oracle, args...).output
			data = bytes.ReplaceAll(data, []byte(sourcePath), []byte("@profile-case@"))
			if err := os.WriteFile(filepath.Join(dir, mode), data, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	read := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(product, name))
		if err != nil {
			t.Fatal(err)
		}
		return bytes.ReplaceAll(data, []byte("@profile-case@"), []byte(sourcePath))
	}
	return read("wire"), read("count")
}

func compilationLowered(t *testing.T, inputs buildcache.Inputs) string {
	return buildcache.Product(t, inputs, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(packageDirectory, "main.ts")})
		if err != nil {
			return err
		}
		value, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		file, err := os.Create(filepath.Join(dir, "program.gob"))
		if err != nil {
			return err
		}
		defer file.Close()
		return gob.NewEncoder(file).Encode(value)
	})
}

func compilationEmission(t *testing.T, inputs buildcache.Inputs, kind string) string {
	lowered := compilationLowered(t, inputs)
	data, err := os.ReadFile(filepath.Join(lowered, "program.gob"))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "lint-profile-emit-" + kind
	inputs.Flags = append(append([]string{}, inputs.Flags...), fmt.Sprintf("ir-sha256=%x", sha256.Sum256(data)))
	return buildcache.Product(t, inputs, func(dir string) error {
		var value ir.Program
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&value); err != nil {
			return err
		}
		if kind == "c" {
			return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(&value)), 0644)
		}
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(&value)), 0644)
	})
}

// Build products are individually enumerable so cold work never waits behind
// a monolithic lowering-plus-emission build. The build phase owns their ceiling.

// Bound child commands without requiring an external timeout executable. Kill
// the entire process group so compiler descendants cannot outlive the deadline.
func compilationCommand(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

// Setup has no test-side clock; runtime shards start their clock after preparation.

func compilationPrepare(t *testing.T) (string, string, string, []byte, []byte) {
	t.Helper()
	compilationCase(t)
	directory := t.TempDir()
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	products := compilationProducts(t)
	path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
	oracle := compilationOracle(t)
	want, countedWant := compilationOracleOutputs(t, oracle, path)
	return directory, products, path, want, countedWant
}

func compilationOracle(t *testing.T) string {
	t.Helper()
	inputs := testShardsAgreeInputs(t, "profile-oracle")
	inputs.Files = append(inputs.Files, "stage1/cohere/lint/profile_compilation_main_test.go")
	inputs.Flags = append(inputs.Flags, "overlay=full-live-registry")
	product := buildcache.Product(t, inputs, func(dir string) error { _, err := goOracleIn(packageDirectory, dir); return err })
	return filepath.Join(product, "oracle")
}
