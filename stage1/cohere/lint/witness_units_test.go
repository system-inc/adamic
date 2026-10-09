package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// No package cache: until internal/buildcache lands, each maker runs once in the
// parent and every unit receives its immutable products. Keep inputs beside the
// func(dir string) error makers so they can use buildcache.Product when available.
type unitBuildInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

func unitProduct(t *testing.T, inputs unitBuildInputs, makeProduct func(dir string) error) string {
	t.Helper()
	dir := t.TempDir()
	started := time.Now()
	err := makeProduct(dir)
	elapsed := time.Since(started)
	t.Logf("build product %s: %s (files=%d flags=%q toolchain=%s)", inputs.Name, elapsed, len(inputs.Files), inputs.Flags, inputs.Toolchain)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type witnessUnitProducts struct {
	Directory, Source, Oracle, Native, JavaScript string
}

// Cold measurement isolates the Go compiler cache as well as disabling the native
// object cache in the invoking command. Normal runs reuse compiler caches.
func unitGoEnvironment(dir string) []string {
	if os.Getenv("ADAMIC_TEST_COLD_PRODUCTS") == "1" {
		return []string{"GOCACHE=" + filepath.Join(dir, "go-cache")}
	}
	return nil
}

func unitSourceFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	var files []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			if !entry.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func witnessProducts(t *testing.T) witnessUnitProducts {
	t.Helper()
	if path := os.Getenv("ADAMIC_LINT_WITNESS_PRODUCTS"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var p witnessUnitProducts
		if err := json.Unmarshal(data, &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	directory := mutant(t, "", "")
	witness := filepath.Join(directory, "rules/no-debugger/testdata/witness.ts.txt")
	renamed := strings.TrimSuffix(witness, ".ts.txt") + ".tsx.txt"
	if err := os.Rename(witness, renamed); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(renamed, []byte("const node = 1; debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	source := ownedWitnesses(t, directory, "no-debugger")[0]
	prepareRegistry(t, directory)
	files := unitSourceFiles(t, directory, filepath.Join(repository, "internal"), filepath.Join(repository, "bridge"), filepath.Join(repository, "cohere"), filepath.Join(repository, "go.mod"), filepath.Join(repository, "go.work"))
	goVersion, err := exec.Command("go", "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	clangVersion, err := exec.Command("clang", "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	goToolchain := strings.TrimSpace(string(goVersion))
	clangToolchain := strings.Split(string(clangVersion), "\n")[0]
	oracleInputs := unitBuildInputs{Name: "witness-go-oracle", Files: append(append([]string(nil), files...), filepath.Join(packageDirectory, "testdata/oracle.go")), Flags: []string{"build", "-overlay"}, Toolchain: goToolchain}
	oracle := unitProduct(t, oracleInputs, func(dir string) error {
		_, err := goOracleIn(directory, dir, unitGoEnvironment(dir)...)
		return err
	})
	bridge := false
	loweringInputs := unitBuildInputs{Name: "witness-lowered-C-and-JavaScript", Files: files, Flags: []string{"EnableTSGo"}, Toolchain: goToolchain}
	lowered := unitProduct(t, loweringInputs, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		program.EnableTSGo()
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		bridge = native.UsesTSGo(lowered)
		c := ""
		if bridge {
			c, err = native.TSGoC(lowered)
		} else {
			c = native.C(lowered)
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(c), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	})
	archive := ""
	if bridge {
		archiveInputs := unitBuildInputs{Name: "witness-sanitized-checker-archive", Files: files, Flags: []string{"-buildmode=c-archive", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"}, Toolchain: goToolchain + "; " + clangToolchain}
		dir := unitProduct(t, archiveInputs, func(dir string) error {
			root, err := filepath.Abs(repository)
			if err != nil {
				return err
			}
			_, err = run(root, append(unitGoEnvironment(dir), "GOMAXPROCS=4", "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"), "go", "build", "-buildmode=c-archive", "-o", filepath.Join(dir, "checker.a"), "./bridge/tsgo/archive")
			return err
		})
		archive = filepath.Join(dir, "checker.a")
	}
	nativeInputs := unitBuildInputs{Name: "witness-sanitized-native", Files: append(unitSourceFiles(t, filepath.Join(repository, "internal/native")), filepath.Join(lowered, "lint.c"), archive), Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: clangToolchain}
	binary := unitProduct(t, nativeInputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		output := filepath.Join(dir, "scanner")
		return nativeBuild(func() error {
			if archive != "" {
				return buildCheckerWithRuntime(string(data), output, archive, native.Options{Sanitize: true})
			}
			return native.Build(string(data), output, native.Options{Sanitize: true})
		})
	})
	return witnessUnitProducts{Directory: directory, Source: source, Oracle: filepath.Join(oracle, "oracle"), Native: filepath.Join(binary, "scanner"), JavaScript: filepath.Join(lowered, "lint.mjs")}
}

func checkedCaseShards(t *testing.T, ids []string, count int) [][]string {
	t.Helper()
	if count < 1 {
		t.Fatal("invalid shard count")
	}
	shards := make([][]string, count)
	expected := map[string]bool{}
	for i, id := range ids {
		if expected[id] {
			t.Fatalf("repeated unsplit case %q", id)
		}
		expected[id] = true
		shards[i%count] = append(shards[i%count], id)
	}
	seen := map[string]bool{}
	total := 0
	for _, shard := range shards {
		for _, id := range shard {
			if !expected[id] || seen[id] {
				t.Fatalf("missing or repeated shard case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != len(ids) || len(seen) != len(expected) {
		t.Fatalf("shard union: got %d cases, want %d", total, len(ids))
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("missing shard case %q", id)
		}
	}
	t.Logf("shard union: %d cases, %d unique ids, %d shards", total, len(seen), count)
	return shards
}

func selectedCaseShard(t *testing.T, index int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatalf("ADAMIC_TEST_SHARD must be zero-based i/n, got %q", value)
	}
	i, errI := strconv.Atoi(parts[0])
	n, errN := strconv.Atoi(parts[1])
	if errI != nil || errN != nil || n < 1 || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index%n == i
}

func checkWitnessPlantedDisagreement(t *testing.T, p witnessUnitProducts) {
	t.Helper()
	data, err := os.ReadFile(p.JavaScript)
	if err != nil {
		t.Fatal(err)
	}
	p.JavaScript = filepath.Join(t.TempDir(), "planted.mjs")
	if err := os.WriteFile(p.JavaScript, append(data, []byte("\nconsole.log('planted witness disagreement');\n")...), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	products := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(products, data, 0644); err != nil {
		t.Fatal(err)
	}
	for seat := 0; seat < 2; seat++ {
		command := exec.Command(os.Args[0], "-test.run=^TestWitnessScriptKind$/^shard-000$", "-test.v", "-test.timeout=30s")
		command.Env = append(os.Environ(), "ADAMIC_LINT_WITNESS_PRODUCTS="+products, fmt.Sprintf("ADAMIC_TEST_SHARD=%d/2", seat))
		output, err := command.CombinedOutput()
		text := string(output)
		if seat == 0 {
			if err == nil || strings.Count(text, "--- FAIL: TestWitnessScriptKind/shard-000") != 1 || !strings.Contains(text, "emitted JavaScript:") || !strings.Contains(text, "planted witness disagreement") {
				t.Fatalf("wrong planted failure: %v\n%s", err, text)
			}
			t.Log("planted disagreement caught exactly by shard-000 on seat 0/2")
		} else if err != nil || strings.Contains(text, "--- FAIL:") || strings.Contains(text, "=== RUN   TestWitnessScriptKind/shard-") {
			t.Fatalf("non-owning seat ran or caught planted case: %v\n%s", err, text)
		}
	}
}
