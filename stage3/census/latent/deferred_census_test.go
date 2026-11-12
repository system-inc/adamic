//go:build census_overlay

package latent_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The overlay build supplies the measurement APIs. The test itself never starts
// go or a compiler subprocess. Only this test uses the sequential census globals.
func TestLatentWrittenAnyDeferredCensusStopsZb0q0zd(t *testing.T) {
	t.Parallel()
	for _, generic := range []bool{true, false} {
		for _, ordinary := range []bool{false, true} {
			path := filepath.Join(t.TempDir(), "main.ts")
			source := `function writtenAny(values: any[]): string[] { return values as string[]; }
`
			unitName := "writtenAny"
			if generic {
				source = strings.Replace(source, "function writtenAny(", "function writtenAny<T>(", 1) + "function callWrittenAny(values: any[]): string[] { return writtenAny<number>(values); }\n"
				unitName = "callWrittenAny"
			}
			source += "function ledgerBodySuccessPb47(): number { return 1; }\n"
			if ordinary {
				source = strings.Replace(source, "return values as string[];", "const result = values as string[]; const box: { tag?: number } = { tag: 1 }; delete box.tag; return result;", 1)
			}
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.LatentLoad([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			if diagnostics := program.LatentDiagnostics(); len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			var records []lower.LatentFile
			lower.Latent(context.Background(), program, func(record lower.LatentFile) { records = append(records, record) })
			if len(records) != 1 {
				t.Fatalf("records: %+v", records)
			}
			record := records[0]
			for _, observed := range record.Units {
				if observed.Status == "deferred_generic_instantiation" {
					if len(observed.PassLedger) != 0 {
						t.Fatal("unattempted generic has a pass ledger")
					}
					continue
				}
				if len(observed.PassLedger) == 0 {
					t.Fatalf("attempted unit has no pass ledger: %+v", observed)
				}
				whole := 0
				for _, step := range observed.PassLedger {
					if observed.Name == "ledgerBodySuccessPb47" {
						if step.Status == "not_run_whole_program" {
							whole++
						} else if step.Status != "not_run_scoped" {
							t.Fatalf("successful body: %+v", step)
						}
					} else if step.Status != "not_reached_lowering_stop" {
						t.Fatalf("stopped body: %+v", step)
					}
				}
				if observed.Name == "ledgerBodySuccessPb47" && whole != 4 {
					t.Fatalf("whole-program omissions: %d", whole)
				}
			}

			var unit *lower.LatentUnit
			for index := range record.Units {
				if record.Units[index].Name == unitName {
					unit = &record.Units[index]
					break
				}
			}
			if unit == nil {
				t.Fatalf("writtenAny unit missing: %+v", record.Units)
			}
			stopped := false
			for _, finding := range record.Findings {
				if !ordinary && finding.Kind != "Boundary" && finding.Phase != "deferred" && finding.Phase != "deferred_scan" {
					t.Fatalf("fixture has another stop: %+v", finding)
				}
				if finding.Phase == "deferred" || finding.Phase == "deferred_scan" {
					if ordinary {
						t.Fatalf("deferred displaced an ordinary refusal: %+v", finding)
					}
					if finding.Phase == "deferred" && finding.Unit == unit.Where {
						if finding.Kind != "NotYet" || finding.Reason != "a member typed any ([]) related to string at the assertion site" {
							t.Fatalf("wrong deferred stop: %+v", finding)
						}
						stopped = true
					}
				}
			}
			if !ordinary && !stopped {
				t.Fatalf("written-any unit reads as lowering: unit=%+v findings=%+v", unit, record.Findings)
			}
			t.Logf("ordinary=%v unit=%+v findings=%+v", ordinary, unit, record.Findings)
		}
	}
}
