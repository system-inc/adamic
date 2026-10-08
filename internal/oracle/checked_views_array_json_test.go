package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewJSONArrayDomainMutants(t *testing.T) {
	for _, family := range []string{"finite", "string-finite", "keys", "nested", "maps", "functions"} {
		t.Run(family, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/entry-array-json-"+family)
			truth := onNode(t, path)
			fake := -1
			for i, local := range program.Locals {
				if local.Name == "fake" {
					fake = i
				}
			}
			if family != "nested" && fake < 0 {
				t.Fatal("missing fake")
			}
			changed := false
			for i, statement := range program.Main {
				declaration, ok := statement.(ir.Declare)
				if !ok || program.Locals[declaration.Local].Name != "values" {
					continue
				}
				array, ok := declaration.Value.(ir.ArrayLiteral)
				if !ok {
					t.Fatal("missing literal")
				}
				if family == "nested" {
					inner, ok := array.Elements[0].(ir.ArrayLiteral)
					if !ok {
						t.Fatal("missing nested literal")
					}
					inner.Elements[0] = ir.NumberConstant{Value: 2}
					array.Elements[0] = inner
				} else {
					array.Elements[0] = ir.Read{Local: fake, Of: program.Locals[fake].Type}
				}
				declaration.Value = array
				program.Main[i] = declaration
				changed = true
			}
			if !changed {
				t.Fatal("mutation missed")
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "JSON element") {
					t.Fatalf("JSON domain mutant ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; excluded value caught during JSON extraction", truth.stdout)
		})
	}
}
