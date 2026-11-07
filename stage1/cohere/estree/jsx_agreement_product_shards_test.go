package estree

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

const testJSXAgreementShards = 32

func jsxAgreementIDs() []string { return miscIDs("stage1/cohere/estree/jsx_test.go:jsx", jsxCases()) }

// The original selector checks the live union; actual case work is in top-level shards.
func TestJSXAgreement(test *testing.T) {
	test.Parallel()
	estreeTopUnion(test, "TestJSXAgreement", testJSXAgreementShards, jsxAgreementIDs())
}

func jsxAgreementShard(test *testing.T, shard int) {
	test.Helper()
	cases, ids := jsxCases(), jsxAgreementIDs()
	indexes := estreeScopedIndexes(test, testJSXAgreementShards, ids, shard)
	if estreeScopedProof(test, "TestJSXAgreement", ids, indexes) || len(indexes) == 0 {
		return
	}
	// scalarEdgePrepare uses sync.Once and the declared shared Product recipes.
	// Setup has no deadline; only this shard's case work starts the clock below.
	ready := scalarEdgePrepare(test)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	main := filepath.Join(root(test), "stage1/cohere/estree/main.ts")
	for _, index := range indexes {
		list := jsxManifestCases(test, cases[index:index+1])
		want := scalarEdgeExecute(test, timeoutContext, ready.Oracle, "--manifest", list)
		for name, got := range map[string][]byte{
			"Node":       scalarEdgeNode(test, timeoutContext, main, "--manifest", list),
			"native":     scalarEdgeExecute(test, timeoutContext, ready.Binary, "--manifest", list),
			"emitted JS": scalarEdgeNode(test, timeoutContext, ready.Script, "--manifest", list),
		} {
			if miscCompareError := miscCompare(want, got, false); miscCompareError != nil {
				test.Fatalf("%s %s: %v", ids[index], name, miscCompareError)
			}
		}
	}
}

func TestJSXAgreement_000(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 0) }
func TestJSXAgreement_001(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 1) }
func TestJSXAgreement_002(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 2) }
func TestJSXAgreement_003(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 3) }
func TestJSXAgreement_004(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 4) }
func TestJSXAgreement_005(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 5) }
func TestJSXAgreement_006(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 6) }
func TestJSXAgreement_007(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 7) }
func TestJSXAgreement_008(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 8) }
func TestJSXAgreement_009(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 9) }
func TestJSXAgreement_010(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 10) }
func TestJSXAgreement_011(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 11) }
func TestJSXAgreement_012(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 12) }
func TestJSXAgreement_013(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 13) }
func TestJSXAgreement_014(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 14) }
func TestJSXAgreement_015(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 15) }
func TestJSXAgreement_016(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 16) }
func TestJSXAgreement_017(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 17) }
func TestJSXAgreement_018(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 18) }
func TestJSXAgreement_019(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 19) }
func TestJSXAgreement_020(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 20) }
func TestJSXAgreement_021(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 21) }
func TestJSXAgreement_022(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 22) }
func TestJSXAgreement_023(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 23) }
func TestJSXAgreement_024(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 24) }
func TestJSXAgreement_025(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 25) }
func TestJSXAgreement_026(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 26) }
func TestJSXAgreement_027(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 27) }
func TestJSXAgreement_028(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 28) }
func TestJSXAgreement_029(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 29) }
func TestJSXAgreement_030(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 30) }
func TestJSXAgreement_031(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 31) }

// Drive each actual top-level wrapper alone; exactly one owns the planted case.
func TestJSXAgreementPlantedDisagreement(test *testing.T) {
	test.Parallel()
	ids := jsxAgreementIDs()
	if _, miscPlanError := miscPlan(ids, testJSXAgreementShards); miscPlanError != nil {
		test.Fatal(miscPlanError)
	}
	expected := miscShard(ids[len(ids)-1], testJSXAgreementShards)
	failures := 0
	for shard := range testJSXAgreementShards {
		name := fmt.Sprintf("TestJSXAgreement_%03d", shard)
		timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := threePortCommand(timeoutContext, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") &&
				!strings.HasPrefix(entry, "ADAMIC_ESTREE_TOP_PROOF=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "ADAMIC_ESTREE_TOP_PROOF=TestJSXAgreement")
		// Read through childguard.Capture for threePortCommand's reason, stdout then stderr.
		stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
		output := append(stdout, stderr...)
		deadlineError := timeoutContext.Err()
		cancel()
		if deadlineError != nil {
			test.Fatalf("%s proof exceeded 90s: %v", name, deadlineError)
		}
		if captureError != nil {
			failures++
			exit, ok := captureError.(*exec.ExitError)
			if !ok ||
				exit.ExitCode() != 1 ||
				shard != expected ||
				!strings.Contains(string(output), "--- FAIL: "+name+" ") {
				test.Fatalf("unexpected failure in %s: %v\n%s", name, captureError, output)
			}
		}
	}
	if failures != 1 {
		test.Fatalf("planted disagreement caught by %d shards, want one", failures)
	}
	test.Logf("planted case %s caught only by TestJSXAgreement_%03d", ids[len(ids)-1], expected)
}
