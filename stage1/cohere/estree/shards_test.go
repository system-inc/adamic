package estree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These inputs describe the directory-writing build callbacks. There is no
// package cache: until internal/buildcache lands, each selected product is built
// once per invocation and shared by all of its parallel shards.
type buildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func prepareProduct(t *testing.T, inputs buildInputs, build func(string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatal(err)
	}
	t.Logf("cold build %s total: %.3fs; files=%v flags=%v toolchain=%s", inputs.Name, time.Since(start).Seconds(), inputs.Files, inputs.Flags, inputs.Toolchain)
	return dir
}

func portBuildInputs(path string, sanitize bool) buildInputs {
	return buildInputs{Name: filepath.Base(filepath.Dir(path)), Files: []string{filepath.Dir(path), "../../typescript", "../../../internal/load", "../../../internal/lower", "../../../internal/native", "../../../internal/javascript", "../../../cohere"}, Flags: []string{fmt.Sprintf("sanitize=%t", sanitize)}, Toolchain: "go+clang+node"}
}

const testThreePortMutantsShards = 12
const mutantCasesPerShard = 20

type portMutation struct{ name, file, from, to, witness string }

func threePortMutations() []portMutation {
	return []portMutation{
		{"member-computed", "convert.ts", "boolValue(node.kind === 'ElementAccessExpression')", "boolValue(node.kind === 'PropertyAccessExpression')", "a.b; a[b];"},
		{"logical-rebalance", "postprocess.ts", "completed.set(id, this.rebalance(id));", "completed.set(id, id);", "a && (b && (c && d));"},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*", "x; /**\n * one\n *//**\n * two\n */ y;"},
	}
}

type mutantCase struct {
	mutant, index int
	witness       bool
}
type mutantShard struct {
	name  string
	cases []mutantCase
}

func threePortShards(t *testing.T) []mutantShard {
	return mutantShardPlan(t, generated(), threePortMutations(), testThreePortMutantsShards)
}

func mutantShardPlan(t *testing.T, cases []string, mutations []portMutation, declared int) []mutantShard {
	t.Helper()
	var shards []mutantShard
	var unsplit []mutantCase
	for m, mutation := range mutations {
		witnesses := 0
		var pairs []mutantCase
		for i, text := range cases {
			pair := mutantCase{m, i, strings.Contains(text, mutation.witness)}
			if pair.witness {
				witnesses++
			}
			pairs = append(pairs, pair)
			unsplit = append(unsplit, pair)
		}
		if witnesses != 1 {
			t.Fatalf("%s: expected one witness, got %d", mutation.name, witnesses)
		}
		for start := 0; start < len(pairs); start += mutantCasesPerShard {
			end := min(start+mutantCasesPerShard, len(pairs))
			shards = append(shards, mutantShard{fmt.Sprintf("shard-%03d", len(shards)), pairs[start:end]})
		}
	}
	if len(shards) != declared {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), declared)
	}
	want := make(map[[2]int]bool)
	for _, pair := range unsplit {
		id := [2]int{pair.mutant, pair.index}
		if want[id] {
			t.Fatalf("duplicate unsplit case %v", id)
		}
		want[id] = true
	}
	seen := make(map[[2]int]bool)
	count := 0
	for _, shard := range shards {
		for _, pair := range shard.cases {
			id := [2]int{pair.mutant, pair.index}
			if !want[id] || seen[id] {
				t.Fatalf("%s: extra or repeated case %v", shard.name, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(unsplit) || len(seen) != len(want) {
		t.Fatalf("shard union: %d/%d cases, %d/%d ids", count, len(unsplit), len(seen), len(want))
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("shard union missing %v", id)
		}
	}
	t.Logf("shard union: %d pairs, %d distinct ids, %d mutants x %d cases", count, len(seen), len(mutations), len(cases))
	return shards
}

func selectedShard(t *testing.T, index int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD=%q: want zero-based i/n", value)
	}
	i, e1 := strconv.Atoi(fields[0])
	n, e2 := strconv.Atoi(fields[1])
	if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD=%q: want 0 <= i < n", value)
	}
	return index%n == i
}

// Canonical output ends every whole-file record with one escaped "stripped"
// line. Do not divide a file's AST, comments, or stripped-source checks.
func canonicalRecords(t *testing.T, data []byte, count int) [][]byte {
	t.Helper()
	var records [][]byte
	start, offset := 0, 0
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		offset += len(line)
		if bytes.HasPrefix(line, []byte("stripped ")) {
			records = append(records, data[start:offset])
			start = offset
		}
	}
	if start != len(data) || len(records) != count {
		t.Fatalf("canonical records: %d, want %d; trailing bytes %d", len(records), count, len(data)-start)
	}
	return records
}

func checkMutantCase(pair mutantCase, want, source, native []byte) error {
	sourceDifference := firstDifference(want, source)
	nativeDifference := firstDifference(want, native)
	if pair.witness {
		if sourceDifference == "" {
			return fmt.Errorf("case %03d: source Node mutant survived", pair.index)
		}
		if nativeDifference == "" {
			return fmt.Errorf("case %03d: sanitized native mutant survived", pair.index)
		}
	}
	if (sourceDifference == "") != (nativeDifference == "") {
		return fmt.Errorf("case %03d: mutant ports differ in agreement with Go", pair.index)
	}
	if diff := firstDifference(source, native); diff != "" {
		return fmt.Errorf("case %03d: mutant ports disagree: %s", pair.index, diff)
	}
	return nil
}

// TestThreePortMutants covers every generated case on both mutant ports, with
// ASan/UBSan and the existing leak/stderr checks on every shard. ADAMIC_TEST_SHARD
// i/n selects zero-based shard indices modulo n; unset runs all. The gate can
// independently select TestThreePortMutants/shard-NNN with -run.
func TestThreePortMutants(t *testing.T) {
	runMutantShards(t, generated(), threePortMutations(), threePortShards(t))
}

func runMutantShards(t *testing.T, cases []string, mutations []portMutation, shards []mutantShard) {
	t.Helper()
	start := time.Now()
	type product struct{ source, native string }
	products := make([]product, len(mutations))
	needed := make([]bool, len(mutations))
	oracle := ""
	setup := time.Duration(0)
	t.Cleanup(func() { t.Logf("setup excluding parallel shard logic: %.3fs", setup.Seconds()) })
	for index, shard := range shards {
		if !selectedShard(t, index) {
			continue
		}
		t.Run(shard.name, func(t *testing.T) {
			m := shard.cases[0].mutant
			needed[m] = true // Parent prepares only products used by -run-selected shards.
			t.Parallel()
			texts := make([]string, len(shard.cases))
			for i, pair := range shard.cases {
				texts[i] = cases[pair.index]
			}
			list := manifest(t, texts)
			want := canonicalRecords(t, execute(t, "", oracle, "--manifest", list), len(texts))
			source := canonicalRecords(t, onNode(t, products[m].source, "--manifest", list), len(texts))
			native := canonicalRecords(t, execute(t, "", products[m].native, "--manifest", list), len(texts))
			witnesses := 0
			for i, pair := range shard.cases {
				if err := checkMutantCase(pair, want[i], source[i], native[i]); err != nil {
					t.Fatalf("%s: %v", mutations[m].name, err)
				}
				if pair.witness {
					witnesses++
				}
			}
			t.Logf("%s: cases [%d,%d), %d pairs, %d killing witnesses; Go/source Node/sanitized native checked", mutations[m].name, shard.cases[0].index, shard.cases[len(shard.cases)-1].index+1, len(texts), witnesses)
		})
	}
	for m, yes := range needed {
		if !yes {
			continue
		}
		if oracle == "" {
			oracle = goOracle(t)
		}
		mutation := mutations[m]
		path := mutantPort(t, mutation.file, mutation.from, mutation.to)
		binary, _ := build(t, path, true)
		products[m] = product{path, binary}
	}
	setup = time.Since(start)
}

// Plant one surviving native mutant in its sole witness case, then execute all
// named shards in a child test process. Assert exactly one failing shard and
// every other shard passing, including the actual fatal diagnostic and case id.
func TestThreePortMutantShardProof(t *testing.T) {
	proveMutantShards(t, threePortMutations(), threePortShards(t))
}

func proveMutantShards(t *testing.T, mutations []portMutation, shards []mutantShard) {
	t.Helper()
	parent := t.Name()
	if value := os.Getenv("ADAMIC_ESTREE_PLANTED_MUTANT"); value != "" {
		m, err := strconv.Atoi(value)
		if err != nil || m < 0 || m >= len(mutations) {
			t.Fatal("invalid planted mutant")
		}
		for _, shard := range shards {
			t.Run(shard.name, func(t *testing.T) {
				t.Parallel()
				for _, pair := range shard.cases {
					want, source, native := []byte("oracle"), []byte("oracle"), []byte("oracle")
					if pair.witness {
						source, native = []byte("mutant"), []byte("mutant")
					}
					if pair.mutant == m && pair.witness {
						native = want
					}
					if err := checkMutantCase(pair, want, source, native); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for m, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			expected, caseID := "", -1
			for _, shard := range shards {
				for _, pair := range shard.cases {
					if pair.mutant == m && pair.witness {
						expected, caseID = shard.name, pair.index
					}
				}
			}
			command := exec.Command(executable, "-test.run=^"+parent+"$", "-test.v", "-test.count=1")
			command.Env = append(os.Environ(), fmt.Sprintf("ADAMIC_ESTREE_PLANTED_MUTANT=%d", m))
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				t.Fatalf("planted failure: exit=%v; %s", err, output)
			}
			failed, passed := 0, 0
			for _, line := range strings.Split(string(output), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "--- FAIL: "+parent+"/shard-") {
					failed++
					if !strings.HasPrefix(line, "--- FAIL: "+parent+"/"+expected+" (") {
						t.Fatalf("wrong shard caught failure: %s", line)
					}
				}
				if strings.HasPrefix(line, "--- PASS: "+parent+"/shard-") {
					passed++
				}
			}
			diagnostic := fmt.Sprintf("case %03d: sanitized native mutant survived", caseID)
			if failed != 1 || passed != len(shards)-1 || !strings.Contains(string(output), diagnostic) {
				t.Fatalf("planted failure: failed=%d passed=%d want 1/%d, diagnostic %q; %s", failed, passed, len(shards)-1, diagnostic, output)
			}
			t.Logf("planted surviving native mutant %s case %03d caught by %s; exactly one shard failed, %d passed", mutation.name, caseID, expected, passed)
		})
	}
}
