//go:build census_overlay

package latent_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestLatentScopedTailEntryFindingLoom03dg(t *testing.T) {
	t.Parallel()
	if os.Getenv("LATENT_TAIL_ENTRY_GUARD_CHILD") != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLatentScopedTailEntryFindingLoom03dg$", "-test.timeout=19s")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "LATENT_TAIL=") && !strings.HasPrefix(value, "LATENT_FULL=") && !strings.HasPrefix(value, "LATENT_SPECULATIVE=") && !strings.HasPrefix(value, "LATENT_TAIL_ENTRY_GUARD_CHILD=") {
				child.Env = append(child.Env, value)
			}
		}
		child.Env = append(child.Env, "LATENT_TAIL=1", "LATENT_FULL=0", "LATENT_SPECULATIVE=0", "LATENT_TAIL_ENTRY_GUARD_CHILD=1")
		output, err := child.CombinedOutput()
		if err != nil {
			t.Fatalf("entry guard child exit: %v\n%s", err, output)
		}
		return
	}
	const stopPass = "checkViewOptionalWrites"
	const stopReason = "a write to optional field name through a checked view, whose slot a fixed shape can't add"
	path := filepath.Join(t.TempDir(), "main.ts")
	const source = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; }
function probe(n: Node): void { const e = n as Element; e.name = "w"; }
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.LatentLoad([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := program.LatentDiagnostics(); len(diagnostics) > 0 {
		t.Fatal(diagnostics)
	}
	var records []lower.LatentFile
	lower.Latent(context.Background(), program, func(record lower.LatentFile) { records = append(records, record) })
	var owner string
	found := false
	for _, record := range records {
		for _, unit := range record.Units {
			if unit.Name == "probe" {
				owner = unit.Where
				if unit.Status != "attempted" {
					t.Fatalf("tail changed body status: %+v", unit)
				}
				stoppedInLedger := false
				for _, step := range unit.PassLedger {
					if step.Name == stopPass {
						stoppedInLedger = step.Status == "stopped_scoped"
					}
				}
				if !stoppedInLedger {
					t.Fatalf("probe ledger lacks the %s stop: %+v", stopPass, unit.PassLedger)
				}
			}
		}
		for _, finding := range record.Findings {
			if finding.Phase == "tail" && finding.Pass == stopPass && finding.Kind == "NotYet" && finding.Reason == stopReason {
				t.Logf("scoped tail stop: %+v", finding)
				found = true
				if finding.Unit != owner {
					t.Fatalf("tail owner: %+v", finding)
				}
			}
			if finding.Phase == "lowering" || finding.Phase == "deferred" || finding.Kind == "panic" {
				t.Fatalf("unexpected body stop or panic: %+v", finding)
			}
		}
	}
	if owner == "" || !found {
		t.Fatalf("no scoped tail finding for probe: %+v", records)
	}
}
