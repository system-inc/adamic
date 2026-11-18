package lower

import (
	"strings"
	"testing"
)

// A shorthand property's name is the property to GetSymbolAtLocation, so a walk matching names by symbol
// missed { x } reading x and ({ x } = other) writing it (#3sdynvn). Each program below lowered past its
// walk on 1d3efe7b (probe J-wj7maxs) and now meets valueSymbol's answer, stopping where its plain form,
// ({ x } = other) written x = other.x and { x } written { x: x }, stops:
//
//   - enumeration_binding: enumerationBindingWritten saw no write to held, so Object.values was read
//     unchecked and native read a string as a number (Node "3 1,2,x"). Now checked, it stops loudly.
//   - predicate_overload_binding: predicateOverloadBinding saw no nested write to node, so findIt's result
//     was believed (native "radius 2", Node "radius undefined"). Now the result is checked and stops.
//   - predicate_overload_callback: predicateOverloadAssigned saw no write to predicate, so first's result
//     was believed (native "radius 2", Node "radius undefined"). Now the overload's predicate is refused.
//   - predicate_overload_nested_pattern and predicate_overload_for_of_pattern (Language's rows):
//     predicate_overload_callback's program with predicate written a level down, ({ inner: { predicate } }
//     = nest), and as a for-of's target, for ({ predicate } of liars). Not probed on 1d3efe7b; each is held
//     to the stop its plain form makes.
//   - record_read_discarded: recordReadDiscarded saw no read of v, so dict["toString"] went unfetched
//     (native "undefined", Node "function"). Now the Object.prototype key is refused.
//   - static_class_escape: staticInitializationReads saw no { Box }, so the class escaped before its
//     static fields were set (native ReferenceError, Node "undefined"). Now it's refused there.
//   - static_outside_helper: classDeadZoneGap saw no { Box } in a helper that runs in the class's dead
//     zone. Native and Node both threw, so it was no lie, but the read is stopped now as its plain form is.
//
// Mutant: valueSymbol's shorthand arm dropped, and every row lowers past its walk again.
func TestShorthandValueSymbol(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct {
		name, source, want string
		native             bool
	}{
		{"enumeration_binding", "const wide = { a: 1, b: 2, c: \"x\" };\nconst other: { held: { a: number; b: number } } = { held: wide };\nlet held = { a: 4, b: 5 };\n({ held } = other);\nconst values = Object.values(held);\nconsole.log(`${values.length} ${values.join(\",\")}`);\n", "adamic: panic: Object enumeration key 'c': actual string, declared number", true},
		{"predicate_overload_binding", "interface Circle { readonly kind: \"circle\"; readonly radius: number }\ninterface Square { readonly kind: \"square\"; readonly side: number }\ntype Shape = Circle | Square;\nlet hook: () => void = () => {};\nconst other: { node: Shape } = { node: { kind: \"square\", side: 2 } };\nfunction findIt(node: Shape, callback: (element: Shape) => element is Circle): Circle | undefined;\nfunction findIt(node: Shape, callback: (element: Shape) => boolean): Shape | undefined;\nfunction findIt(node: Shape, callback: (element: Shape) => boolean): Shape | undefined {\n    hook = () => { ({ node } = other); };\n    const result = callback(node);\n    if (result) {\n        return node;\n    }\n    return undefined;\n}\nconst found = findIt({ kind: \"circle\", radius: 3 }, (x): x is Circle => { hook(); return x.kind === \"circle\"; });\nconsole.log(found === undefined ? \"none\" : `radius ${found.radius}`);\n", "main.ts:16:15: expected Circle | undefined", true},
		{"predicate_overload_callback", "interface Circle { readonly kind: \"circle\"; readonly radius: number }\ninterface Square { readonly kind: \"square\"; readonly side: number }\ntype Shape = Circle | Square;\nconst liar: { predicate: (element: Shape) => boolean } = { predicate: (element: Shape) => element.kind === \"square\" };\nfunction first(array: readonly Shape[], predicate: (element: Shape) => element is Circle): Circle | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined {\n    ({ predicate } = liar);\n    for (const value of array) {\n        if (predicate(value)) {\n            return value;\n        }\n    }\n    return undefined;\n}\nconst shapes: Shape[] = [{ kind: \"square\", side: 2 }, { kind: \"circle\", radius: 3 }];\nconst circle = first(shapes, (x): x is Circle => x.kind === \"circle\");\nconsole.log(circle === undefined ? \"none\" : `radius ${circle.radius}`);\n", "main.ts:5:72: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter)", false},
		{"predicate_overload_nested_pattern", "interface Circle { readonly kind: \"circle\"; readonly radius: number }\ninterface Square { readonly kind: \"square\"; readonly side: number }\ntype Shape = Circle | Square;\nconst nest: { inner: { predicate: (element: Shape) => boolean } } = { inner: { predicate: (element: Shape) => element.kind === \"square\" } };\nfunction first(array: readonly Shape[], predicate: (element: Shape) => element is Circle): Circle | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined {\n    ({ inner: { predicate } } = nest);\n    for (const value of array) {\n        if (predicate(value)) {\n            return value;\n        }\n    }\n    return undefined;\n}\nconst shapes: Shape[] = [{ kind: \"square\", side: 2 }, { kind: \"circle\", radius: 3 }];\nconst circle = first(shapes, (x): x is Circle => x.kind === \"circle\");\nconsole.log(circle === undefined ? \"none\" : `radius ${circle.radius}`);\n", "main.ts:5:72: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter)", false},
		{"predicate_overload_for_of_pattern", "interface Circle { readonly kind: \"circle\"; readonly radius: number }\ninterface Square { readonly kind: \"square\"; readonly side: number }\ntype Shape = Circle | Square;\nconst liars: { predicate: (element: Shape) => boolean }[] = [{ predicate: (element: Shape) => element.kind === \"square\" }];\nfunction first(array: readonly Shape[], predicate: (element: Shape) => element is Circle): Circle | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined;\nfunction first(array: readonly Shape[], predicate: (element: Shape) => boolean): Shape | undefined {\n    for ({ predicate } of liars) {}\n    for (const value of array) {\n        if (predicate(value)) {\n            return value;\n        }\n    }\n    return undefined;\n}\nconst shapes: Shape[] = [{ kind: \"square\", side: 2 }, { kind: \"circle\", radius: 3 }];\nconst circle = first(shapes, (x): x is Circle => x.kind === \"circle\");\nconsole.log(circle === undefined ? \"none\" : `radius ${circle.radius}`);\n", "main.ts:5:72: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter)", false},
		{"record_read_discarded", "const dict: Record<string, number> = { a: 1 };\nconst v = dict[\"toString\"];\nconst held = { v };\nconsole.log(`${typeof held.v}`);\n", "main.ts:2:11: Adamic 0.1 refuses record key toString names an Object.prototype member", false},
		{"static_class_escape", "class Box {\n    static first: number = Box.peek();\n    static second: number = 7;\n    static peek(): number {\n        const held = { Box };\n        return held.Box.second;\n    }\n}\nconsole.log(`${Box.first}`);\n", "main.ts:5:24: Adamic 0.1 refuses a constructor object escaping before static fields are initialized", false},
		{"static_outside_helper", "function peek(): number {\n    const held = { Box };\n    return held.Box.second;\n}\nclass Box {\n    static first: number = peek();\n    static second: number = 7;\n}\nconsole.log(`${Box.first}`);\n", "main.ts:2:20: stage 0 can't lower a class used from a function that may run before the class's declaration (its temporal dead zone) yet", false},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, pin.source, ".ts")
			if !pin.native {
				if err == nil || !strings.Contains(err.Error(), pin.want) {
					t.Fatalf("want refused with %q, got %v", pin.want, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want it lowered and checked, got %v", err)
			}
			if code, stdout, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, pin.want) {
				t.Fatalf("want exit 70 with %q, got exit %d, stdout %q, stderr %q", pin.want, code, stdout, stderr)
			}
		})
	}
}
