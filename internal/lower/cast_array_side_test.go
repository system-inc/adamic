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

const arrayArmMembers = "interface Circular { readonly buildOrder: readonly string[]; readonly circularDiagnostics: readonly string[] } interface Builder { readonly program: Circular } "

// arrayArmProbes are casts arrayArmCast proves, one per arm, with the narrowing .a's refusal offers for each.
var arrayArmProbes = []struct {
	name, source, narrowing string
}{
	{"array side", arrayArmMembers + "function f(order: readonly string[] | Circular): readonly string[] { return order as readonly string[]; }",
		"if (!Array.isArray(order)) panic('why')"},
	{"object side", arrayArmMembers + "function f(order: readonly string[] | Circular): Circular { return order as Circular; }",
		"if (Array.isArray(order)) panic('why')"},
	{"object side dropping undefined", arrayArmMembers + "function f(source: readonly string[] | Circular | Builder | undefined): Circular | Builder { return source as Circular | Builder; }",
		"if (source === undefined || Array.isArray(source)) panic('why')"},
	{"array side dropping undefined", arrayArmMembers + "function f(source: readonly string[] | Circular | Builder | undefined): readonly string[] { return source as readonly string[]; }",
		"if (!Array.isArray(source)) panic('why')"},
	{"array side keeping undefined", arrayArmMembers + "function f(source: readonly string[] | Circular | undefined): readonly string[] | undefined { return source as readonly string[] | undefined; }",
		"if (source !== undefined && !Array.isArray(source)) panic('why')"},
	{"operand no narrowing names", arrayArmMembers + "function g(order: readonly string[] | Circular): readonly string[] | Circular { return order; } function f(order: readonly string[] | Circular): Circular { return g(order) as Circular; }",
		"hold it as const value, then if (Array.isArray(value)) panic('why')"},
}

// A cast arrayArmCast proves lowers in .ts through the arm's helper, the checked cast oracle's
// cast_array_arm fixtures hold to Node and the lies stop. Behavior can't say which road lowered it, so the
// helper's name is asserted: a cast that lowered some other way would be a proof nobody pinned here.
func TestArrayArmCastLowersInTS(t *testing.T) {
	t.Parallel()
	for _, probe := range arrayArmProbes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			helpers := 0
			for _, function := range program.Functions {
				if function.Name == "cast_array_arm" {
					helpers++
				}
			}
			if helpers != 1 {
				t.Fatalf("got %d cast_array_arm helpers, want 1", helpers)
			}
		})
	}
}

// In .a the same casts are refused, as every checked downcast is (item 8, keptCheckDowncastRefusal), and the
// refusal names the Array.isArray narrowing .a writes instead.
func TestArrayArmCastIsRefusedInA(t *testing.T) {
	t.Parallel()
	for _, probe := range arrayArmProbes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".a")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "in .a whose checks no proof erases") || !strings.Contains(refused.Fix, "narrow it, as .a writes !: "+probe.narrowing+", ") {
				t.Fatalf("want .a's downcast refusal naming %q, got %v", probe.narrowing, err)
			}
		})
	}
}

// What arrayArmCast leaves at its stop, NotYet as before: a target narrower than the side it picks, by an
// element or by one member of several; an object member Array.isArray's domain doesn't admit (TextRange's pos and
// end, which tsc's NodeArray holds) or a tuple; and a null. Each arm row says castArrayArm's words.
func TestArrayArmCastStaysNotYet(t *testing.T) {
	t.Parallel()
	const p = "interface P { readonly p: number } "
	for _, probe := range []struct {
		name, source string
		arm          bool
	}{
		{"element downcast", p + "function f(x: P | readonly (string | number)[]): readonly string[] { return x as readonly string[]; }", false},
		{"one array member of several", p + "function f(x: P | readonly string[] | readonly number[]): readonly string[] { return x as readonly string[]; }", true},
		{"one object member of several", p + "interface Q { readonly q: string } function f(x: P | Q | readonly string[]): P { return x as P; }", true},
		{"text range", "interface TextRange { pos: number; end: number } interface NodeArray<T> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma?: boolean } function f(x: TextRange | readonly string[]): readonly string[] { return x as readonly string[]; }", true},
		{"tuple", p + "function f(x: P | readonly [string, string]): P { return x as P; }", true},
		{"null", p + "function f(x: P | readonly string[] | null): readonly string[] { return x as readonly string[]; }", true},
		{"target keeps null", p + "function f(x: P | readonly string[] | null): readonly string[] | null { return x as readonly string[] | null; }", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet, got %v", err)
			}
			if probe.arm && !strings.Contains(notYet.What, "which needs Array.isArray to tell") {
				t.Fatalf("want castArrayArm's words, got %v", err)
			}
		})
	}
}

// The container fit walk judges a cast's operand against the cast's target (expression.go), and an object-side
// arm's arrays read to it as arrays going where a plain object type is held. arrayArmOperandType exempts those
// arrays only at a cast arrayArmCast proves (#9z2ksv7): its operand fits once they're gone, and fails without the
// exemption. A cast arrayArmCast doesn't prove keeps the stop, here a TextRange arm outside Array.isArray's domain,
// whose cast stops first with castArrayArm's words at lowering (TestArrayArmCastStaysNotYet), so the walk is asked
// directly; and so does the array side, which the walk has no stop for.
func TestArrayArmOperandFitIsExemptOnlyWhereProved(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		exempt       bool
		stop         string
	}{
		{"object side proved", arrayArmMembers + "function f(order: readonly string[] | Circular): Circular { return order as Circular; }", true, ""},
		{"object side outside the domain", "interface TextRange { pos: number; end: number } interface NodeArray<T> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma?: boolean } function f(x: TextRange | readonly string[]): TextRange { return x as TextRange; }", false, shapeFitArrayStop},
		{"array side", arrayArmMembers + "function f(order: readonly string[] | Circular): readonly string[] { return order as readonly string[]; }", false, ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
			t.Cleanup(release)
			l := &lowering{program: loaded, checker: checked, result: &ir.Program{}}
			var operand *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindAsExpression {
					operand = ast.SkipParentheses(node.AsAsExpression().Expression)
				}
				node.ForEachChild(visit)
				return false
			}
			loaded.Files()[0].AsNode().ForEachChild(visit)
			if operand == nil {
				t.Fatal("no cast in the probe")
			}
			slot := l.slotType(operand)
			judged := l.arrayArmOperandType(operand)
			if exempted := judged != l.checker.GetTypeAtLocation(operand); exempted != probe.exempt {
				t.Fatalf("exempted %v (judged %s), want %v", exempted, l.checker.TypeToString(judged), probe.exempt)
			}
			if stop := l.containerWhereOtherGoes(judged, slot, false); stop != probe.stop {
				t.Fatalf("the walk says %q at the operand, want %q", stop, probe.stop)
			}
			if probe.exempt {
				if stop := l.containerWhereOtherGoes(l.checker.GetTypeAtLocation(operand), slot, false); stop != shapeFitArrayStop {
					t.Fatalf("without the exemption the walk says %q, want %q, so the exemption isn't what clears it", stop, shapeFitArrayStop)
				}
			}
		})
	}
}

// A checked view checks at a read, never at its cast (Lower's must on #9z2ksv7): the overlap view of Named over
// Other | [number, number] puts the tuple behind Named unread, and natively a tuple is an object, so the arm's
// Array.isArray would call it no array where Node calls it one. The cast records its arms where its test is built,
// and checkArrayArmViews stops it in the pass's own words, at the cast. Mutant: arrayArmValue recording nothing;
// this row lowers and goes red.
func TestArrayArmCastBesideAViewOverATupleStaysNotYet(t *testing.T) {
	t.Parallel()
	source := "interface Named { readonly name: string }\nconst pair: [number, number] = [1, 2];\ninterface Other { readonly name: string; readonly kind: 'o' }\nfunction hold(): [number, number] | Other { return pair; }\nconst u = hold();\nconst viewed = u as Named;\nfunction side(value: Named | readonly string[]): Named { return value as Named; }\nconsole.log(side(viewed).name);\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, ":7:65") || !strings.HasPrefix(notYet.What, "Array.isArray over Named, which the checked view of Named over Other | [number, number] at ") || !strings.HasSuffix(notYet.What, ":6:16 can hold an array behind, unread") {
		t.Fatalf("want checkArrayArmViews' NotYet at the cast (:7:65), got %v", err)
	}
}
