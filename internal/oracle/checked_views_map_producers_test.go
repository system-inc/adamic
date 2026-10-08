package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewMapCallableProducerMutants(t *testing.T) {
	for _, site := range []string{"constructor", "set"} {
		t.Run(site, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-callable-producer")
			truth := onNode(t, path)
			original, _ := nativelyUncached(t, program)
			if diff := disagreement(truth, original); diff != "" {
				t.Fatal(diff)
			}
			wrong := -1
			for index, local := range program.Locals {
				if local.Name == "wrong" {
					wrong = index
				}
			}
			if wrong < 0 {
				t.Fatal("missing wrong producer")
			}
			changed := false
			for index, statement := range program.Main {
				if site == "constructor" {
					if declaration, ok := statement.(ir.Declare); ok {
						if creation, ok := declaration.Value.(ir.MapNew); ok && len(creation.Entries) > 0 {
							creation.Entries[0][1] = ir.Read{Local: wrong, Of: ir.Closure}
							declaration.Value = creation
							program.Main[index] = declaration
							changed = true
							break
						}
					}
				} else if evaluation, ok := statement.(ir.Evaluate); ok {
					if store, ok := evaluation.Value.(ir.MapSet); ok {
						store.Value = ir.Read{Local: wrong, Of: ir.Closure}
						evaluation.Value = store
						program.Main[index] = evaluation
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("producer mutation missed storage site")
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "incompatible result representation") {
					t.Fatalf("wrong physical producer ran on: %#v", got)
				}
			}
			t.Logf("Node control stdout=%q; mismatched actual producer stopped at %s", truth.stdout, site)
		})
	}
}
