package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const representationReasonArray = "interface JSDocArray extends Array<string> { jsDocCache?: string; other?: number }"
const representationReasonWeak = "import type { Weak } from 'adamic'; interface Item { readonly label: string }"

func TestRepresentationReasonStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, prelude, parameters, typ, displayed, reason string
		direct                                            bool
	}{
		{"array_member", representationReasonArray, "", "JSDocArray", "JSDocArray", "an array type with its own member jsDocCache, held as neither an array nor an object", false},
		{"index_signature", "interface BareArray extends Array<string> {}", "", "BareArray", "BareArray", "a type with an index signature that isn't a string record", false},
		{"intersection_array", "type Target = string[] & { tag: number };", "", "Target", "Target", "an intersection whose member string[] isn't a plain object shape (an array)", false},
		{"intersection_tuple", "type Target = [string] & { tag: number };", "", "Target", "Target", "an intersection whose member [string] isn't a plain object shape (a tuple)", false},
		{"intersection_function", "type Fn = () => string; type Target = Fn & { tag: number };", "", "Target", "Target", "an intersection whose member Fn isn't a plain object shape (a function)", false},
		{"intersection_class", "class Item { label = 'x'; } type Target = Item & { tag: number };", "", "Target", "Target", "an intersection whose member Item isn't a plain object shape (a class instance)", false},
		{"intersection_dictionary", "type Dict = Record<string, number>; type Target = Dict & { tag: number };", "", "Target", "Target", "an intersection whose member Dict isn't a plain object shape (a dictionary)", false},
		{"intersection_collection", "type Target = Map<string, number> & { tag: number };", "", "Target", "Target", "an intersection whose member Map<string, number> isn't a plain object shape (a collection)", false},
		{"intersection_primitive", "type Target = symbol & { tag: number };", "", "Target", "Target", "an intersection whose member symbol isn't a plain object shape (a primitive)", false},
		{"intersection_parameter", "", "<T>", "NonNullable<T>", "NonNullable<T>", "an intersection whose member T isn't a plain object shape (a type parameter)", false},
		{"optional_boolean", "type Target = { flag?: boolean } & { flag?: boolean | string } & { tag: number };", "", "Target", "Target", "an intersection whose field flag is an optional boolean, which its record storage doesn't hold", false},
		{"function_field", "type Target = { call: () => number } & { tag: number };", "", "Target", "Target", "an intersection whose field call is a function, which its record storage doesn't hold", false},
		{"weak_field", representationReasonWeak + "; type Target = { item: Weak<Item> } & { tag: number };", "", "Target", "Target", "an intersection whose field item is a weak reference, which its record storage doesn't hold", false},
		{"unknown_field", representationReasonArray + "; type Target = { jsDoc?: JSDocArray } & { tag: number };", "", "Target", "Target", "an intersection whose field jsDoc has no representation: a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object", false},
		{"union_member", representationReasonArray + "; type Target = JSDocArray | string;", "", "Target", "Target", "a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object", false},
		{"nullable_union_member", representationReasonArray + "; type Target = JSDocArray | null;", "", "Target", "Target", "a union whose member JSDocArray has no representation: an array type with its own member jsDocCache, held as neither an array nor an object", false},
		{"weak_union", representationReasonWeak + "; type Target = Weak<Item> | number;", "", "Target", "Target", "a union of a weak reference and other members", false},
		{"phantom_union", "type Target = (void & { readonly first: void }) | (void & { readonly second: void });", "", "Target", "Target", "a union with no member that holds a value", false},
		{"type_parameter", "", "<T>", "T", "T", "a type parameter this instantiation doesn't fix", false},
		{"never", "", "", "never", "never", "never (no value has it)", false},
		{"symbol", "", "", "symbol", "symbol", "a symbol", false},
		{"bigint", "", "", "bigint", "bigint", "a bigint", false},
		{"indexed_access", "", "<T, K extends keyof T>", "T[K]", "T[K]", "an unresolved indexed access", false},
		{"conditional", "type PickValue<T> = T extends string ? string : number;", "<T>", "PickValue<T>", "PickValue<T>", "an unresolved conditional type", false},
		{"keyof", "", "<T>", "keyof T", "keyof T", "an unresolved keyof", false},
		{"fallback", "", "", "any", "any", "a type no storage is defined for", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := probe.prelude + "\nfunction probe(value: " + probe.typ + "): void {}"
			if probe.parameters != "" {
				source = probe.prelude + "\nconst probe = " + probe.parameters + "(value: " + probe.typ + "): void => {};"
			}
			l, node := representationReasonParameter(t, source)
			proven := l.checker.GetTypeAtLocation(node)
			checkRepresentationReason(t, l, proven)
			if _, known, got := l.representationReason(proven); known || got != probe.reason {
				t.Errorf("representationReason = (%t, %q), want (false, %q)", known, got, probe.reason)
			}
			// Explicit any is rejected by the existing refusal before either
			// diagnostic can use the fallback reason. Pin it directly above.
			if probe.direct {
				return
			}

			want := "stage 0 can't lower a value of type " + probe.displayed + ", " + probe.reason + " yet"
			if probe.name == "array_member" {
				_, err := l.typeOfSymbol(node, l.checker.GetSymbolAtLocation(node))
				checkRepresentationStop(t, err, want)
			}
			_, err := lowerSourceAs(t, source, ".ts")
			checkRepresentationStop(t, err, want)
		})
	}
}

func checkRepresentationStop(t *testing.T, err error, want string) {
	t.Helper()
	var stop *NotYet
	if !errors.As(err, &stop) || strings.TrimPrefix(stop.Error(), stop.Where+": ") != want {
		t.Fatalf("got %v, want NotYet %q", err, want)
	}
}

// objectIntersection's verdict on an optional boolean field (#tq6yzqy): admitted as an object where every
// member that has the field holds it as an optional boolean by the plain rule, the inherited member asked
// through the checker's property, and stopped in today's words where a member holds it otherwise.
func TestObjectIntersectionOptionalBoolean(t *testing.T) {
	t.Parallel()
	const stop = "an intersection whose field flag is an optional boolean, which its record storage doesn't hold"
	for _, probe := range []struct{ name, prelude, reason string }{
		{"declared_by_one", "type Target = { flag?: boolean } & { tag: number };", ""},
		{"agreeing", "interface Left { flag?: boolean; left: number } interface Right { flag?: boolean; right: string } type Target = Left & Right;", ""},
		{"inherited", "interface Base { flag?: boolean } interface Derived extends Base { text: string } type Target = Derived & { kind: 1 };", ""},
		{"optional_true", "type Target = { flag?: boolean } & { flag?: true } & { tag: number };", ""},
		{"union_member", "type Target = { flag?: boolean } & { flag?: boolean | string } & { tag: number };", stop},
		{"unknown_member", "type Target = { flag?: boolean } & { flag?: unknown } & { tag: number };", stop},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, node := representationReasonParameter(t, probe.prelude+"\nfunction probe(value: Target): void {}")
			proven := l.checker.GetTypeAtLocation(node)
			field := l.checker.GetPropertyOfType(proven, "flag")
			if field == nil {
				t.Fatal("Target has no property flag")
			}
			if held, known := l.representation(l.checker.GetTypeOfSymbol(field)); !known || held != ir.MaybeBoolean {
				t.Fatalf("flag is held as (%v, %t), want an optional boolean, so the pin asks the rule it means to", held, known)
			}
			of, known, reason := l.objectIntersection(proven)
			if probe.reason == "" && (!known || of != ir.Object || reason != "") {
				t.Errorf("objectIntersection = (%v, %t, %q), want an admitted object", of, known, reason)
			}
			if probe.reason != "" && (known || reason != probe.reason) {
				t.Errorf("objectIntersection = (%v, %t, %q), want the stop %q", of, known, reason, probe.reason)
			}
		})
	}
}

func TestRepresentationReasonScalarUndefined(t *testing.T) {
	t.Parallel()
	l, node := representationReasonParameter(t, "function probe<T>(value: T | undefined): void {}")
	proven := l.checker.GetTypeAtLocation(node)
	// Number and boolean have Maybe forms, and every other ordinary value is
	// a reference. A synthetic scalar substitution exercises the fallback.
	parameter := node.Parent.Parent.TypeParameters()[0].Name()
	l.substitution = map[*checker.Type]ir.Type{l.checker.GetTypeAtLocation(parameter): ir.UndefinedRepresentation}
	checkRepresentationReason(t, l, proven)
	const want = "T | undefined, which has no present-and-value pair"
	if _, known, got := l.representationReason(proven); known || got != want {
		t.Fatalf("representationReason = (%t, %q), want (false, %q)", known, got, want)
	}
}

func TestRepresentationReasonAgreement(t *testing.T) {
	t.Parallel()
	const prelude = "class Left { left = 1; } class Right { right = 2; }\n"
	for _, typ := range []string{
		"number", "string", "boolean", "void", "null", "undefined", "unknown", "object",
		"Left", "Left | Right", "Left | null", "number | undefined", "boolean | undefined",
		"number | string", "string[]", "readonly number[]", "[number, string]", "Uint8Array",
		"Map<string, number>", "Set<string>", "Promise<number>", "Record<string, number>",
		"() => void", "{ left: number } & { right: string }",
	} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			l, node := representationReasonParameter(t, prelude+"function probe(value: "+typ+"): void {}")
			checkRepresentationReason(t, l, l.checker.GetTypeAtLocation(node))
		})
	}
}

func representationReasonParameter(t *testing.T, source string) (*lowering, *ast.Node) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	t.Cleanup(release)
	probe := entry.Statements.Nodes[len(entry.Statements.Nodes)-1]
	if probe.Kind == ast.KindVariableStatement {
		probe = probe.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Initializer()
	}
	return &lowering{checker: checked, program: program, result: &ir.Program{}}, probe.Parameters()[0].Name()
}

func checkRepresentationReason(t *testing.T, l *lowering, typ *checker.Type) {
	t.Helper()
	of, known := l.representation(typ)
	got, gotKnown, reason := l.representationReason(typ)
	if got != of || gotKnown != known || (!gotKnown && reason == "") {
		t.Errorf("representationReason = (%v, %t, %q), representation = (%v, %t)", got, gotKnown, reason, of, known)
	}
}
