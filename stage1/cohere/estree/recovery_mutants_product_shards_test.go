package estree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func runRecoveryMutantTop(t *testing.T, shard int) {
	selected, n := recoveryShardSelection(t)
	if shard%n != selected {
		t.Skip("ADAMIC_TEST_SHARD selects another shard")
	}
	sources := recoveredGrammar()
	cases := recoveryMutantLiveSlices(t, sources)[shard]
	products := map[string]string{}
	emitted := map[string]string{}
	sourcePaths := map[string]string{}
	var want []byte
	if os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR") == "" {
		for _, m := range recoveryMutations {
			if recoveryMutantOwner(m.name) == shard {
				products[m.name] = recoveryMutantPrepared(t, m.name)
				sourcePaths[m.name] = mutantPort(t, m.file, m.from, m.to)
				emitted[m.name] = mutantEmittedProduct(t, sourcePaths[m.name])
			}
		}
		if len(cases) != 0 {
			want = recoveryMutantAnswer(t)
		}
	}
	// Preparation is outside the shard work budget; Loom bounds the whole unit.
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	defer func() { t.Logf("shard work wall=%.6fs", time.Since(started).Seconds()) }()
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
		main := sourcePaths[m.name]
		product := products[m.name]
		binary := filepath.Join(product, "port")
		for name, got := range map[string][]byte{"Node": recoveryMutantExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, "--manifest", list), "native": recoveryMutantExecute(t, ctx, binary, "--manifest", list), "emitted": mutantEmittedOutput(t, main, emitted[m.name], "--manifest", list)} {
			if failure := recoveryComparison(want, got, true); failure != "" {
				t.Fatal(m.name + " " + name + ": " + failure)
			}
			t.Log(m.name + " " + name + ": " + firstDifference(want, got))
		}
	}
}

type recoveryMutantProduct struct {
	once      sync.Once
	directory string
}

var recoveryMutantProducts sync.Map

func recoveryMutantPrepared(t *testing.T, name string) string {
	t.Helper()
	value, _ := recoveryMutantProducts.LoadOrStore(name, &recoveryMutantProduct{})
	product := value.(*recoveryMutantProduct)
	product.once.Do(func() {
		product.directory = buildRecoveryMutantProduct(t, name)
	})
	if product.directory == "" {
		t.Fatalf("preparation failed for mutant %s", name)
	}
	return product.directory
}

func recoveryMutantMutation(t *testing.T, name string) threePortMutation {
	t.Helper()
	for _, mutation := range recoveryMutations {
		if mutation.name == name {
			return threePortMutation{name: "recovery-" + name, file: mutation.file, from: mutation.from, to: mutation.to}
		}
	}
	t.Fatalf("unknown recovery mutant %q", name)
	return threePortMutation{}
}

func buildRecoveryMutantProduct(t *testing.T, name string) string {
	t.Helper()
	return filepath.Dir(threePortNativeProduct(t, recoveryMutantMutation(t, name)))
}

var recoveryMutantAnswerPrepared struct {
	once   sync.Once
	answer []byte
}

func recoveryMutantAnswer(t *testing.T) []byte {
	t.Helper()
	recoveryMutantAnswerPrepared.once.Do(func() {
		list := manifest(t, recoveredGrammar())
		recoveryMutantAnswerPrepared.answer = recoveryAnswer(t, scalarEdgeOracle(t), list, "--manifest")
	})
	if len(recoveryMutantAnswerPrepared.answer) == 0 {
		t.Fatal("recovery answer preparation failed")
	}
	return recoveryMutantAnswerPrepared.answer
}

func TestProduct_RecoveryMutantsAnswer(t *testing.T) {
	t.Parallel()
	recoveryMutantAnswer(t)
}
func TestProduct_RecoveryMutantsLowered0(t *testing.T) {
	t.Parallel()
	threePortLoweredProduct(t, recoveryMutantMutation(t, recoveryMutations[0].name))
}
func TestProduct_RecoveryMutantsNative0(t *testing.T) {
	t.Parallel()
	threePortNativeProduct(t, recoveryMutantMutation(t, recoveryMutations[0].name))
}
func TestProduct_RecoveryMutantsLowered1(t *testing.T) {
	t.Parallel()
	threePortLoweredProduct(t, recoveryMutantMutation(t, recoveryMutations[1].name))
}
func TestProduct_RecoveryMutantsNative1(t *testing.T) {
	t.Parallel()
	threePortNativeProduct(t, recoveryMutantMutation(t, recoveryMutations[1].name))
}
func TestProduct_RecoveryMutantsLowered2(t *testing.T) {
	t.Parallel()
	threePortLoweredProduct(t, recoveryMutantMutation(t, recoveryMutations[2].name))
}
func TestProduct_RecoveryMutantsNative2(t *testing.T) {
	t.Parallel()
	threePortNativeProduct(t, recoveryMutantMutation(t, recoveryMutations[2].name))
}

func recoveryMutantExecute(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := estreeScopedCommand(ctx, name, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v (deadline: %v)\n%s", name, args, err, ctx.Err(), &stderr)
	}
	return output
}
