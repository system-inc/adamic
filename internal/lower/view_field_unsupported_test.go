package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A field whose own contract can't be built is unsupported alone (#rgrtt32): tsc's Node graph always
// has one, and a view's whole contract used to become the representation-conversion descriptor for
// it, so a Node-typed view reached only its root. Two roads lead there today, each measured on
// Node-like shapes: an optional Map, whose Map member has no union member contract, and a variable
// tuple. Each shape gives the unsupported field a readonly scalar sibling, size, beside its tag.
// Line 6 of each source is its use, after the view is made on line 5.
var unbuiltFieldShapes = []struct{ name, field, declared, read string }{
	{"optional map", "locals", "locals?: Map<string, Nd>", "console.log(String(view.locals === undefined));"},
	{"variable tuple", "pair", "pair: readonly [number, ...string[]]", "console.log(String(view.pair[0]));"},
}

// unbuiltFieldSource is a shape's program: raw holds its tag and rawFields, and use is line 6.
func unbuiltFieldSource(declared, rawFields, use string) string {
	return `interface Base { readonly kind: 'box' | 'other' }
interface Nd extends Base { readonly kind: 'box'; readonly size: number; ` + declared + ` }
const raw = { kind: 'box' as const, ` + rawFields + ` };
const held: Base = raw;
const view = held as Nd;
` + use + `
`
}

// The sibling's read is checked and agrees with Node, on the JavaScript backend and natively: the
// view's contract keeps its fields, so size is read through the view, not refused with the whole
// type. Under the old whole-type descriptor this read stopped at lowering.
func TestUnbuiltFieldSiblingReadIsCheckedAndAgrees(t *testing.T) {
	t.Parallel()
	for _, shape := range unbuiltFieldShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, unbuiltFieldSource(shape.declared, "size: 7", "console.log(String(view.size));"), ".ts")
			accessorReachNative(t, program, "7\n")
			checked := 0
			walk(program.Main, func(node any) bool {
				if read, ok := node.(ir.Property); ok && read.Name == "size" && read.View != "" && program.CheckedRead(read) {
					checked++
				}
				return true
			})
			if checked != 1 {
				t.Fatalf("want size read once through the checked view, got %d", checked)
			}
		})
	}
}

// A read of the unsupported field is NotYet where it's made, in words naming the field and why its
// contract isn't built, never read raw. Mutant: the judge passing the field-level family, and the
// read reaches emission's own rule, whose words are another's.
func TestUnbuiltFieldReadIsNotYet(t *testing.T) {
	t.Parallel()
	for _, shape := range unbuiltFieldShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, unbuiltFieldSource(shape.declared, "size: 7", shape.read), ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, "main.ts:6:20") || !strings.Contains(notYet.What, "field "+shape.field+" whose contract isn't built") {
				t.Fatalf("want a NotYet at main.ts:6:20 naming field %s, got %v", shape.field, err)
			}
		})
	}
}

// A lie in the sibling, a string where the view says number, stops loudly at its read on both
// backends, as any checked read does, naming the read and what it found. A release native build's
// panic carries no source position, so the read is named by its expression, as the oracle's checked
// view pins name theirs.
func TestUnbuiltFieldSiblingLieStopsAtTheRead(t *testing.T) {
	t.Parallel()
	for _, shape := range unbuiltFieldShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			overloadCheckStops(t, unbuiltFieldSource(shape.declared, "size: 'seven'", "console.log(String(view.size));"), "", "field read failed: view.size is not a number", "expected number, found string")
		})
	}
}

// Whole-value readers keep their stop: a spread of the viewed value, which would copy every field
// unchecked, the unsupported one among them.
func TestUnbuiltFieldSpreadStops(t *testing.T) {
	t.Parallel()
	for _, shape := range unbuiltFieldShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, unbuiltFieldSource(shape.declared, "size: 7", "const copy = { ...view }; console.log(String(copy.size));"), ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, "main.ts:6:16") {
				t.Fatalf("want a NotYet at the spread, main.ts:6:16, got %v", err)
			}
		})
	}
}

// A write through the view to the unsupported field is NotYet where it's written, naming the field
// and why, wherever its receiver may hold a viewed allocation, as its read is: no contract says what
// the slot may take. The optional Map's raw holds a locals slot, so the absent-optional write rule
// (checkSlotWrites) lets it through to this one. Where raw was built without locals, that older rule
// answers first, at the same place, in words of its own. A write to the tuple whose program never
// reads the view again lowered before this rule and trapped at run time on every backend, "<write>.pair
// is not initialized", where Node prints done (J-vktzq8t). Two writes stop today at older rules that
// answer first, each held only to a NotYet on the write's line, so each keeps a stop when that rule lifts
// (Language, J-3z6n90v): a destructuring assignment into the field, at "destructuring a value made in the
// assignment itself", and Object.assign onto the view, at "Object.assign on a shape not proven by a plain
// literal". Mutant: the write stop dropped, and the tuple's writes lower.
func TestUnbuiltFieldWriteIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, declared, rawFields, write, at, field string }{
		{"optional map", "locals?: Map<string, Nd>", "size: 7, locals: new Map<string, Nd>()", "view.locals = new Map(); console.log(String(view.size));", "6:1", "locals"},
		{"optional map without its slot", "locals?: Map<string, Nd>", "size: 7", "view.locals = new Map(); console.log(String(view.size));", "6:1", ""},
		{"optional map destructured", "locals?: Map<string, Nd>", "size: 7, locals: new Map<string, Nd>()", "({ locals: view.locals } = { locals: new Map<string, Nd>() }); console.log(String(view.size));", "6", ""},
		{"optional map assigned onto", "locals?: Map<string, Nd>", "size: 7, locals: new Map<string, Nd>()", "Object.assign(view, { locals: new Map<string, Nd>() }); console.log(String(view.size));", "6", ""},
		{"variable tuple", "pair: readonly [number, ...string[]]", "size: 7", "view.pair = [1, 'a']; console.log(String(view.size));", "6:1", "pair"},
		{"variable tuple then done", "pair: readonly [number, ...string[]]", "size: 7", "view.pair = [1, 'a']; console.log('done');", "6:1", "pair"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, unbuiltFieldSource(probe.declared, probe.rawFields, probe.write), ".ts")
			// at is a line and column, held exactly, or a line alone.
			where := "main.ts:" + probe.at
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet at %s, got %v", where, err)
			}
			placed := strings.HasSuffix(notYet.Where, where) || !strings.Contains(probe.at, ":") && strings.Contains(notYet.Where, where+":")
			if !placed || probe.field != "" && !strings.Contains(notYet.What, "field "+probe.field+" whose contract isn't built") {
				t.Fatalf("want a NotYet at %s naming field %q, got %v", where, probe.field, err)
			}
		})
	}
}

// The contract itself: the view's object keeps its fields, the unsupported one's descriptor names
// why without a place, and a type whose own representation is unknown stays whole.
func TestUnbuiltFieldContractIsTheFieldsAlone(t *testing.T) {
	t.Parallel()
	for _, shape := range unbuiltFieldShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, unbuiltFieldSource(shape.declared, "size: 7", "console.log(String(view.size));"), ".ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, contract := range program.ViewContracts {
				if contract.Name != "Nd" {
					continue
				}
				fields := map[string]ir.ViewContract{}
				for _, field := range contract.Fields {
					fields[field.Name] = program.ViewContracts[field.Contract-1]
				}
				unbuilt := fields[shape.field]
				if contract.Unsupported != "" || contract.Kind != ir.ViewObject || fields["size"].Kind != ir.ViewScalar || !strings.HasPrefix(unbuilt.Unsupported, unbuiltFieldFamily) || strings.Contains(unbuilt.Unsupported, "main.ts") {
					t.Fatalf("Nd: %+v, its %s %+v", contract, shape.field, unbuilt)
				}
				return
			}
			t.Fatal("no contract for Nd")
		})
	}
}
