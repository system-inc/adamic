package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func TestNamespaceRuledMutants(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"live-write", "repeat-initializer", "primitive-tag"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "stage3/namespace-live-export/live.a"
			if family == "repeat-initializer" {
				fixture = "internal/oracle/testdata/namespaces_repeated_var.a"
			}
			if family == "primitive-tag" {
				fixture = "internal/oracle/testdata/namespaces_observed_narrowing.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if family == "primitive-tag" {
				for i := range program.Functions {
					if !program.Functions[i].CheckedUnionNarrow {
						continue
					}
					for j, statement := range program.Functions[i].Body {
						if guard, ok := statement.(ir.If); ok {
							guard.Condition = ir.BooleanConstant{Value: true}
							program.Functions[i].Body[j] = guard
							changed = true
						}
					}
				}
			} else {
				name := "isDebugging"
				if family == "repeat-initializer" {
					name = "value"
				}
				for i, statement := range program.Main {
					if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == name {

						program.Main[i] = ir.Evaluate{Value: assign.Value}
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			if disagreement(truth, onJavaScriptBackend(t, program)) == "" {
				t.Fatal("mutant survived JavaScript differential")
			}
			observed, _ := nativelyUncached(t, program)
			if disagreement(truth, observed) == "" {
				t.Fatal("mutant survived sanitized native differential")
			}
			t.Log("caught by Node in JavaScript and sanitized native backends")
		})
	}
}
