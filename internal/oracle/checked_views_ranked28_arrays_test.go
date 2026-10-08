package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewRanked28OriginalArrays(t *testing.T) { originalRankedArrayOracle(t, "28", 2, 3) }

// Not parallel: refreshing writes this group's measured rows.
func TestCheckedViewRanked28ArrayCounts(t *testing.T) { originalRankedArrayCountTest(t, "28") }
func TestCheckedViewRanked28ArrayMutants(t *testing.T) {
	declarations, directory, probes := originalArrayInputs(t, "28")
	if declarations == "" {
		t.Skip("original declaration inputs required")
	}
	for _, probe := range probes {
		if !strings.HasSuffix(probe.Name, "wrong-element") {
			continue
		}
		t.Run(probe.Name, func(t *testing.T) {
			file := originalArrayFile(t, declarations, directory, probe.Name)
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			node := run{stdout: []byte(probe.Source)}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal(diff)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
			id := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
			changed := 0
			rewrite := func(n any) any {
				if read, ok := n.(ir.ArrayIndex); ok && read.View == "items[0]" {
					read.Element = ir.Number
					read.ViewContract = id
					read.ViewType = "number"
					changed++
					return read
				}
				return n
			}
			for i := range program.Functions {
				program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
			}
			if changed == 0 {
				t.Fatal("mutant changed no reached element")
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatal("mutant escaped pinned stopping oracle")
				}
				if diff := disagreement(node, got); diff != "" {
					t.Errorf("mutant must finish with Node output: %s; stderr %q", diff, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("element mutant caught by pinned stopping oracle in all three modes")
		})
	}
}
