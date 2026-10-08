package lower

import (
	"errors"
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestFunctionValueUnionViewsStayNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function show(value: string | number): string { return `${value}`; }\nconst numeric: (value: number) => string = show;\nconsole.log(numeric(7));\n",
		"const show = (value: string | number): string => `${value}`;\nconst numeric: (value: number) => string = show;\nconsole.log(numeric(7));\n",
		"function number(): number { return 7; }\nconst mixed: () => string | number = number;\nconsole.log(`${mixed()}`);\n",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Errorf("callable view requires a union slot adapter, got %v", err)
		}
	}
}

// The native callable ABI carries each union in the reference word. A raw scalar
// in that word is not an alternate representation of the same value.
func TestFunctionValueBoundaryBoxing(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/function_values_boundary.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	calls, boxedReturns := 0, 0
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.CallClosure); ok && call.Returns == ir.String {
			calls++
			if len(call.Arguments) != 1 || call.Arguments[0].Type() != ir.Union {
				t.Errorf("call boundary argument is not boxed: %#v", call.Arguments)
			}
		}
		return true
	})
	if calls != 4 {
		t.Fatalf("want four mixed-member callable arguments, got %d", calls)
	}
	for _, function := range program.Functions {
		if function.Returns != ir.Union {
			continue
		}
		walk(function.Body, func(node any) bool {
			if returned, ok := node.(ir.Return); ok {
				if returned.Value == nil || returned.Value.Type() != ir.Union {
					t.Errorf("return boundary value is not boxed in %s: %#v", function.Name, returned.Value)
				}
				if _, boxed := returned.Value.(ir.Box); boxed {
					boxedReturns++
				}
			}
			return true
		})
	}
	if boxedReturns < 4 {
		t.Fatalf("want four mixed-member boxed returns, got %d", boxedReturns)
	}
}
