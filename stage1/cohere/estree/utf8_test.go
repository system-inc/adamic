package estree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func cookedSurrogates() []string {
	return []string{"'\\ud800a\\udc00';", "'\\ud800';", "'\\udfff';", "'\\ud800\\udc00';", "'\\ud800\\ud800';", "`a\\ud800b`;", "tag`a\\ud800b`;", "const key = {'\\ud800': 1};"}
}

const testCookedSurrogatesShards = 8

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
func TestCookedSurrogates(t *testing.T) {
	estreeAgreementShards(t, testCookedSurrogatesShards, cookedSurrogates())
}
func TestCookedSurrogatesShardFailure(t *testing.T) {
	sources := cookedSurrogates()
	estreeShardFailure(t, testCookedSurrogatesShards, len(sources), estreeSingles(len(sources)), "agreement", 4)
}

const testCookedSurrogateMutantShards = 1

func cookedSurrogateMutantGroups() [][]int {
	group := make([]int, len(cookedSurrogates()))
	for id := range group {
		group[id] = id
	}
	return [][]int{group}
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
// Keep the complete manifest together: the original predicate requires some
// input to kill this mutant, and paired-surrogate controls need not differ.
func TestCookedSurrogateMutant(t *testing.T) {
	estreeAccounting(t)
	started := time.Now()
	sources := cookedSurrogates()
	selected := estreeShardPlan(t, testCookedSurrogateMutantShards, len(sources), cookedSurrogateMutantGroups())
	oracle := estreeTimedOracle(t)
	main := mutantPort(t, "protocol.ts", `result += '\\ufffd\\ufffd\\ufffd';`, `result += '\\ufffd';`)
	binary, _ := estreeTimedBuild(t, main, true)
	t.Logf("setup including builds: %.3fs", time.Since(started).Seconds())
	if selected[0] {
		t.Run("shard-000", func(t *testing.T) {
			t.Parallel()
			list := manifest(t, sources)
			want := execute(t, "", oracle, "--manifest", list)
			for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
				if err := estreeMutantVerdict(want, got); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
			}
		})
	}
}
func TestCookedSurrogateMutantShardFailure(t *testing.T) {
	estreeShardFailure(t, testCookedSurrogateMutantShards, len(cookedSurrogates()), cookedSurrogateMutantGroups(), "mutant", 0)
}

func TestCookedSurrogateLibraryGap(t *testing.T) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	list := manifest(t, []string{cookedSurrogates()[0]})
	goAnswer := string(execute(t, "", goOracle(t), "--manifest", list))
	if !strings.Contains(goAnswer, `.value string \ufffd\ufffd\ufffda\ufffd\ufffd\ufffd`) {
		t.Fatal("Go WTF-8 gap changed")
	}
	code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);const ast=lib.parse("'\\ud800a\\udc00';",{warnOnUnsupportedTypeScriptVersion:false});console.log(JSON.stringify(ast.body[0].expression.value));`
	got := string(execute(t, "", "node", "--input-type=module", "-e", code, library))
	if got != `"\ud800a\udc00"`+"\n" {
		t.Fatalf("library surrogate gap changed: %q", got)
	}
	t.Log("Go serializes each unpaired WTF-8 surrogate as three U+FFFD; original keeps UTF-16 surrogates")
}

const testLossyInputRefusalShards = 2

func lossyInputs() [][]byte {
	return [][]byte{{47, 47, 240, 144, 128, 10, 120, 59}, {47, 47, 239, 191, 189, 10, 120, 59}}
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
func TestLossyInputRefusal(t *testing.T) {
	estreeAccounting(t)
	started := time.Now()
	bodies := lossyInputs()
	selected := estreeShardPlan(t, testLossyInputRefusalShards, len(bodies), estreeSingles(len(bodies)))
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := estreeTimedBuild(t, main, true)
	t.Logf("setup including builds: %.3fs", time.Since(started).Seconds())
	for i, body := range bodies {
		if !selected[i] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "input.ts")
			if err := os.WriteFile(path, body, 0644); err != nil {
				t.Fatal(err)
			}
			for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
				estreeRefused(t, argv, "cannot recover original UTF-8 bytes")
			}
		})
	}
}
func TestLossyInputRefusalShardFailure(t *testing.T) {
	bodies := lossyInputs()
	estreeShardFailure(t, testLossyInputRefusalShards, len(bodies), estreeSingles(len(bodies)), "refusal", 1)
}

const testLossyInputControlShards = 1

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
func TestLossyInputControl(t *testing.T) {
	estreeAccounting(t)
	started := time.Now()
	selected := estreeShardPlan(t, testLossyInputControlShards, 1, estreeSingles(1))
	oracle := estreeTimedOracle(t)
	main := mutantPort(t, "pipeline.ts", `if(text.includes('\ufffd'))`, `if(false)`)
	binary, _ := estreeTimedBuild(t, main, true)
	t.Logf("setup including builds: %.3fs", time.Since(started).Seconds())
	if selected[0] {
		t.Run("shard-000", func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "malformed.ts")
			if err := os.WriteFile(path, lossyInputs()[0], 0644); err != nil {
				t.Fatal(err)
			}
			want := execute(t, "", oracle, path)
			for name, got := range map[string][]byte{"Node": onNode(t, main, path), "native": execute(t, "", binary, path)} {
				if err := estreeMutantVerdict(want, got); err != nil {
					t.Fatalf("%s lossy-input: %v", name, err)
				}
			}
		})
	}
}
func TestLossyInputControlShardFailure(t *testing.T) {
	estreeShardFailure(t, testLossyInputControlShards, 1, estreeSingles(1), "mutant", 0)
}
