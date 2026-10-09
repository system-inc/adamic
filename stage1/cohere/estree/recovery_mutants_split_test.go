package estree

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const testRecoveryMutantsShards = 3

var recoveryMutations = []struct{ name, file, from, to string }{
	{"first-accessibility", "modifiers.ts", "return stringValue(kind.slice(0, -7).toLowerCase());", "return stringValue('public');"},
	{"empty-type-list-range", "typeLists.ts", "arena.node(result).set('params', listValue([]));", "arena.node(result).end -= 1; arena.node(result).set('params', listValue([]));"},
	{"module-await", "pipeline.ts", "&& externalModule(parser.nodes, root)", "&& false && externalModule(parser.nodes, root)"},
}

func recoveryMutantCases() []recoveryCase {
	var cases []recoveryCase
	// Interleave mutants so partition i%3 assigns one complete corpus per mutant.
	for i, source := range recoveredGrammar() {
		for _, m := range recoveryMutations {
			cases = append(cases, recoveryCase{fmt.Sprintf("%s/%03d", m.name, i), source})
		}
	}
	return cases
}

// ADAMIC_TEST_SHARD=i/n runs shards whose index modulo n is i; unset runs all.
// Each mutant retains the entire grammar corpus on both Node and sanitized native.
func TestRecoveryMutants(t *testing.T) {
	t.Parallel()
	// Compatibility enumeration only; execution lives in the top-level shards.
	checkRecoveryMutantUnion(t)
}

func TestRecoveryMutantsShardSurvivor(t *testing.T) {
	t.Parallel()
	proveRecoveryShard(t, recoveryMutantCases(), testRecoveryMutantsShards, true)
}

// ADAMIC_TEST_SHARD=i/n selects top-level shards whose index modulo n is i.
// The fixed shard count has headroom: growing the live grammar does not move a
// mutant. Each mutant checks the complete grammar on Node and sanitized native.
func TestRecoveryMutants_000(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 0) }
func TestRecoveryMutants_001(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 1) }
func TestRecoveryMutants_002(t *testing.T) { t.Parallel(); runRecoveryMutantTop(t, 2) }

func recoveryMutantOwner(name string) int {
	key := "stage1/cohere/estree/recovery_test.go/TestRecoveryMutants/" + name
	sum := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(testRecoveryMutantsShards))
}
func recoveryMutantLiveSlices(t *testing.T, sources []string) [][]recoveryCase {
	t.Helper()
	if len(sources) == 0 || len(recoveryMutations) == 0 {
		t.Fatal("empty recovery mutant corpus")
	}
	shards := make([][]recoveryCase, testRecoveryMutantsShards)
	seen := map[string]bool{}
	total := 0
	for _, m := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			id := fmt.Sprintf("%s/%x/%d", m.name, sum, occurrences[sum])
			occurrences[sum]++
			if seen[id] {
				t.Fatalf("duplicate case %s", id)
			}
			seen[id] = true
			owner := recoveryMutantOwner(m.name)
			shards[owner] = append(shards[owner], recoveryCase{id: id, source: source})
			total++
		}
	}
	if total != len(sources)*len(recoveryMutations) {
		t.Fatal("live enumeration mismatch")
	}
	return shards
}
func checkRecoveryMutantUnion(t *testing.T) {
	t.Helper()
	declared := [...]func(*testing.T){TestRecoveryMutants_000, TestRecoveryMutants_001, TestRecoveryMutants_002}
	if len(declared) != testRecoveryMutantsShards {
		t.Fatal("top-level shard enumeration changed")
	}
	sources := recoveredGrammar()
	shards := recoveryMutantLiveSlices(t, sources)
	if len(shards) != testRecoveryMutantsShards {
		t.Fatal("shard count mismatch")
	}
	// Independent live enumeration: no fixed corpus total or positional keys.
	expected := map[string]bool{}
	for _, m := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			expected[fmt.Sprintf("%s/%x/%d", m.name, sum, occurrences[sum])] = true
			occurrences[sum]++
		}
	}
	seen := map[string]bool{}
	for i, cases := range shards {
		for _, c := range cases {
			name := strings.SplitN(c.id, "/", 2)[0]
			if !expected[c.id] || seen[c.id] || recoveryMutantOwner(name) != i {
				t.Fatalf("bad union case %s in shard %03d", c.id, i)
			}
			seen[c.id] = true
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("union %d != live enumeration %d", len(seen), len(expected))
	}
	t.Logf("union: %d cases, %d unique IDs, %d shards", len(seen), len(expected), len(shards))
	// Adding a live source must preserve the assignment of every old case.
	grown := recoveryMutantLiveSlices(t, append(append([]string{}, sources...), "new corpus growth sentinel"))
	for i, cases := range shards {
		ids := map[string]bool{}
		for _, c := range grown[i] {
			ids[c.id] = true
		}
		for _, c := range cases {
			if !ids[c.id] {
				t.Fatalf("growth moved %s", c.id)
			}
		}
	}
}
func TestRecoveryMutantsUnion(t *testing.T) { t.Parallel(); checkRecoveryMutantUnion(t) }

func runRecoveryMutantTop(t *testing.T, shard int) {
	beginRecoverySetup(t)
	selected, n := recoveryShardSelection(t)
	if shard%n != selected {
		t.Skip("ADAMIC_TEST_SHARD selects another shard")
	}
	sources := recoveredGrammar()
	cases := recoveryMutantLiveSlices(t, sources)[shard]
	planted := os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR")
	if planted != "" {
		for _, c := range cases {
			got := []byte("disagreement")
			if c.id == planted {
				got = []byte("oracle")
			}
			if failure := recoveryComparison([]byte("oracle"), got, true); failure != "" {
				t.Fatal(c.id + ": " + failure)
			}
		}
		return
	}

	if len(cases) == 0 {
		return
	}
	list := manifest(t, sources)
	for _, m := range recoveryMutations {
		if recoveryMutantOwner(m.name) != shard {
			continue
		}
		count := 0
		for _, c := range cases {
			if strings.HasPrefix(c.id, m.name+"/") {
				count++
			}
		}
		if count != len(sources) {
			t.Fatalf("%s lost grammar cases: %d/%d", m.name, count, len(sources))
		}
		main := mutantPort(t, m.file, m.from, m.to)
		product := recoveryMutantPrepared(t, m.name, nil)
		binary := filepath.Join(product, "port")
		want, err := os.ReadFile(filepath.Join(product, "answer"))
		if err != nil {
			t.Fatal(err)
		}
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if failure := recoveryComparison(want, got, true); failure != "" {
				t.Fatal(m.name + " " + name + ": " + failure)
			}
			t.Log(m.name + " " + name + ": " + firstDifference(want, got))
		}
	}
}
func TestRecoveryMutantsTopSurvivor(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	planted := recoveryMutantLiveSlices(t, recoveredGrammar())[recoveryMutantOwner(recoveryMutations[0].name)][0].id
	var output []byte
	err = nil
	for shard := 0; shard < testRecoveryMutantsShards; shard++ {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "-test.v", "-test.timeout=75s", fmt.Sprintf("-test.run=^TestRecoveryMutants_%03d$", shard))
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = time.Second
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_RECOVERY_MUTANT_SURVIVOR=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "ADAMIC_RECOVERY_MUTANT_SURVIVOR="+planted)
		part, childErr := command.CombinedOutput()
		output = append(output, part...)
		if childErr != nil {
			err = childErr
		}
		if ctx.Err() != nil {
			t.Fatalf("planted survivor subprocess cooked: %v", ctx.Err())
		}

	}
	owner := fmt.Sprintf("TestRecoveryMutants_%03d", recoveryMutantOwner(strings.SplitN(planted, "/", 2)[0]))
	prefix := "--- FAIL: TestRecoveryMutants_"
	if err == nil || strings.Count(string(output), prefix) != 1 || !strings.Contains(string(output), "--- FAIL: "+owner+" ") || !strings.Contains(string(output), "mutant survived") {
		t.Fatalf("planted survivor must fail only %s: %v\n%s", owner, err, output)
	}
	t.Logf("planted surviving mutant %s caught only by %s", planted, owner)
}

// Not parallel: prepare shared products before the parallel execution shards resume.
func TestRecoveryMutants_Setup(t *testing.T) {
	started := time.Now()
	cpu := recoveryCPU(t)
	t.Cleanup(func() {
		t.Logf("TestRecoveryMutants (setup): wall=%.6fs totalCPU=%.6fs", time.Since(started).Seconds(), (recoveryCPU(t) - cpu).Seconds())
	})
	setup := beginRecoverySetup(t)
	list := manifest(t, recoveredGrammar())
	want := recoveryAnswer(t, recoveryOracle(t, setup), list, "--manifest")
	for _, mutation := range recoveryMutations {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()
			recoveryMutantPrepared(t, mutation.name, func(directory string) error {
				main := mutantPort(t, mutation.file, mutation.from, mutation.to)
				binary, _ := recoveryPort(t, beginRecoverySetup(t), main)
				data, err := os.ReadFile(binary)
				if err != nil {
					return err
				}
				if err = os.WriteFile(filepath.Join(directory, "port"), data, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(directory, "answer"), want, 0644)
			})
		})
	}
}

// These are setup's Product handles, including uncached products when the gate
// sets ADAMIC_BUILD_CACHE=off. No execution shard invokes a builder.
var recoveryMutantProducts sync.Map

func recoveryMutantPrepared(t *testing.T, name string, build func(string) error) string {
	t.Helper()
	if build == nil {
		if product, ok := recoveryMutantProducts.Load(name); ok {
			return product.(string)
		}
	}
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	inputs := recoveryCacheInputs(t, main)
	inputs.Name = "estree-recovery-mutant-prepared-" + name
	inputs.Files = append(inputs.Files, "stage1/cohere/estree/recovery_mutants_split_test.go", "stage1/cohere/estree/recovery_test.go", "stage1/cohere/estree/testdata/oracle.go", "cohere/internal/format", "oracle")
	inputs.Flags = append(inputs.Flags, "sanitized", "full-recovered-grammar")
	for _, m := range recoveryMutations {
		if m.name == name {
			inputs.Flags = append(inputs.Flags, m.file, m.from, m.to)
		}
	}
	inputs.Flags = append(inputs.Flags, recoveredGrammar()...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"), buildcache.Tool("node", "--version"), buildcache.Tool("getconf", "GNU_LIBC_VERSION"))

	if build == nil {
		build = func(string) error { return fmt.Errorf("run TestRecoveryMutants_Setup first; shards never build") }
	}
	product := buildcache.Product(t, inputs, build)
	recoveryMutantProducts.Store(name, product)
	return product
}
