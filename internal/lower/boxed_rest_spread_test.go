package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestBoxedRestSpreadMustStartInRest(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `type A = number | string;
const arr: A[] = [1, "two"];
function target(first: A, ...rest: A[]): void {}
target(0, ...arr);
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	lower := lowering{program: program, checker: typeChecker}
	lower.result = &ir.Program{Functions: []ir.Function{{Parameters: []int{0, 1}, RestElement: ir.Union}}}
	var visit ast.Visitor
	checked := false
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			operand := call.Arguments.Nodes[1].AsSpreadElement().Expression
			element := lower.elementTypeOfArray(typeChecker.GetTypeAtLocation(operand))
			if !lower.boxedRestSpread(call, 0, 1, element) {
				t.Fatal("spread after the fixed position should be admitted")
			}
			call.Arguments.Nodes[0], call.Arguments.Nodes[1] = call.Arguments.Nodes[1], call.Arguments.Nodes[0]
			defer func() {
				call.Arguments.Nodes[0], call.Arguments.Nodes[1] = call.Arguments.Nodes[1], call.Arguments.Nodes[0]
			}()
			if lower.boxedRestSpread(call, 0, 0, element) {
				t.Fatal("spread into the fixed position was admitted")
			}
			checked = true
		}
		return node.ForEachChild(visit)
	}
	file.Node.ForEachChild(visit)
	if !checked {
		t.Fatal("missing probe call")
	}
}

func TestBoxedRestSpreadKeepsExistingStops(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, reason string
		refused      bool
	}{
		{"p7-tuple", "a call spreading something other than an array", false},
		{"p8", "a call spreading elements that cannot be packed", false},
		{"23-subclass-construction", "a computed class base; name the base class directly", false},
		{"23-iterator-write", "inherited library member prototype read as an own field", true},
		{"23-own-iterator-write", "inherited library member iterator read as an own field", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "boxed_rest_spread", test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			if test.refused {
				var stop *Refused
				if !errors.As(err, &stop) || !strings.HasPrefix(stop.What, test.reason) {
					t.Fatalf("got %v, want Refused %q", err, test.reason)
				}
				return
			}
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != test.reason {
				t.Fatalf("got %v, want NotYet %q", err, test.reason)
			}
		})
	}
}

func TestBoxedRestSpreadKeepsOneOperand(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `type A = number | string;
function operand(): A[] { return [1, "two"]; }
function target(...rest: A[]): string { return String(rest.length); }
console.log(target("start", ...operand(), "end"));
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range program.Main {
		line, ok := statement.(ir.WriteLine)
		if !ok {
			continue
		}
		call, ok := line.Value.(ir.Call)
		if !ok || program.Functions[call.Function].Name != "target" ||
			len(call.Arguments) != 3 || len(call.Spread) != 3 ||
			call.Spread[0] || !call.Spread[1] || call.Spread[2] {
			t.Fatalf("expected a single spread in argument position 1, got %#v", line.Value)
		}
		spread, ok := call.Arguments[1].(ir.Call)
		if !ok || program.Functions[spread.Function].Name != "operand" {
			t.Fatalf("expected the operand call exactly at the spread, got %#v", call.Arguments[1])
		}
		return
	}
	t.Fatal("target call missing")
}
