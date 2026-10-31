package lower

import (
	"errors"
	"strings"
	"testing"
)

// An as or satisfies goes into its slot as the value it is (#b6m00y7, Language's a1): its relation to its
// operand passing says nothing of the slot, and before this inner satisfies string[] went into a
// (string | number)[], which wrote a number past the strings and read past the array's end natively. Each
// row is refused with its uncast twin's words, at the cast. a3 and a4 keep the stops other rules give the
// same shape. The checked cast's row is a .ts downcast whose own type goes into a wider writable slot, which
// stringWideningRefusal let through. Mutant: castSlotWidened finding nothing lowers a1, and the native
// binary reads past the end under ASan.
func TestCastWideningSlotIsRefused(t *testing.T) {
	t.Parallel()
	const invariant = "Adamic 0.1 refuses a value of type string[] seen as (string | number)[], which can write string | number where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	const inner = "const inner: string[] = [\"a\"];\n"
	const wide = "wide.push(5);\nconst last = inner[1] ?? \"\";\nconsole.log(last.toUpperCase());\n"
	const outer = "const outer: (string | number)[][] = [];\n"
	const pushed = "const first = outer[0];\nif (first !== undefined) first.push(5);\nconst last = inner[1] ?? \"\";\nconsole.log(last.toUpperCase());\n"
	const shapes = "type Circle = { readonly kind: 'Circle'; readonly radius: number };\ntype Square = { readonly kind: 'Square'; readonly side: number };\ntype Shape = Circle | Square;\nconst shapes: Shape[] = [{ kind: 'Circle', radius: 2 }];\nconst first = shapes[0];\n"
	for _, row := range []struct{ name, extension, source, want string }{
		{"a1, satisfies in .a", ".a", inner + "const wide: (string | number)[] = inner satisfies string[];\n" + wide, "main.a:2:35: " + invariant},
		{"a1 uncast", ".a", inner + "const wide: (string | number)[] = inner;\n" + wide, "main.a:2:35: " + invariant},
		{"a2, as in .ts", ".ts", inner + "const wide: (string | number)[] = inner as string[];\n" + wide, "main.ts:2:35: " + invariant},
		{"push with as", ".ts", inner + outer + "outer.push(inner as string[]);\n" + pushed, "main.ts:3:12: " + invariant},
		{"push with satisfies", ".ts", inner + outer + "outer.push(inner satisfies string[]);\n" + pushed, "main.ts:3:12: " + invariant},
		{"push uncast", ".ts", inner + outer + "outer.push(inner);\n" + pushed, "main.ts:3:12: " + invariant},
		{"checked cast", ".ts", shapes + "if (first !== undefined) {\n\tconst loose: { kind: string; radius: number } = first as Circle;\n\tloose.kind = 'Square';\n}\n", "main.ts:7:50: Adamic 0.1 refuses a value of type Circle seen as { kind: string; radius: number; }, whose readonly field kind becomes writable: a readonly field may hold something narrower than \"Circle\", which a write of string would replace; keep kind readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"},
		{"a3", ".a", "class Point { x = 1; describe(): string { return \"point\"; } }\nclass Other { x = 2; label = \"other\"; describe(): string { return this.label; } }\nconst other = new Other();\nconst point: Point = other satisfies Other;\nconsole.log(point.describe().toUpperCase());\n", "main.a:4:22: Adamic 0.1 refuses a value without nominal ancestry seen as Point; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"},
		{"a4", ".a", "const make = (): void => {};\nconst wide: () => unknown = make satisfies () => void;\nconsole.log(String(wide() === undefined));\n", "main.a:2:29: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, row.extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "/"+row.want) {
				t.Fatalf("want refusal ending %q, got %v", row.want, err)
			}
		})
	}
}

// The judgment refuses only a slot that can write what the value can't hold: an honest as const, a
// readonly target and a literal made right there under the cast each lower and agree with Node.
func TestCastIntoItsOwnSlotLowers(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"as const":         "const point = { x: 1 } as const;\nconst view: { readonly x: number } = point;\nconsole.log(String(view.x));\n",
		"readonly target":  "const inner: string[] = ['a'];\nconst wide: readonly (string | number)[] = inner satisfies string[];\nconst view: readonly (string | number)[] = inner as string[];\nconsole.log(`${wide.length} ${view.length}`);\n",
		"fresh under cast": "const wide: (string | number)[] = ['a'] satisfies string[];\nwide.push(5);\nconsole.log(String(wide.length));\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, ".a")
		})
	}
}
