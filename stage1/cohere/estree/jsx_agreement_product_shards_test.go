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
)

const testJSXAgreementShards = 32

func jsxAgreementIDs() []string { return miscIDs("stage1/cohere/estree/jsx_test.go:jsx", jsxCases()) }

// The original selector checks the live union; actual case work is in top-level shards.
func TestJSXAgreement(t *testing.T) {
	t.Parallel()
	estreeTopUnion(t, "TestJSXAgreement", testJSXAgreementShards, jsxAgreementIDs())
}

func jsxAgreementShard(t *testing.T, shard int) {
	t.Helper()
	cases, ids := jsxCases(), jsxAgreementIDs()
	indexes := estreeScopedIndexes(t, testJSXAgreementShards, ids, shard)
	if estreeScopedProof(t, "TestJSXAgreement", ids, indexes) || len(indexes) == 0 {
		return
	}
	// scalarEdgePrepare uses sync.Once and the declared shared Product recipes.
	// Setup has no deadline; only this shard's case work starts the clock below.
	ready := scalarEdgePrepare(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	main := filepath.Join(root(t), "stage1/cohere/estree/main.ts")
	for _, i := range indexes {
		list := jsxManifestCases(t, cases[i:i+1])
		want := scalarEdgeExecute(t, ctx, ready.Oracle, "--manifest", list)
		for name, got := range map[string][]byte{
			"Node":       scalarEdgeNode(t, ctx, main, "--manifest", list),
			"native":     scalarEdgeExecute(t, ctx, ready.Binary, "--manifest", list),
			"emitted JS": scalarEdgeNode(t, ctx, ready.Script, "--manifest", list),
		} {
			if err := miscCompare(want, got, false); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	}
}

func TestJSXAgreement_000(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 0) }
func TestJSXAgreement_001(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 1) }
func TestJSXAgreement_002(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 2) }
func TestJSXAgreement_003(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 3) }
func TestJSXAgreement_004(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 4) }
func TestJSXAgreement_005(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 5) }
func TestJSXAgreement_006(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 6) }
func TestJSXAgreement_007(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 7) }
func TestJSXAgreement_008(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 8) }
func TestJSXAgreement_009(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 9) }
func TestJSXAgreement_010(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 10) }
func TestJSXAgreement_011(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 11) }
func TestJSXAgreement_012(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 12) }
func TestJSXAgreement_013(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 13) }
func TestJSXAgreement_014(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 14) }
func TestJSXAgreement_015(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 15) }
func TestJSXAgreement_016(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 16) }
func TestJSXAgreement_017(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 17) }
func TestJSXAgreement_018(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 18) }
func TestJSXAgreement_019(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 19) }
func TestJSXAgreement_020(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 20) }
func TestJSXAgreement_021(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 21) }
func TestJSXAgreement_022(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 22) }
func TestJSXAgreement_023(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 23) }
func TestJSXAgreement_024(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 24) }
func TestJSXAgreement_025(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 25) }
func TestJSXAgreement_026(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 26) }
func TestJSXAgreement_027(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 27) }
func TestJSXAgreement_028(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 28) }
func TestJSXAgreement_029(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 29) }
func TestJSXAgreement_030(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 30) }
func TestJSXAgreement_031(t *testing.T) { t.Parallel(); jsxAgreementShard(t, 31) }

// Drive each actual top-level wrapper alone; exactly one owns the planted case.
func TestJSXAgreementPlantedDisagreement(t *testing.T) {
	t.Parallel()
	ids := jsxAgreementIDs()
	if _, err := miscPlan(ids, testJSXAgreementShards); err != nil {
		t.Fatal(err)
	}
	expected := miscShard(ids[len(ids)-1], testJSXAgreementShards)
	failures := 0
	for shard := range testJSXAgreementShards {
		name := fmt.Sprintf("TestJSXAgreement_%03d", shard)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := threePortCommand(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_ESTREE_TOP_PROOF=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "ADAMIC_ESTREE_TOP_PROOF=TestJSXAgreement")
		output, err := command.CombinedOutput()
		deadlineErr := ctx.Err()
		cancel()
		if deadlineErr != nil {
			t.Fatalf("%s proof exceeded 90s: %v", name, deadlineErr)
		}
		if err != nil {
			failures++
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || shard != expected || !strings.Contains(string(output), "--- FAIL: "+name+" ") {
				t.Fatalf("unexpected failure in %s: %v\n%s", name, err, output)
			}
		}
	}
	if failures != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want one", failures)
	}
	t.Logf("planted case %s caught only by TestJSXAgreement_%03d", ids[len(ids)-1], expected)
}
