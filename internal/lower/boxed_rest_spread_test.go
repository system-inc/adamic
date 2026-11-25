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
		{"x10", "a call spreading elements that cannot be packed", false},
		{"x11", "a call spreading elements that cannot be packed", false},
		{"x12", "a call spreading elements that cannot be packed", false},
		{"x13", "a call spreading elements that cannot be packed", false},
		{"23-subclass", "a value of type Derived, a type with an index signature that isn't a string record", false},
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

func TestBoxedRestSpreadKeepsBrandedStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"b01.ts", "b02.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "boxed_rest_spread", name))
			if err != nil {
				t.Fatal(err)
			}
			if name == "b02.ts" {
				path := writeAgreementSource(t, "main.ts", string(source))
				_, err = load.Load([]string{path})
				const want = "error TS2741: Property '\" __sortedArrayBrand\"' is missing in type '(string | number)[]' but required in type 'SortedArray<string | number>'."
				if err == nil || !strings.Contains(err.Error(), "main.ts:8:8: "+want) {
					t.Fatalf("got %v, want TS2741 at the branded rest spread", err)
				}
				return
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != "a call spreading elements that cannot be packed" || !strings.HasSuffix(gap.Where, "main.ts:8:8") {
				t.Fatalf("got %v, want the existing NotYet at the branded operand", err)
			}
		})
	}
}

func TestBoxedRestSpreadRejectsBrandedRestLayout(t *testing.T) {
	t.Parallel()
	// The checker rejects a plain array at the branded rest call. Resolve
	// the signature through any and give the helper the typed union directly.
	source := brandMintPrelude + `declare const spread: any;
function target(...rest: SortedArray<number | string>): void {}
target(...spread);
`
	path := writeAgreementSource(t, "main.ts", source)
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	lower := lowering{program: program, checker: typeChecker, result: &ir.Program{
		Functions: []ir.Function{{Parameters: []int{0}, RestElement: ir.Union}},
	}}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind != ast.KindExpressionStatement {
			continue
		}
		call := statement.AsExpressionStatement().Expression.AsCallExpression()
		signature := typeChecker.GetResolvedSignature(call.AsNode())
		if signature == nil || !signature.HasRestParameter() || signature.Declaration().Body() == nil {
			t.Fatal("branded rest signature did not resolve to the declaration")
		}
		rest := typeChecker.GetTypeOfSymbol(signature.Parameters()[0])
		element := lower.elementTypeOfArray(rest)
		if !lower.isArrayType(rest) || lower.unbranded(rest) == rest || element == nil {
			t.Fatal("rest is missing the phantom array brand")
		}
		if held, known := lower.representation(element); !known || held != ir.Union {
			t.Fatal("branded array does not hold a boxed union")
		}
		if lower.boxedRestSpread(call, 0, 0, element) {
			t.Fatal("spread into a phantom branded rest was admitted")
		}
		return
	}
	t.Fatal("missing probe call")
}

func TestBoxedRestSpreadDefersToASetSlot(t *testing.T) {
	t.Parallel()
	source := `type A = number | string;
const arr: A[] = [1, "two"];
function target(first: A, ...rest: A[]): void {}
target(0, ...arr);
`
	path := writeAgreementSource(t, "main.ts", source)
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	for _, statement := range file.Statements.Nodes {
		if statement.Kind != ast.KindExpressionStatement {
			continue
		}
		call := statement.AsExpressionStatement().Expression.AsCallExpression()
		spread := call.Arguments.Nodes[1]
		operand := spread.AsSpreadElement().Expression
		array := typeChecker.GetTypeAtLocation(operand)
		check := func(node *ast.Node) bool {
			lower := lowering{program: program, checker: typeChecker, result: &ir.Program{
				Functions: []ir.Function{{Parameters: []int{0, 1}, RestElement: ir.Union}},
			}}
			if node != nil {
				lower.keepSlot(node, array)
			}
			return lower.boxedRestSpread(call, 0, 1, lower.elementTypeOfArray(array))
		}
		if !check(nil) {
			t.Fatal("spread with no caller slot was refused")
		}
		if check(spread) {
			t.Fatal("spread with a caller slot on its argument was admitted")
		}
		if check(operand) {
			t.Fatal("spread with a caller slot on its operand was admitted")
		}
		return
	}
	t.Fatal("missing probe call")
}
