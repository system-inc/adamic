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
	path := filepath.Join(t.TempDir(), "main.ts")
	const source = `const counts: Record<string, number> = {};
counts["a"] = 1;
interface A { readonly kind: 1; readonly label: string | number }
interface B { readonly kind: 2; readonly other: string }
function probe(n: A | B): void { const s = n as A; console.log(String(s.label)); }
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
			}
		}
		for _, finding := range record.Findings {
			if finding.Phase == "tail" && finding.Pass == "checkRecordGuardedReads" && finding.Reason == "a union or optional field read in a program with record storage" {
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
