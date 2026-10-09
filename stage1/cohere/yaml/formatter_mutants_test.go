package yaml

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testFormatterMutantsShards = 6

var yamlFormatterMutants = []struct{ name, file, from, to string }{
	{"root final newline lost", "printer.ts", "const hard = !(", "const hard = false && !("},
	{"colon separation lost", "printer.ts", "this.layout.text(': '), printedValue", "this.layout.text(':'), printedValue"},
	{"flow trailing comma lost", "printer.ts", "this.layout.ifBreak(this.layout.text(','), this.empty, -1)", "this.layout.ifBreak(this.layout.text(''), this.empty, -1)"},
	{"batch backslash scan misses escapes", "main.ts", "text.charCodeAt(index) !== 92", "text.charCodeAt(index) !== 13"},
	{"emoji first unit range loses endpoint", "width.ts", "code <= last", "code < last"},
	{"emoji surrogate slot shifted", "width.ts", "code - 0x100000 + 0xd800", "code - 0x100000 + 0xd900"},
}

// Each fixed top-level unit checks one mutant against every live corpus case.
// ADAMIC_TEST_SHARD=i/n selects unit indices modulo n; unset runs all.
// Inputs are cohere at 7945d102a6c18dd36adf9114a758ce646e8b2359 and
// TypeScript-Go at d92d9bfee114c80be2c375d72edae966176e3a4f, enforced by
// corpusfiles.Upstream, plus deterministic generated cases. No mutable
// repository corpus is scanned. Adding cases cannot change the shard count.
func TestFormatterMutants_000(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 0) }
func TestFormatterMutants_001(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 1) }
func TestFormatterMutants_002(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 2) }
func TestFormatterMutants_003(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 3) }
func TestFormatterMutants_004(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 4) }
func TestFormatterMutants_005(t *testing.T) { t.Parallel(); yamlRunFormatterMutant(t, 5) }

func yamlFormatterMutantPlan(t *testing.T, count int) [][]string {
	t.Helper()
	if len(yamlFormatterMutants) != testFormatterMutantsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(yamlFormatterMutants), testFormatterMutantsShards)
	}
	plan := make([][]string, testFormatterMutantsShards)
	for shard, mutant := range yamlFormatterMutants {
		for index := 0; index < count; index++ {
			plan[shard] = append(plan[shard], fmt.Sprintf("%s/case-%05d", mutant.name, index))
		}
	}
	return plan
}

func TestFormatterMutantsUnion(t *testing.T) {
	t.Parallel()
	_, _, count := formatCases(t)
	if count == 0 {
		t.Fatal("empty corpus")
	}
	patterns := []string{"*.yaml", "*.yml"}
	cohere := corpusfiles.Upstream(t, filepath.Join(repository, "cohere"), corpusfiles.CohereCommit, []string{".github/workflows"}, patterns)
	ts := corpusfiles.Upstream(t, filepath.Join(repository, "cohere/TypeScript"), corpusfiles.TypeScriptGoCommit, []string{".custom-gcl.yml", ".github", ".golangci.yml", "tools/pipelines"}, patterns)
	if len(cohere) != 1 || len(ts) != 35 {
		t.Fatalf("pinned file counts: cohere=%d want 1, TypeScript-Go=%d want 35", len(cohere), len(ts))
	}
	wanted := make(map[string]bool)
	for _, mutant := range yamlFormatterMutants {
		for index := 0; index < count; index++ {
			id := fmt.Sprintf("%s/case-%05d", mutant.name, index)
			if wanted[id] {
				t.Fatalf("duplicate enumeration %s", id)
			}
			wanted[id] = true
		}
	}
	seen := make(map[string]bool)
	for shard, ids := range yamlFormatterMutantPlan(t, count) {
		for _, id := range ids {
			if !wanted[id] || seen[id] {
				t.Fatalf("TestFormatterMutants_%03d: unexpected or repeated %s", shard, id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(wanted) {
		t.Fatalf("union: got %d cases, enumerated %d", len(seen), len(wanted))
	}
	for id := range wanted {
		if !seen[id] {
			t.Fatalf("missing %s", id)
		}
	}
	t.Logf("union: %d unique mutant/case ids, %d mutants x %d live cases", len(seen), testFormatterMutantsShards, count)
}

func yamlFormatterMutantSurvivor(unit, side string, actual, expected []byte) error {
	if bytes.Equal(actual, expected) {
		return fmt.Errorf("%s: %s missed mutant", unit, side)
	}
	return nil
}

func TestFormatterMutantsPlantedSurvivor(t *testing.T) {
	t.Parallel()
	for _, side := range []string{"native", "Node"} {
		caught := 0
		for index := range yamlFormatterMutants {
			unit := fmt.Sprintf("TestFormatterMutants_%03d", index)
			actual := []byte("ok\tmutated\n")
			if index == 3 {
				actual = []byte("ok\toriginal\n")
			}
			err := yamlFormatterMutantSurvivor(unit, side, actual, []byte("ok\toriginal\n"))
			if err != nil {
				caught++
				if index != 3 || !strings.Contains(err.Error(), unit) {
					t.Fatalf("wrong shard caught survivor: %v", err)
				}
				t.Logf("planted survivor case-0 caught by %s (%s)", unit, side)
			}
		}
		if caught != 1 {
			t.Fatalf("%s: survivor caught by %d shards, want 1", side, caught)
		}
	}
}

func yamlRunFormatterMutant(t *testing.T, shard int) {
	t.Helper()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			t.Errorf("cooked: unit took %s, budget 60s", elapsed)
		}
	}()
	if selector := os.Getenv("ADAMIC_TEST_SHARD"); selector != "" {
		parts := strings.Split(selector, "/")
		if len(parts) != 2 {
			t.Fatal("ADAMIC_TEST_SHARD must be i/n")
		}
		i, errI := strconv.Atoi(parts[0])
		n, errN := strconv.Atoi(parts[1])
		if errI != nil || errN != nil || n <= 0 || i < 0 || i >= n {
			t.Fatal("ADAMIC_TEST_SHARD must be zero-based i/n")
		}
		if shard%n != i {
			t.Skip("unit excluded by ADAMIC_TEST_SHARD")
		}
	}
	cases, _, count := formatCases(t)
	plan := yamlFormatterMutantPlan(t, count)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	if rows := len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")); rows != len(plan[shard]) {
		t.Fatalf("input has %d cases, plan has %d", rows, len(plan[shard]))
	}
	expected := goFormat(t, cases) // The Go oracle uses -overlay; buildcache refuses overlays.
	mutant := yamlFormatterMutants[shard]
	entry, binary := yamlFormatterMutantProduct(t, mutant.name, mutant.file, mutant.from, mutant.to)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content: mutant %q, all %d live corpus cases", mutant.name, count)
	nativeOut := run(t, "", nil, binary, "--cases", cases)
	nodeOut := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native", nativeOut}, {"Node", nodeOut}} {
		if err := yamlFormatterMutantSurvivor(t.Name(), side.name, side.out, expected); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
	}
}

// A product contains immutable sources, lowered C and the native binary. Keys
// cover original repository inputs and the mutation, never temporary paths.
func yamlFormatterMutantProduct(t *testing.T, name, file, from, to string) (string, string) {
	t.Helper()
	entries, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"internal", "bridge/tsgo", "cohere", "go.mod", "stage1/cohere/yaml/formatter_mutants_test.go"}
	for _, entry := range entries {
		inputs = append(inputs, "stage1/cohere/yaml/"+entry)
	}
	options := native.Options{}
	flags := append([]string{file, from, to, "main.ts", "C", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, native.Flags(options)...)
	product := buildcache.Product(t, buildcache.Inputs{
		Name: "yaml-formatter-mutant-" + name, Files: inputs, Flags: flags,
		Toolchain: []string{runtime.Version() + "/" + runtime.GOOS + "/" + runtime.GOARCH, buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		for _, entry := range entries {
			source, err := os.ReadFile(entry)
			if err != nil {
				return err
			}
			if entry == file {
				if strings.Count(string(source), from) != 1 {
					return fmt.Errorf("mutation site must occur once")
				}
				source = []byte(strings.Replace(string(source), from, to, 1))
			}
			if err := os.WriteFile(filepath.Join(directory, entry), source, 0644); err != nil {
				return err
			}
		}
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		source := native.C(ir)
		if err := os.WriteFile(filepath.Join(directory, "mutant.c"), []byte(source), 0644); err != nil {
			return err
		}
		return native.Build(source, filepath.Join(directory, "mutant"), options)
	})
	return filepath.Join(product, "main.ts"), filepath.Join(product, "mutant")
}
