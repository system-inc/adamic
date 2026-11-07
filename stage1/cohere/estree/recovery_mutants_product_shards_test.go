package estree

import (
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

	"github.com/system-inc/adamic/internal/childguard"
)

const testRecoveryMutantsShards = 3

var recoveryMutations = []struct{ name, file, from, to string }{
	{
		"first-accessibility",
		"modifiers.ts",
		"return stringValue(kind.slice(0, -7).toLowerCase());",
		"return stringValue('public');",
	},
	{
		"empty-type-list-range",
		"typeLists.ts",
		"arena.node(result).set('params', listValue([]));",
		"arena.node(result).end -= 1; arena.node(result).set('params', listValue([]));",
	},
	{
		"module-await",
		"pipeline.ts",
		"&& externalModule(parser.nodes, root)",
		"&& false && externalModule(parser.nodes, root)",
	},
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
func recoveryMutantLiveSlices(test *testing.T, sources []string) [][]recoveryCase {
	test.Helper()
	if len(sources) == 0 || len(recoveryMutations) == 0 {
		test.Fatal("empty recovery mutant corpus")
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
				test.Fatalf("duplicate case %s", id)
			}
			seen[id] = true
			owner := recoveryMutantOwner(m.name)
			shards[owner] = append(shards[owner], recoveryCase{id: id, source: source})
			total++
		}
	}
	if total != len(sources)*len(recoveryMutations) {
		test.Fatal("live enumeration mismatch")
	}
	return shards
}
func checkRecoveryMutantUnion(test *testing.T) {
	test.Helper()
	declared := [...]func(*testing.T){TestRecoveryMutants_000, TestRecoveryMutants_001, TestRecoveryMutants_002}
	if len(declared) != testRecoveryMutantsShards {
		test.Fatal("top-level shard enumeration changed")
	}
	sources := recoveredGrammar()
	shards := recoveryMutantLiveSlices(test, sources)
	if len(shards) != testRecoveryMutantsShards {
		test.Fatal("shard count mismatch")
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
				test.Fatalf("bad union case %s in shard %03d", c.id, i)
			}
			seen[c.id] = true
		}
	}
	if len(seen) != len(expected) {
		test.Fatalf("union %d != live enumeration %d", len(seen), len(expected))
	}
	test.Logf("union: %d cases, %d unique IDs, %d shards", len(seen), len(expected), len(shards))
	// Adding a live source must preserve the assignment of every old case.
	grown := recoveryMutantLiveSlices(test, append(append([]string{}, sources...), "new corpus growth sentinel"))
	for i, cases := range shards {
		ids := map[string]bool{}
		for _, c := range grown[i] {
			ids[c.id] = true
		}
		for _, c := range cases {
			if !ids[c.id] {
				test.Fatalf("growth moved %s", c.id)
			}
		}
	}
}

func runRecoveryMutantTop(test *testing.T, shard int) {
	selected, n := recoveryShardSelection(test)
	if shard%n != selected {
		test.Skip("ADAMIC_TEST_SHARD selects another shard")
	}
	sources := recoveredGrammar()
	cases := recoveryMutantLiveSlices(test, sources)[shard]
	products := map[string]string{}
	var want []byte
	if os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR") == "" {
		for _, m := range recoveryMutations {
			if recoveryMutantOwner(m.name) == shard {
				products[m.name] = recoveryMutantPrepared(test, m.name)
			}
		}
		if len(cases) != 0 {
			want = recoveryMutantAnswer(test)
		}
	}
	// Preparation is outside the shard work budget; Loom bounds the whole unit.
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	defer func() { test.Logf("shard work wall=%.6fs", time.Since(started).Seconds()) }()
	planted := os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR")
	if planted != "" {
		for _, c := range cases {
			got := []byte("disagreement")
			if c.id == planted {
				got = []byte("oracle")
			}
			if failure := recoveryComparison([]byte("oracle"), got, true); failure != "" {
				test.Fatal(c.id + ": " + failure)
			}
		}
		return
	}

	if len(cases) == 0 {
		return
	}
	list := manifest(test, sources)
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
			test.Fatalf("%s lost grammar cases: %d/%d", m.name, count, len(sources))
		}
		main := mutantPort(test, m.file, m.from, m.to)
		product := products[m.name]
		binary := filepath.Join(product, "port")
		for name, got := range map[string][]byte{"Node": recoveryMutantExecute(
			test,
			ctx,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(root(test), "oracle/node.mjs"),
			main,
			"--manifest",
			list,
		), "native": recoveryMutantExecute(test, ctx, binary, "--manifest", list)} {
			if failure := recoveryComparison(want, got, true); failure != "" {
				test.Fatal(m.name + " " + name + ": " + failure)
			}
			test.Log(m.name + " " + name + ": " + firstDifference(want, got))
		}
	}
}

type recoveryMutantProduct struct {
	once      sync.Once
	directory string
}

var recoveryMutantProducts sync.Map

func recoveryMutantPrepared(test *testing.T, name string) string {
	test.Helper()
	value, _ := recoveryMutantProducts.LoadOrStore(name, &recoveryMutantProduct{})
	product := value.(*recoveryMutantProduct)
	product.once.Do(func() {
		product.directory = buildRecoveryMutantProduct(test, name)
	})
	if product.directory == "" {
		test.Fatalf("preparation failed for mutant %s", name)
	}
	return product.directory
}

func recoveryMutantMutation(test *testing.T, name string) threePortMutation {
	test.Helper()
	for _, mutation := range recoveryMutations {
		if mutation.name == name {
			return threePortMutation{
				name: "recovery-" + name,
				file: mutation.file,
				from: mutation.from,
				to:   mutation.to,
			}
		}
	}
	test.Fatalf("unknown recovery mutant %q", name)
	return threePortMutation{}
}

func buildRecoveryMutantProduct(test *testing.T, name string) string {
	test.Helper()
	return filepath.Dir(threePortNativeProduct(test, recoveryMutantMutation(test, name)))
}

var recoveryMutantAnswerPrepared struct {
	once   sync.Once
	answer []byte
}

func recoveryMutantAnswer(test *testing.T) []byte {
	test.Helper()
	recoveryMutantAnswerPrepared.once.Do(func() {
		list := manifest(test, recoveredGrammar())
		recoveryMutantAnswerPrepared.answer = recoveryAnswer(test, scalarEdgeOracle(test), list, "--manifest")
	})
	if len(recoveryMutantAnswerPrepared.answer) == 0 {
		test.Fatal("recovery answer preparation failed")
	}
	return recoveryMutantAnswerPrepared.answer
}

func TestProduct_EstreeFindsRecoveryMutantsAnswer(test *testing.T) {
	test.Parallel()
	recoveryMutantAnswer(test)
}
func TestProduct_RecoveryMutantsLowered0(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, recoveryMutantMutation(test, recoveryMutations[0].name))
}
func TestProduct_RecoveryMutantsNative0(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, recoveryMutantMutation(test, recoveryMutations[0].name))
}
func TestProduct_RecoveryMutantsLowered1(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, recoveryMutantMutation(test, recoveryMutations[1].name))
}
func TestProduct_RecoveryMutantsNative1(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, recoveryMutantMutation(test, recoveryMutations[1].name))
}
func TestProduct_RecoveryMutantsLowered2(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, recoveryMutantMutation(test, recoveryMutations[2].name))
}
func TestProduct_RecoveryMutantsNative2(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, recoveryMutantMutation(test, recoveryMutations[2].name))
}

func recoveryMutantExecute(test *testing.T, ctx context.Context, name string, args ...string) []byte {
	test.Helper()
	// Read through childguard.Capture for estreeScopedCommand's reason; the caller's 90 s deadline ends a hung run.
	output, stderr, err := childguard.Capture(ctx, estreeScopedCommand(ctx, name, args...))
	if err != nil || len(stderr) != 0 {
		test.Fatalf("%s %v: %v (deadline: %v)\n%s", name, args, err, ctx.Err(), stderr)
	}
	return output
}
