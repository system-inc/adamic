package reference_test

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestIdentifiersNamedIsTheWholeFileWalk holds the index to the walk it replaced in prefer-const: for
// every name in the file, the identifiers a ForEachChild walk from the source file reaches with that
// text, in the order it reaches them. Order matters, since a caller takes the first write it finds.
func TestIdentifiersNamedIsTheWholeFileWalk(t *testing.T) {
	t.Parallel()

	source := `let a = 1, b;
function f(a) { b = a; return () => { let a; a = b; return a; }; }
class C { a = 0; m() { return this.a + a; } }
const { a: renamed, b: [x = a] = [] } = { a, b };
for (let a of [b]) { a; }
label: while (a) { break label; }
type T = { a: typeof a };
const jsx = <a.b a={a}>{a}</a.b>;
`
	var mismatches []string
	compared := 0
	rule_testing.Run(t, rule.Rule{
		Name: "identifiers-probe",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					walked := map[string][]*ast.Node{}
					var visit func(*ast.Node)
					visit = func(current *ast.Node) {
						if current.Kind == ast.KindIdentifier {
							walked[current.Text()] = append(walked[current.Text()], current)
						}
						current.ForEachChild(func(child *ast.Node) bool {
							visit(child)
							return false
						})
					}
					visit(node)
					for name, want := range walked {
						compared++
						got := reference.IdentifiersNamed(ctx, name)
						if !slices.Equal(got, want) {
							mismatches = append(mismatches, name)
						}
						// The lists share one backing slice, so each is capped at its length: an
						// append must copy rather than write over the next name's identifiers.
						if cap(got) != len(got) {
							mismatches = append(mismatches, name+" (uncapped)")
						}
					}
					if got := reference.IdentifiersNamed(ctx, "absent"); got != nil {
						mismatches = append(mismatches, "absent")
					}
					if fills := ctx.FileCache.Fills()["reference.identifiersByName"]; fills != 1 {
						t.Errorf("the index was built %d times for one file, want once", fills)
					}
				},
			}
		},
	}, "probe.tsx", source)

	if compared < 10 {
		t.Fatalf("compared only %d names, so this proved little", compared)
	}
	if len(mismatches) > 0 {
		t.Errorf("the index disagrees with the whole-file walk for %v", mismatches)
	}
}
