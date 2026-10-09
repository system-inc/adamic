package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/function_values_return.a",
		"internal/oracle/testdata/function_values_boundary.a",
		"internal/oracle/testdata/function_values_options.a",
		"internal/oracle/testdata/function_values_signature.a",
		"internal/oracle/testdata/function_values_chain.a",
		"internal/oracle/testdata/function_values_diagnostic.a",
		"internal/oracle/testdata/function_values_array.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// The native callable ABI carries each union in the reference word. A raw scalar
// in that word is not an alternate representation of the same value.
// Not parallel: lowered writes the package-level sourceImports map.
func TestFunctionValueBoundaryBoxing(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/function_values_boundary.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	calls, boxedReturns := 0, 0
	mutateReadiness(program.Main, func(node any) any {
		if call, ok := node.(ir.CallClosure); ok && call.Returns == ir.String {
			calls++
			if len(call.Arguments) != 1 || call.Arguments[0].Type() != ir.Union {
				t.Errorf("call boundary argument is not boxed: %#v", call.Arguments)
			}
		}
		return node
	})
	if calls != 4 {
		t.Fatalf("want four mixed-member callable arguments, got %d", calls)
	}
	for _, function := range program.Functions {
		if function.Returns != ir.Union {
			continue
		}
		mutateReadiness(function.Body, func(node any) any {
			if returned, ok := node.(ir.Return); ok {
				if returned.Value == nil || returned.Value.Type() != ir.Union {
					t.Errorf("return boundary value is not boxed in %s: %#v", function.Name, returned.Value)
				}
				if _, boxed := returned.Value.(ir.Box); boxed {
					boxedReturns++
				}
			}
			return node
		})
	}
	if boxedReturns < 4 {
		t.Fatalf("want four mixed-member boxed returns, got %d", boxedReturns)
	}
}
