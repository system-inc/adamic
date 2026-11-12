package javascript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The actual constructor supplies both the prototype and the inherited name.
// A later assignment must create an enumerable own name, as Node does.
func TestNamedErrorConstructorsKeepTheirPrototypes(t *testing.T) {
	t.Parallel()
	for _, class := range []string{"TypeError", "RangeError", "ReferenceError", "SyntaxError"} {
		t.Run(class, func(t *testing.T) {
			t.Parallel()
			e := emitter{program: &ir.Program{Strings: []string{"x", class}}}
			made := e.valueRaw(ir.MakeError{Message: ir.StringConstant{Index: 0}, Name: ir.StringConstant{Index: 1}})
			source := fmt.Sprintf("const error = %s;\nconsole.log(error instanceof %s);\nconsole.log(Object.hasOwn(error, 'name'));\nconsole.log(JSON.stringify(error));\nerror.name = 'Changed';\nconsole.log(JSON.stringify(error));\n", made, class)
			runViewNode(t, source, "true\nfalse\n{}\n{\"name\":\"Changed\"}\n", "", 0)
		})
	}
}

// Lowering only produces four constant names. A new producer must extend the
// constructor contract explicitly, rather than silently making name enumerable.
func TestNamedErrorRejectsAnUnprovenConstructor(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		label string
		name  ir.Expression
		text  string
	}{
		{"runtime", ir.Read{Local: 0, Of: ir.String}, "TypeError"},
		{"custom", ir.StringConstant{Index: 1}, "CustomError"},
		{"eval", ir.StringConstant{Index: 1}, "EvalError"},
		{"uri", ir.StringConstant{Index: 1}, "URIError"},
	} {
		t.Run(probe.label, func(t *testing.T) {
			t.Parallel()
			defer func() {
				failure := recover()
				if failure == nil || !strings.HasPrefix(fmt.Sprint(failure), "compiler bug: ") {
					t.Fatalf("want a compiler bug for an unproven constructor, got %v", failure)
				}
			}()
			e := emitter{program: &ir.Program{Strings: []string{"x", probe.text}}}
			e.valueRaw(ir.MakeError{Message: ir.StringConstant{Index: 0}, Name: probe.name})
		})
	}
}
