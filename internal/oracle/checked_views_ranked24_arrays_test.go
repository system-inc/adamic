package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewRanked24OriginalArrays(t *testing.T) { originalRankedArrayOracle(t, "24", 3, 3) }

// Not parallel: refreshing writes this group's measured rows.
func TestCheckedViewRanked24ArrayCounts(t *testing.T) { originalRankedArrayCountTest(t, "24") }

func TestCheckedViewRanked24ArrayMutants(t *testing.T) {
	declarations, directory, probes := originalArrayInputs(t, "24")
	if declarations == "" {
		t.Skip("original declaration inputs required")
	}
	for _, probe := range probes {
		if probe.Diagnostic == "" || (strings.HasPrefix(probe.Name, "flow-") && (strings.HasSuffix(probe.Name, "wrong-flags") || strings.HasSuffix(probe.Name, "missing-flags"))) {
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
			changed := 0
			scalar := ir.Number
			if strings.HasSuffix(probe.Name, "wrong-flags") || strings.HasSuffix(probe.Name, "missing-flags") {
				scalar = ir.String
			}
			id := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: scalar, Name: "mutated scalar"})
			if strings.HasSuffix(probe.Name, "wrong-array") {
				field := strings.Split(probe.Name, "-")[0]
				if field == "flow" {
					field = "antecedents"
				}
				changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "viewed."+field }, func(p ir.Property) ir.Property {
					p.Of = ir.Number
					p.Nullish = false
					p.NullishKinds = 0
					p.NullAllowed = false
					p.UndefinedAllowed = false
					p.Optional = false
					p.ViewContract = id
					p.ViewType = "number"
					return p
				})
			} else if strings.HasSuffix(probe.Name, "wrong-element") {
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
			} else {
				changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "items[0]!.flags" }, func(p ir.Property) ir.Property {
					p.Of = ir.String
					p.ViewContract = id
					p.ViewType = "string"
					if strings.HasSuffix(probe.Name, "missing-flags") {
						p.Absent = true
						p.Optional = true
					}
					return p
				})
			}
			if changed == 0 {
				t.Fatal("mutant changed no reached read")
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatal("mutant escaped pinned stopping oracle")
				}
				if diff := disagreement(node, got); diff != "" {
					t.Errorf("mutant must finish with Node output: %s; stdout %q stderr %q", diff, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("mutant caught by pinned stopping oracle in sanitized native, release native and JavaScript")
		})
	}
}
