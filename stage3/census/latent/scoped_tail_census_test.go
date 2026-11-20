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

const scopedTailTestStop = "forwarderValues@probe"

func TestLatentScopedTailEntryFindingLoom03dg(t *testing.T) {
	t.Parallel()
	if os.Getenv("LATENT_TAIL_ENTRY_GUARD_CHILD") != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLatentScopedTailEntryFindingLoom03dg$", "-test.timeout=19s")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "LATENT_TAIL=") && !strings.HasPrefix(value, "LATENT_FULL=") && !strings.HasPrefix(value, "LATENT_SPECULATIVE=") && !strings.HasPrefix(value, "LATENT_TAIL_ENTRY_GUARD_CHILD=") && !strings.HasPrefix(value, "LATENT_TAIL_TEST_STOP=") {
				child.Env = append(child.Env, value)
			}
		}
		child.Env = append(child.Env, "LATENT_TAIL=1", "LATENT_FULL=0", "LATENT_SPECULATIVE=0", "LATENT_TAIL_ENTRY_GUARD_CHILD=1", "LATENT_TAIL_TEST_STOP="+scopedTailTestStop)
		output, err := child.CombinedOutput()
		if err != nil {
			t.Fatalf("entry guard child exit: %v\n%s", err, output)
		}
		return
	}
	// The stop is the census's own injected one (scoped_tail.go.txt's latentTailTestStop), never a production refusal:
	// what this pins is how the tail records a scoped stop, and two probes pinned to production passes stopped
	// stopping as those passes changed (#hcxsh1k). forwarderValues is the first scoped step and always executes.
	const stopPass = "forwarderValues"
	const stopReason = "a stop the census test injects"
	path := filepath.Join(t.TempDir(), "main.ts")
	const source = "export function probe(): number { return 1; }\n"
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
				after := false
				for _, step := range unit.PassLedger {
					if step.Name == stopPass {
						stoppedInLedger = step.Status == "stopped_scoped" && strings.Contains(step.Reason, stopReason)
						// Later steps are judged only once the stop is recorded, so a stop that never happened fails
						// as missing, below, rather than as a later step running.
						after = stoppedInLedger
						continue
					}
					if after && (step.Status == "executed_scoped" || step.Status == "stopped_scoped") {
						t.Fatalf("a step after the stop ran: %+v", step)
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
