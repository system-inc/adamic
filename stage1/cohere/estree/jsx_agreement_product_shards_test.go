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

const testJsxAgreementShards = 32

func jsxAgreementIds() []string { return miscIds("stage1/cohere/estree/jsx_test.go:jsx", jsxCases()) }

// The original selector checks the live union; actual case work is in top-level shards.
func TestJSXAgreement(test *testing.T) {
	test.Parallel()
	estreeTopUnion(test, "TestJSXAgreement", testJsxAgreementShards, jsxAgreementIds())
}

func jsxAgreementShard(test *testing.T, shard int) {
	test.Helper()
	cases, ids := jsxCases(), jsxAgreementIds()
	indexes := estreeScopedIndexes(test, testJsxAgreementShards, ids, shard)
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

func TestJsxAgreement_000(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 0) }
func TestJsxAgreement_001(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 1) }
func TestJsxAgreement_002(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 2) }
func TestJsxAgreement_003(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 3) }
func TestJsxAgreement_004(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 4) }
func TestJsxAgreement_005(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 5) }
func TestJsxAgreement_006(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 6) }
func TestJsxAgreement_007(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 7) }
func TestJsxAgreement_008(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 8) }
func TestJsxAgreement_009(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 9) }
func TestJsxAgreement_010(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 10) }
func TestJsxAgreement_011(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 11) }
func TestJsxAgreement_012(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 12) }
func TestJsxAgreement_013(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 13) }
func TestJsxAgreement_014(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 14) }
func TestJsxAgreement_015(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 15) }
func TestJsxAgreement_016(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 16) }
func TestJsxAgreement_017(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 17) }
func TestJsxAgreement_018(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 18) }
func TestJsxAgreement_019(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 19) }
func TestJsxAgreement_020(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 20) }
func TestJsxAgreement_021(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 21) }
func TestJsxAgreement_022(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 22) }
func TestJsxAgreement_023(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 23) }
func TestJsxAgreement_024(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 24) }
func TestJsxAgreement_025(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 25) }
func TestJsxAgreement_026(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 26) }
func TestJsxAgreement_027(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 27) }
func TestJsxAgreement_028(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 28) }
func TestJsxAgreement_029(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 29) }
func TestJsxAgreement_030(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 30) }
func TestJsxAgreement_031(test *testing.T) { test.Parallel(); jsxAgreementShard(test, 31) }

// Drive each actual top-level wrapper alone; exactly one owns the planted case.
func TestJsxAgreementPlantedDisagreement(test *testing.T) {
	test.Parallel()
	ids := jsxAgreementIds()
	if _, miscPlanError := miscPlan(ids, testJsxAgreementShards); miscPlanError != nil {
		test.Fatal(miscPlanError)
	}
	expected := miscShard(ids[len(ids)-1], testJsxAgreementShards)
	failures := 0
	for shard := range testJsxAgreementShards {
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
