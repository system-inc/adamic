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
	for index, source := range recoveredGrammar() {
		for _, recoveryMutation := range recoveryMutations {
			cases = append(cases, recoveryCase{fmt.Sprintf("%s/%03d", recoveryMutation.name, index), source})
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
	for _, recoveryMutation := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			id := fmt.Sprintf("%s/%x/%d", recoveryMutation.name, sum, occurrences[sum])
			occurrences[sum]++
			if seen[id] {
				test.Fatalf("duplicate case %s", id)
			}
			seen[id] = true
			owner := recoveryMutantOwner(recoveryMutation.name)
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
	for _, recoveryMutation := range recoveryMutations {
		occurrences := map[[32]byte]int{}
		for _, source := range sources {
			sum := sha256.Sum256([]byte(source))
			expected[fmt.Sprintf("%s/%x/%d", recoveryMutation.name, sum, occurrences[sum])] = true
			occurrences[sum]++
		}
	}
	seen := map[string]bool{}
	for index, cases := range shards {
		for _, recoveryCase := range cases {
			name := strings.SplitN(recoveryCase.id, "/", 2)[0]
			if !expected[recoveryCase.id] || seen[recoveryCase.id] || recoveryMutantOwner(name) != index {
				test.Fatalf("bad union case %s in shard %03d", recoveryCase.id, index)
			}
			seen[recoveryCase.id] = true
		}
	}
	if len(seen) != len(expected) {
		test.Fatalf("union %d != live enumeration %d", len(seen), len(expected))
	}
	test.Logf("union: %d cases, %d unique IDs, %d shards", len(seen), len(expected), len(shards))
	// Adding a live source must preserve the assignment of every old case.
	grown := recoveryMutantLiveSlices(test, append(append([]string{}, sources...), "new corpus growth sentinel"))
	for index, cases := range shards {
		ids := map[string]bool{}
		for _, recoveryCase := range grown[index] {
			ids[recoveryCase.id] = true
		}
		for _, recoveryCase := range cases {
			if !ids[recoveryCase.id] {
				test.Fatalf("growth moved %s", recoveryCase.id)
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
		for _, recoveryMutation := range recoveryMutations {
			if recoveryMutantOwner(recoveryMutation.name) == shard {
				products[recoveryMutation.name] = recoveryMutantPrepared(test, recoveryMutation.name)
			}
		}
		if len(cases) != 0 {
			want = recoveryMutantAnswer(test)
		}
	}
	// Preparation is outside the shard work budget; Loom bounds the whole unit.
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	defer func() { test.Logf("shard work wall=%.6fs", time.Since(started).Seconds()) }()
	planted := os.Getenv("ADAMIC_RECOVERY_MUTANT_SURVIVOR")
	if planted != "" {
		for _, recoveryCase := range cases {
			got := []byte("disagreement")
			if recoveryCase.id == planted {
				got = []byte("oracle")
			}
			if failure := recoveryComparison([]byte("oracle"), got, true); failure != "" {
				test.Fatal(recoveryCase.id + ": " + failure)
			}
		}
		return
	}

	if len(cases) == 0 {
		return
	}
	list := manifest(test, sources)
	for _, recoveryMutation := range recoveryMutations {
		if recoveryMutantOwner(recoveryMutation.name) != shard {
			continue
		}
		count := 0
		for _, recoveryCase := range cases {
			if strings.HasPrefix(recoveryCase.id, recoveryMutation.name+"/") {
				count++
			}
		}
		if count != len(sources) {
			test.Fatalf("%s lost grammar cases: %d/%d", recoveryMutation.name, count, len(sources))
		}
		main := mutantPort(test, recoveryMutation.file, recoveryMutation.from, recoveryMutation.to)
		product := products[recoveryMutation.name]
		binary := filepath.Join(product, "port")
		for name, got := range map[string][]byte{"Node": recoveryMutantExecute(
			test,
			timeoutContext,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(root(test), "oracle/node.mjs"),
			main,
			"--manifest",
			list,
		), "native": recoveryMutantExecute(test, timeoutContext, binary, "--manifest", list)} {
			if failure := recoveryComparison(want, got, true); failure != "" {
				test.Fatal(recoveryMutation.name + " " + name + ": " + failure)
			}
			test.Log(recoveryMutation.name + " " + name + ": " + firstDifference(want, got))
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

func recoveryMutantExecute(test *testing.T, callerContext context.Context, name string, args ...string) []byte {
	test.Helper()
	// Read through childguard.Capture for estreeScopedCommand's reason; the caller's 90 s deadline ends a hung run.
	output, stderr, captureError := childguard.Capture(callerContext, estreeScopedCommand(callerContext, name, args...))
	if captureError != nil || len(stderr) != 0 {
		test.Fatalf("%s %v: %v (deadline: %v)\n%s", name, args, captureError, callerContext.Err(), stderr)
	}
	return output
}
