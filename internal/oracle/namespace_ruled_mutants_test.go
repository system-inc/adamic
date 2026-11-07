package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func TestNamespaceRuledMutants(test *testing.T) {
	test.Parallel()
	for _, family := range []string{"live-write", "repeat-initializer", "primitive-tag"} {
		test.Run(family, func(test *testing.T) {
			test.Parallel()
			fixture := "stage3/namespace-live-export/live.a"
			if family == "repeat-initializer" {
				fixture = "internal/oracle/testdata/namespaces_repeated_var.a"
			}
			if family == "primitive-tag" {
				fixture = "internal/oracle/testdata/namespaces_observed_narrowing.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
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
				test.Fatal("mutant changed nothing")
			}
			truth := onNode(test, path)
			if disagreement(truth, onJavaScriptBackend(test, program)) == "" {
				test.Fatal("mutant survived JavaScript differential")
			}
			observed, _ := nativelyUncached(test, program)
			if disagreement(truth, observed) == "" {
				test.Fatal("mutant survived sanitized native differential")
			}
			test.Log("caught by Node in JavaScript and sanitized native backends")
		})
	}
}
